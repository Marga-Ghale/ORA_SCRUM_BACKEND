package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Invitation struct {
	ID        string
	Email     string
	Token     string
	Type      string // workspace, team, project
	TargetID  string
	Role      string
	InvitedBy string
	Status    string // pending, accepted, expired
	ExpiresAt time.Time
	CreatedAt time.Time
}

type InvitationRepository interface {
	Create(ctx context.Context, invitation *Invitation) error
	FindByID(ctx context.Context, id string) (*Invitation, error)
	FindByToken(ctx context.Context, token string) (*Invitation, error)
	FindByEmail(ctx context.Context, email string) ([]*Invitation, error)
	FindPendingByTarget(ctx context.Context, targetType, targetID string) ([]*Invitation, error)
	Update(ctx context.Context, invitation *Invitation) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int, error)
}

type pgInvitationRepository struct {
	pool *pgxpool.Pool
}

func NewInvitationRepository(pool *pgxpool.Pool) InvitationRepository {
	return &pgInvitationRepository{pool: pool}
}

func (r *pgInvitationRepository) Create(ctx context.Context, invitation *Invitation) error {
	invitation.Token = uuid.New().String()
	query := `
		INSERT INTO invitations (email, token, type, target_id, role, invited_by, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		invitation.Email, invitation.Token, invitation.Type, invitation.TargetID,
		invitation.Role, invitation.InvitedBy, invitation.Status, invitation.ExpiresAt,
	).Scan(&invitation.ID, &invitation.CreatedAt)
}

func (r *pgInvitationRepository) FindByID(ctx context.Context, id string) (*Invitation, error) {
	query := `
		SELECT id, email, token, type, target_id, role, invited_by, status, expires_at, created_at
		FROM invitations WHERE id = $1
	`
	invitation := &Invitation{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&invitation.ID, &invitation.Email, &invitation.Token, &invitation.Type,
		&invitation.TargetID, &invitation.Role, &invitation.InvitedBy, &invitation.Status,
		&invitation.ExpiresAt, &invitation.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *pgInvitationRepository) FindByToken(ctx context.Context, token string) (*Invitation, error) {
	query := `
		SELECT id, email, token, type, target_id, role, invited_by, status, expires_at, created_at
		FROM invitations WHERE token = $1
	`
	invitation := &Invitation{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&invitation.ID, &invitation.Email, &invitation.Token, &invitation.Type,
		&invitation.TargetID, &invitation.Role, &invitation.InvitedBy, &invitation.Status,
		&invitation.ExpiresAt, &invitation.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *pgInvitationRepository) FindByEmail(ctx context.Context, email string) ([]*Invitation, error) {
	query := `
		SELECT id, email, token, type, target_id, role, invited_by, status, expires_at, created_at
		FROM invitations WHERE email = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*Invitation
	for rows.Next() {
		invitation := &Invitation{}
		if err := rows.Scan(
			&invitation.ID, &invitation.Email, &invitation.Token, &invitation.Type,
			&invitation.TargetID, &invitation.Role, &invitation.InvitedBy, &invitation.Status,
			&invitation.ExpiresAt, &invitation.CreatedAt,
		); err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, nil
}

func (r *pgInvitationRepository) FindPendingByTarget(ctx context.Context, targetType, targetID string) ([]*Invitation, error) {
	query := `
		SELECT id, email, token, type, target_id, role, invited_by, status, expires_at, created_at
		FROM invitations WHERE type = $1 AND target_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, targetType, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*Invitation
	for rows.Next() {
		invitation := &Invitation{}
		if err := rows.Scan(
			&invitation.ID, &invitation.Email, &invitation.Token, &invitation.Type,
			&invitation.TargetID, &invitation.Role, &invitation.InvitedBy, &invitation.Status,
			&invitation.ExpiresAt, &invitation.CreatedAt,
		); err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, nil
}

func (r *pgInvitationRepository) Update(ctx context.Context, invitation *Invitation) error {
	query := `UPDATE invitations SET status = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, invitation.ID, invitation.Status)
	return err
}

func (r *pgInvitationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM invitations WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgInvitationRepository) DeleteExpired(ctx context.Context) (int, error) {
	query := `DELETE FROM invitations WHERE expires_at < NOW() AND status = 'pending'`
	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}
