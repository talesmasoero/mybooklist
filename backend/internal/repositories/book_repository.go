package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

type BookRepository interface {
	Create(ctx context.Context, book *domain.Book) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	GetByGoogleBooksID(ctx context.Context, googleBooksID string) (*domain.Book, error)
}

type postgresBookRepository struct {
	db *sql.DB
}

func NewPostgresBookRepository(db *sql.DB) BookRepository {
	return &postgresBookRepository{db: db}
}

func (r *postgresBookRepository) Create(ctx context.Context, book *domain.Book) error {
	query := `
		INSERT INTO books (id, google_books_id, title, authors, genres, isbn, synopsis, cover_url, total_pages, source, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	authors := book.Authors
	if authors == nil {
		authors = []string{}
	}
	genres := book.Genres
	if genres == nil {
		genres = []string{}
	}
	_, err := r.db.ExecContext(ctx, query,
		book.ID,
		book.GoogleBooksID,
		book.Title,
		pq.Array(authors),
		pq.Array(genres),
		book.ISBN,
		book.Synopsis,
		book.CoverURL,
		book.TotalPages,
		book.Source,
		book.CreatedAt,
	)
	return err
}

func (r *postgresBookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	query := `
		SELECT id, google_books_id, title, authors, genres, isbn, synopsis, cover_url, total_pages, source, created_at
		FROM books
		WHERE id = $1
	`
	return scanBook(r.db.QueryRowContext(ctx, query, id))
}

func (r *postgresBookRepository) GetByGoogleBooksID(ctx context.Context, googleBooksID string) (*domain.Book, error) {
	query := `
		SELECT id, google_books_id, title, authors, genres, isbn, synopsis, cover_url, total_pages, source, created_at
		FROM books
		WHERE google_books_id = $1
	`
	return scanBook(r.db.QueryRowContext(ctx, query, googleBooksID))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBook(row rowScanner) (*domain.Book, error) {
	book := &domain.Book{}
	var authors, genres pq.StringArray
	err := row.Scan(
		&book.ID,
		&book.GoogleBooksID,
		&book.Title,
		&authors,
		&genres,
		&book.ISBN,
		&book.Synopsis,
		&book.CoverURL,
		&book.TotalPages,
		&book.Source,
		&book.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBookNotFound
		}
		return nil, err
	}
	book.Authors = []string(authors)
	book.Genres = []string(genres)
	return book, nil
}
