package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	BookSourceGoogleBooks = "google_books"
	BookSourceManual      = "manual"
)

type Book struct {
	ID             uuid.UUID `json:"id"`
	GoogleBooksID  *string   `json:"google_books_id,omitempty"`
	Title          string    `json:"title"`
	Authors        []string  `json:"authors"`
	Genres         []string  `json:"genres"`
	ISBN           *string   `json:"isbn,omitempty"`
	Synopsis       *string   `json:"synopsis,omitempty"`
	CoverURL       *string   `json:"cover_url,omitempty"`
	TotalPages     *int      `json:"total_pages,omitempty"`
	Source         string    `json:"source"`
	CreatedAt      time.Time `json:"created_at"`
}
