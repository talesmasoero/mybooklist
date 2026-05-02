package domain

import "context"

type BookSearchResult struct {
	GoogleBooksID string   `json:"google_books_id"`
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Genres        []string `json:"genres"`
	ISBN          string   `json:"isbn,omitempty"`
	Synopsis      string   `json:"synopsis,omitempty"`
	CoverURL      string   `json:"cover_url,omitempty"`
	TotalPages    int      `json:"total_pages,omitempty"`
}

type BookSearcher interface {
	Search(ctx context.Context, query string, maxResults int) ([]BookSearchResult, error)
}
