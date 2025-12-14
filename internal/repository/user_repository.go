package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string
	Email        string
	Password     string
	Name         string
	Avatar       *string
	Status       string
	LastActiveAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByName(ctx context.Context, name string) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	UpdateLastActive(ctx context.Context, userID string) error
	UpdateStatusForInactive(ctx context.Context, inactiveDuration time.Duration) error
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteUserRefreshTokens(ctx context.Context, userID string) error
}

type pgUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &pgUserRepository{pool: pool}
}

func (r *pgUserRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, password, name, avatar, status, last_active_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	user.LastActiveAt = &now
	if user.Status == "" {
		user.Status = "online"
	}
	return r.pool.QueryRow(ctx, query,
		user.Email, user.Password, user.Name, user.Avatar, user.Status, now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *pgUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, password, name, avatar, status, last_active_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	user := &User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
		&user.Status, &user.LastActiveAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *pgUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password, name, avatar, status, last_active_at, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER($1)
	`
	user := &User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
		&user.Status, &user.LastActiveAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *pgUserRepository) FindByName(ctx context.Context, name string) (*User, error) {
	query := `
		SELECT id, email, password, name, avatar, status, last_active_at, created_at, updated_at
		FROM users WHERE LOWER(name) LIKE LOWER($1)
		LIMIT 1
	`
	user := &User{}
	err := r.pool.QueryRow(ctx, query, "%"+name+"%").Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
		&user.Status, &user.LastActiveAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *pgUserRepository) FindAll(ctx context.Context) ([]*User, error) {
	query := `
		SELECT id, email, password, name, avatar, status, last_active_at, created_at, updated_at
		FROM users ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
			&user.Status, &user.LastActiveAt, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *pgUserRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users SET email = $2, name = $3, avatar = $4, status = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.Name, user.Avatar, user.Status)
	return err
}

func (r *pgUserRepository) UpdateLastActive(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_active_at = NOW(), status = 'online' WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *pgUserRepository) UpdateStatusForInactive(ctx context.Context, inactiveDuration time.Duration) error {
	query := `
		UPDATE users SET status = 'away'
		WHERE status = 'online' AND last_active_at < $1
	`
	threshold := time.Now().Add(-inactiveDuration)
	_, err := r.pool.Exec(ctx, query, threshold)
	return err
}

func (r *pgUserRepository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query, token.Token, token.UserID, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
}

func (r *pgUserRepository) FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
		SELECT id, token, user_id, expires_at, created_at
		FROM refresh_tokens WHERE token = $1
	`
	rt := &RefreshToken{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&rt.ID, &rt.Token, &rt.UserID, &rt.ExpiresAt, &rt.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *pgUserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.pool.Exec(ctx, query, token)
	return err
}

func (r *pgUserRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
