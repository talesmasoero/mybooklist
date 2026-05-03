package services

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	"github.com/talesmasoero/mybooklist/backend/internal/repositories"
)

type BookDataPayload struct {
	Title      string   `json:"title"`
	Authors    []string `json:"authors"`
	Genres     []string `json:"genres"`
	ISBN       string   `json:"isbn"`
	Synopsis   string   `json:"synopsis"`
	CoverURL   string   `json:"cover_url"`
	TotalPages int      `json:"total_pages"`
}

type AddToLibraryPayload struct {
	Source        string          `json:"source"`
	GoogleBooksID string          `json:"google_books_id"`
	BookData      BookDataPayload `json:"book_data"`
	Status        string          `json:"status"`
}

type BookService interface {
	SearchExternal(ctx context.Context, query string, maxResults int) ([]domain.BookSearchResult, error)
	AddToLibrary(ctx context.Context, userID uuid.UUID, payload AddToLibraryPayload) (*domain.Reading, error)
	ListLibrary(ctx context.Context, userID uuid.UUID, status *string) ([]domain.Reading, error)
	UpdateReadingStatus(ctx context.Context, userID, readingID uuid.UUID, status string) (*domain.Reading, error)
}

type bookService struct {
	books    repositories.BookRepository
	readings repositories.ReadingRepository
	searcher domain.BookSearcher
}

func NewBookService(books repositories.BookRepository, readings repositories.ReadingRepository, searcher domain.BookSearcher) BookService {
	return &bookService{
		books:    books,
		readings: readings,
		searcher: searcher,
	}
}

func (s *bookService) SearchExternal(ctx context.Context, query string, maxResults int) ([]domain.BookSearchResult, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "query must have at least 2 characters"}
	}
	results, err := s.searcher.Search(ctx, query, maxResults)
	if err != nil {
		if errors.Is(err, domain.ErrExternalAPIFailed) {
			return nil, &domain.AppError{Code: http.StatusBadGateway, ErrorCode: "ERR_EXTERNAL_API_FAILED", Message: "book search service unavailable", Err: err}
		}
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to search books", Err: err}
	}
	return results, nil
}

func (s *bookService) AddToLibrary(ctx context.Context, userID uuid.UUID, payload AddToLibraryPayload) (*domain.Reading, error) {
	if err := validateAddToLibrary(payload); err != nil {
		return nil, err
	}

	book, err := s.resolveBook(ctx, payload)
	if err != nil {
		return nil, err
	}

	if _, err := s.readings.GetByUserAndBook(ctx, userID, book.ID); err == nil {
		return nil, &domain.AppError{Code: http.StatusConflict, ErrorCode: "ERR_ALREADY_IN_LIBRARY", Message: "book already in your library", Err: domain.ErrAlreadyInLibrary}
	} else if !errors.Is(err, domain.ErrReadingNotFound) {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to check existing reading", Err: err}
	}

	now := time.Now().UTC()
	reading := &domain.Reading{
		ID:          uuid.New(),
		UserID:      userID,
		BookID:      book.ID,
		Status:      payload.Status,
		CurrentPage: 1,
		AddedAt:     now,
		UpdatedAt:   now,
	}
	if err := s.readings.Create(ctx, reading); err != nil {
		if errors.Is(err, domain.ErrAlreadyInLibrary) {
			return nil, &domain.AppError{Code: http.StatusConflict, ErrorCode: "ERR_ALREADY_IN_LIBRARY", Message: "book already in your library", Err: err}
		}
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to create reading", Err: err}
	}
	reading.Book = book
	return reading, nil
}

func (s *bookService) ListLibrary(ctx context.Context, userID uuid.UUID, status *string) ([]domain.Reading, error) {
	if status != nil && !domain.IsValidStatus(*status) {
		return nil, &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "invalid status filter"}
	}
	readings, err := s.readings.ListByUser(ctx, userID, status)
	if err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to list library", Err: err}
	}
	return readings, nil
}

func (s *bookService) UpdateReadingStatus(ctx context.Context, userID, readingID uuid.UUID, status string) (*domain.Reading, error) {
	if !domain.IsValidStatus(status) {
		return nil, &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "invalid status"}
	}
	reading, err := s.readings.GetByIDWithBook(ctx, readingID)
	if err != nil {
		if errors.Is(err, domain.ErrReadingNotFound) {
			return nil, &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_NOT_FOUND", Message: "reading not found"}
		}
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to get reading", Err: err}
	}
	if reading.UserID != userID {
		return nil, &domain.AppError{Code: http.StatusForbidden, ErrorCode: "ERR_FORBIDDEN", Message: "reading does not belong to user"}
	}
	if err := s.readings.UpdateStatus(ctx, readingID, status); err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to update status", Err: err}
	}
	updated, err := s.readings.GetByIDWithBook(ctx, readingID)
	if err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to fetch updated reading", Err: err}
	}
	return updated, nil
}

func (s *bookService) resolveBook(ctx context.Context, payload AddToLibraryPayload) (*domain.Book, error) {
	if payload.Source == domain.BookSourceGoogleBooks {
		existing, err := s.books.GetByGoogleBooksID(ctx, payload.GoogleBooksID)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, domain.ErrBookNotFound) {
			return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to lookup book", Err: err}
		}
	}

	book := buildBook(payload)
	if err := s.books.Create(ctx, book); err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to create book", Err: err}
	}
	return book, nil
}

func buildBook(payload AddToLibraryPayload) *domain.Book {
	now := time.Now().UTC()
	book := &domain.Book{
		ID:        uuid.New(),
		Title:     strings.TrimSpace(payload.BookData.Title),
		Authors:   normalizeStrings(payload.BookData.Authors),
		Genres:    normalizeStrings(payload.BookData.Genres),
		Source:    payload.Source,
		CreatedAt: now,
	}
	if payload.Source == domain.BookSourceGoogleBooks && payload.GoogleBooksID != "" {
		gid := payload.GoogleBooksID
		book.GoogleBooksID = &gid
	}
	if v := strings.TrimSpace(payload.BookData.ISBN); v != "" {
		book.ISBN = &v
	}
	if v := strings.TrimSpace(payload.BookData.Synopsis); v != "" {
		book.Synopsis = &v
	}
	if v := strings.TrimSpace(payload.BookData.CoverURL); v != "" {
		book.CoverURL = &v
	}
	if payload.BookData.TotalPages > 0 {
		v := payload.BookData.TotalPages
		book.TotalPages = &v
	}
	return book
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func validateAddToLibrary(p AddToLibraryPayload) error {
	switch p.Source {
	case domain.BookSourceGoogleBooks:
		if strings.TrimSpace(p.GoogleBooksID) == "" {
			return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "google_books_id is required when source=google_books"}
		}
	case domain.BookSourceManual:
		// manual: nothing extra beyond book_data validation below
	default:
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "invalid source"}
	}

	if strings.TrimSpace(p.BookData.Title) == "" {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "book_data.title is required"}
	}
	if len(normalizeStrings(p.BookData.Authors)) == 0 {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "book_data.authors must contain at least one author"}
	}
	if p.Status != domain.StatusWantToRead && p.Status != domain.StatusReading {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "status must be want_to_read or reading"}
	}
	return nil
}
