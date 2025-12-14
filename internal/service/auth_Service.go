package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/config"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ============================================
// Auth Service
// ============================================

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*repository.User, string, string, error)
	Login(ctx context.Context, email, password string) (*repository.User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(token string) (*jwt.Token, error)
	GetUserIDFromToken(token *jwt.Token) (string, error)
}

type authService struct {
	cfg      *config.Config
	userRepo repository.UserRepository
}

func NewAuthService(cfg *config.Config, userRepo repository.UserRepository) AuthService {
	return &authService{cfg: cfg, userRepo: userRepo}
}

func (s *authService) Register(ctx context.Context, name, email, password string) (*repository.User, string, string, error) {
	existingUser, _ := s.userRepo.FindByEmail(ctx, email)
	if existingUser != nil {
		return nil, "", "", ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &repository.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Status:   "online",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*repository.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	user.Status = "online"
	s.userRepo.Update(ctx, user)
	s.userRepo.UpdateLastActive(ctx, user.ID)

	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	rt, err := s.userRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil || rt == nil {
		return "", "", ErrInvalidToken
	}

	if time.Now().After(rt.ExpiresAt) {
		s.userRepo.DeleteRefreshToken(ctx, refreshToken)
		return "", "", ErrInvalidToken
	}

	s.userRepo.DeleteRefreshToken(ctx, refreshToken)
	s.userRepo.UpdateLastActive(ctx, rt.UserID)

	accessToken, newRefreshToken, err := s.generateTokens(ctx, rt.UserID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.userRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *authService) GetUserIDFromToken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}
	return userID, nil
}

func (s *authService) generateTokens(ctx context.Context, userID string) (string, string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * time.Duration(s.cfg.JWTExpiry)).Unix(),
		"iat": time.Now().Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	refreshTokenString := uuid.New().String()
	refreshTokenExpiry := time.Now().Add(time.Hour * 24 * time.Duration(s.cfg.RefreshExpiry))

	rt := &repository.RefreshToken{
		Token:     refreshTokenString,
		UserID:    userID,
		ExpiresAt: refreshTokenExpiry,
	}

	if err := s.userRepo.SaveRefreshToken(ctx, rt); err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
