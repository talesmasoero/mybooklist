package repositories

import (
	"context"
	"database/sql"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

type SecurityQuestionRepository interface {
	ListAll(ctx context.Context) ([]domain.SecurityQuestion, error)
}

type postgresSecurityQuestionRepository struct {
	db *sql.DB
}

func NewPostgresSecurityQuestionRepository(db *sql.DB) SecurityQuestionRepository {
	return &postgresSecurityQuestionRepository{db: db}
}

func (r *postgresSecurityQuestionRepository) ListAll(ctx context.Context) ([]domain.SecurityQuestion, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, text FROM security_questions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []domain.SecurityQuestion
	for rows.Next() {
		var q domain.SecurityQuestion
		if err := rows.Scan(&q.ID, &q.Text); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}
