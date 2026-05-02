package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

type ReadingRepository interface {
	Create(ctx context.Context, reading *domain.Reading) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Reading, error)
	GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*domain.Reading, error)
	ListByUser(ctx context.Context, userID uuid.UUID, status *string) ([]domain.Reading, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type postgresReadingRepository struct {
	db *sql.DB
}

func NewPostgresReadingRepository(db *sql.DB) ReadingRepository {
	return &postgresReadingRepository{db: db}
}

func (r *postgresReadingRepository) Create(ctx context.Context, reading *domain.Reading) error {
	query := `
		INSERT INTO readings (id, user_id, book_id, status, current_page, added_at, completed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		reading.ID,
		reading.UserID,
		reading.BookID,
		reading.Status,
		reading.CurrentPage,
		reading.AddedAt,
		reading.CompletedAt,
		reading.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyInLibrary
		}
		return err
	}
	return nil
}

func (r *postgresReadingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Reading, error) {
	query := `
		SELECT id, user_id, book_id, status, current_page, added_at, completed_at, updated_at
		FROM readings
		WHERE id = $1
	`
	reading := &domain.Reading{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reading.ID,
		&reading.UserID,
		&reading.BookID,
		&reading.Status,
		&reading.CurrentPage,
		&reading.AddedAt,
		&reading.CompletedAt,
		&reading.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReadingNotFound
		}
		return nil, err
	}
	return reading, nil
}

func (r *postgresReadingRepository) GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*domain.Reading, error) {
	query := `
		SELECT id, user_id, book_id, status, current_page, added_at, completed_at, updated_at
		FROM readings
		WHERE user_id = $1 AND book_id = $2
	`
	reading := &domain.Reading{}
	err := r.db.QueryRowContext(ctx, query, userID, bookID).Scan(
		&reading.ID,
		&reading.UserID,
		&reading.BookID,
		&reading.Status,
		&reading.CurrentPage,
		&reading.AddedAt,
		&reading.CompletedAt,
		&reading.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReadingNotFound
		}
		return nil, err
	}
	return reading, nil
}

func (r *postgresReadingRepository) ListByUser(ctx context.Context, userID uuid.UUID, status *string) ([]domain.Reading, error) {
	query := `
		SELECT
			r.id, r.user_id, r.book_id, r.status, r.current_page, r.added_at, r.completed_at, r.updated_at,
			b.id, b.google_books_id, b.title, b.authors, b.genres, b.isbn, b.synopsis, b.cover_url, b.total_pages, b.source, b.created_at
		FROM readings r
		JOIN books b ON b.id = r.book_id
		WHERE r.user_id = $1
	`
	args := []any{userID}
	if status != nil {
		query += " AND r.status = $2"
		args = append(args, *status)
	}
	query += " ORDER BY r.added_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]domain.Reading, 0)
	for rows.Next() {
		reading := domain.Reading{}
		book := &domain.Book{}
		var authors, genres pq.StringArray
		err := rows.Scan(
			&reading.ID,
			&reading.UserID,
			&reading.BookID,
			&reading.Status,
			&reading.CurrentPage,
			&reading.AddedAt,
			&reading.CompletedAt,
			&reading.UpdatedAt,
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
			return nil, err
		}
		book.Authors = []string(authors)
		book.Genres = []string(genres)
		reading.Book = book
		results = append(results, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *postgresReadingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE readings
		SET status = $1, updated_at = now(),
		    completed_at = CASE WHEN $1 = 'read' AND completed_at IS NULL THEN now() ELSE completed_at END
		WHERE id = $2
	`
	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrReadingNotFound
	}
	return nil
}
