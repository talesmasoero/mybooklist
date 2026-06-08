package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAnswers    = errors.New("one or more security answers are invalid")
	ErrInsufficientAnswers = errors.New("at least 2 correct answers are required")
)

type SecurityQuestion struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type UserSecurityAnswer struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	QuestionID int        `json:"question_id"`
	AnswerHash string     `json:"answer_hash"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type SaveAnswerInput struct {
	QuestionID int    `json:"question_id"`
	Answer     string `json:"answer"`
}

type AnswerInput struct {
	QuestionID int    `json:"question_id"`
	Answer     string `json:"answer"`
}

type PasswordResetSession struct {
	UserID    uuid.UUID         `json:"user_id"`
	Questions []SecurityQuestion `json:"questions"`
}
