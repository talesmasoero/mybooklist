package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	ListByReading(ctx context.Context, readingID uuid.UUID) ([]domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresSessionRepository struct {
	db *sql.DB
}

func NewPostgresSessionRepository(db *sql.DB) SessionRepository {
	return &postgresSessionRepository{db: db}
}

func (r *postgresSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	insertQuery := `
		INSERT INTO sessions (id, reading_id, start_page, end_page, duration_seconds, session_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		session.ID,
		session.ReadingID,
		session.StartPage,
		session.EndPage,
		session.DurationSeconds,
		session.SessionDate,
		session.CreatedAt,
	)
	if err != nil {
		return err
	}

	updateQuery := `
		UPDATE readings
		SET current_page = $1, updated_at = now()
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, updateQuery, session.EndPage, session.ReadingID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT id, reading_id, start_page, end_page, duration_seconds, session_date, created_at
		FROM sessions
		WHERE id = $1
	`
	session := &domain.Session{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.ReadingID,
		&session.StartPage,
		&session.EndPage,
		&session.DurationSeconds,
		&session.SessionDate,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return session, nil
}

func (r *postgresSessionRepository) ListByReading(ctx context.Context, readingID uuid.UUID) ([]domain.Session, error) {
	query := `
		SELECT id, reading_id, start_page, end_page, duration_seconds, session_date, created_at
		FROM sessions
		WHERE reading_id = $1
		ORDER BY session_date DESC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, readingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]domain.Session, 0)
	for rows.Next() {
		s := domain.Session{}
		err := rows.Scan(
			&s.ID,
			&s.ReadingID,
			&s.StartPage,
			&s.EndPage,
			&s.DurationSeconds,
			&s.SessionDate,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *postgresSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	query := `
		UPDATE sessions
		SET start_page = $1, end_page = $2, duration_seconds = $3, session_date = $4
		WHERE id = $5
	`
	res, err := r.db.ExecContext(ctx, query,
		session.StartPage,
		session.EndPage,
		session.DurationSeconds,
		session.SessionDate,
		session.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *postgresSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}
