package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID              uuid.UUID `json:"id"`
	ReadingID       uuid.UUID `json:"reading_id"`
	StartPage       int       `json:"start_page"`
	EndPage         int       `json:"end_page"`
	DurationSeconds *int      `json:"duration_seconds,omitempty"`
	SessionDate     time.Time `json:"session_date"`
	CreatedAt       time.Time `json:"created_at"`
}

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrInvalidPageRange = errors.New("invalid page range")
)
