package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

type UserSecurityAnswerRepository interface {
	SaveAnswersInTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, answers []domain.SaveAnswerInput) error
	DeleteByUserIDInTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserSecurityAnswer, error)
	GetQuestionsForUser(ctx context.Context, userID uuid.UUID) ([]domain.SecurityQuestion, error)
	GetQuestionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]int, error)
}

type postgresUserSecurityAnswerRepository struct {
	db *sql.DB
}

func NewPostgresUserSecurityAnswerRepository(db *sql.DB) UserSecurityAnswerRepository {
	return &postgresUserSecurityAnswerRepository{db: db}
}

func (r *postgresUserSecurityAnswerRepository) SaveAnswersInTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, answers []domain.SaveAnswerInput) error {
	const q = `
		INSERT INTO user_security_answers (user_id, question_id, answer_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, question_id) DO UPDATE
			SET answer_hash = EXCLUDED.answer_hash,
			    updated_at  = now()`

	for _, a := range answers {
		if _, err := tx.ExecContext(ctx, q, userID, a.QuestionID, a.Answer); err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresUserSecurityAnswerRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserSecurityAnswer, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, question_id, answer_hash, created_at, updated_at
		 FROM user_security_answers
		 WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []domain.UserSecurityAnswer
	for rows.Next() {
		var a domain.UserSecurityAnswer
		if err := rows.Scan(&a.ID, &a.UserID, &a.QuestionID, &a.AnswerHash, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

func (r *postgresUserSecurityAnswerRepository) DeleteByUserIDInTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM user_security_answers WHERE user_id = $1`, userID)
	return err
}

func (r *postgresUserSecurityAnswerRepository) GetQuestionsForUser(ctx context.Context, userID uuid.UUID) ([]domain.SecurityQuestion, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT sq.id, sq.text
		 FROM user_security_answers usa
		 JOIN security_questions sq ON usa.question_id = sq.id
		 WHERE usa.user_id = $1
		 ORDER BY sq.id`, userID)
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

func (r *postgresUserSecurityAnswerRepository) GetQuestionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT question_id FROM user_security_answers WHERE user_id = $1 ORDER BY question_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
