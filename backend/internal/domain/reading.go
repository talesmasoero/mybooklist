package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusWantToRead = "want_to_read"
	StatusReading    = "reading"
	StatusRead       = "read"
	StatusAbandoned  = "abandoned"
)

type Reading struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	BookID      uuid.UUID  `json:"book_id"`
	Status      string     `json:"status"`
	CurrentPage int        `json:"current_page"`
	AddedAt     time.Time  `json:"added_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Book *Book `json:"book,omitempty"`
}

func IsValidStatus(s string) bool {
	switch s {
	case StatusWantToRead, StatusReading, StatusRead, StatusAbandoned:
		return true
	}
	return false
}
