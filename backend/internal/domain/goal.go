package domain

import (
	"time"

	"github.com/google/uuid"
)

type Goal struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Year        int       `json:"year"`
	TargetBooks int       `json:"target_books"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GoalProgress struct {
	Goal
	BooksRead  int     `json:"books_read"`
	Percentage float64 `json:"percentage"`
}
