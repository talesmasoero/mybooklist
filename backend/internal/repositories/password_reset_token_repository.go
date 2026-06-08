package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *domain.PasswordResetToken) error
	GetByHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error
}

type postgresPasswordResetTokenRepository struct {
	db *sql.DB
}

func NewPostgresPasswordResetTokenRepository(db *sql.DB) PasswordResetTokenRepository {
	return &postgresPasswordResetTokenRepository{db: db}
}

func (r *postgresPasswordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *postgresPasswordResetTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	var t domain.PasswordResetToken
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM password_reset_tokens
		 WHERE token_hash = $1`, tokenHash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenInvalid
		}
		return nil, err
	}
	return &t, nil
}

func (r *postgresPasswordResetTokenRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = $1 WHERE id = $2`, now, id)
	return err
}

func (r *postgresPasswordResetTokenRepository) InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = $1 WHERE user_id = $2 AND used_at IS NULL`, now, userID)
	return err
}
