package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

type GoalRepository interface {
	Create(ctx context.Context, goal *domain.Goal) error
	GetByUserAndYear(ctx context.Context, userID uuid.UUID, year int) (*domain.Goal, error)
	CountReadBooksByUserAndYear(ctx context.Context, userID uuid.UUID, year int) (int, error)
	UpdateTarget(ctx context.Context, goalID, userID uuid.UUID, targetBooks int) (*domain.Goal, error)
}

type postgresGoalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) GoalRepository {
	return &postgresGoalRepository{db: db}
}

func pgGoalErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrGoalAlreadyExists
	}
	return err
}

func (r *postgresGoalRepository) Create(ctx context.Context, goal *domain.Goal) error {
	const query = `
		INSERT INTO goals (id, user_id, year, target_books, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		goal.ID,
		goal.UserID,
		goal.Year,
		goal.TargetBooks,
		goal.CreatedAt,
		goal.UpdatedAt,
	)
	return pgGoalErr(err)
}

func (r *postgresGoalRepository) GetByUserAndYear(ctx context.Context, userID uuid.UUID, year int) (*domain.Goal, error) {
	const query = `
		SELECT id, user_id, year, target_books, created_at, updated_at
		FROM goals
		WHERE user_id = $1 AND year = $2
	`
	goal := &domain.Goal{}
	err := r.db.QueryRowContext(ctx, query, userID, year).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Year,
		&goal.TargetBooks,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrGoalNotFound
		}
		return nil, err
	}
	return goal, nil
}

func (r *postgresGoalRepository) UpdateTarget(ctx context.Context, goalID, userID uuid.UUID, targetBooks int) (*domain.Goal, error) {
	const query = `
		UPDATE goals
		SET target_books = $3, updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, year, target_books, created_at, updated_at
	`
	goal := &domain.Goal{}
	err := r.db.QueryRowContext(ctx, query, goalID, userID, targetBooks).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Year,
		&goal.TargetBooks,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrGoalNotFound
		}
		return nil, err
	}
	return goal, nil
}

func (r *postgresGoalRepository) CountReadBooksByUserAndYear(ctx context.Context, userID uuid.UUID, year int) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM readings
		WHERE user_id = $1
		  AND status = 'read'
		  AND EXTRACT(YEAR FROM completed_at) = $2
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, year).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
