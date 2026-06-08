package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTokenInvalid = errors.New("password reset token not found")
	ErrTokenExpired = errors.New("password reset token has expired")
	ErrTokenUsed    = errors.New("password reset token has already been used")
)

type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}
