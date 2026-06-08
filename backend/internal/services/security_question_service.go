package services

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	"github.com/talesmasoero/mybooklist/backend/internal/repositories"
)

type SecurityQuestionService interface {
	ListAvailableQuestions(ctx context.Context) ([]domain.SecurityQuestion, error)
	SetUserAnswers(ctx context.Context, userID uuid.UUID, answers []domain.SaveAnswerInput) error
	GetUserQuestionIDs(ctx context.Context, userID uuid.UUID) ([]int, error)
}

type securityQuestionService struct {
	db         *sql.DB
	sqRepo     repositories.SecurityQuestionRepository
	answerRepo repositories.UserSecurityAnswerRepository
}

func NewSecurityQuestionService(
	db *sql.DB,
	sqRepo repositories.SecurityQuestionRepository,
	answerRepo repositories.UserSecurityAnswerRepository,
) SecurityQuestionService {
	return &securityQuestionService{
		db:         db,
		sqRepo:     sqRepo,
		answerRepo: answerRepo,
	}
}

func (s *securityQuestionService) ListAvailableQuestions(ctx context.Context) ([]domain.SecurityQuestion, error) {
	questions, err := s.sqRepo.ListAll(ctx)
	if err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to list security questions", Err: err}
	}
	return questions, nil
}

func (s *securityQuestionService) GetUserQuestionIDs(ctx context.Context, userID uuid.UUID) ([]int, error) {
	ids, err := s.answerRepo.GetQuestionIDsByUserID(ctx, userID)
	if err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to get security answer question IDs", Err: err}
	}
	return ids, nil
}

func (s *securityQuestionService) SetUserAnswers(ctx context.Context, userID uuid.UUID, answers []domain.SaveAnswerInput) error {
	if err := validateAnswerCount(answers); err != nil {
		return err
	}
	if err := validateDistinctQuestionIDs(answers); err != nil {
		return err
	}
	if err := validateAnswerLengths(answers); err != nil {
		return err
	}

	allQuestions, err := s.sqRepo.ListAll(ctx)
	if err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to validate security questions", Err: err}
	}
	if err := validateQuestionIDsExist(answers, allQuestions); err != nil {
		return err
	}

	hashed := make([]domain.SaveAnswerInput, len(answers))
	for i, a := range answers {
		normalized := strings.ToLower(strings.TrimSpace(a.Answer))
		h, err := bcrypt.GenerateFromPassword([]byte(normalized), bcryptCost)
		if err != nil {
			return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to process security answers", Err: err}
		}
		hashed[i] = domain.SaveAnswerInput{QuestionID: a.QuestionID, Answer: string(h)}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to start transaction", Err: err}
	}
	defer tx.Rollback()

	if err := s.answerRepo.DeleteByUserIDInTx(ctx, tx, userID); err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to update security answers", Err: err}
	}
	if err := s.answerRepo.SaveAnswersInTx(ctx, tx, userID, hashed); err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to save security answers", Err: err}
	}
	if err := tx.Commit(); err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to commit transaction", Err: err}
	}
	return nil
}

func validateAnswerCount(answers []domain.SaveAnswerInput) error {
	if len(answers) < 2 || len(answers) > 3 {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "2 to 3 security answers required"}
	}
	return nil
}

func validateDistinctQuestionIDs(answers []domain.SaveAnswerInput) error {
	seen := make(map[int]bool)
	for _, a := range answers {
		if seen[a.QuestionID] {
			return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "security answer question IDs must be distinct"}
		}
		seen[a.QuestionID] = true
	}
	return nil
}

func validateAnswerLengths(answers []domain.SaveAnswerInput) error {
	for _, a := range answers {
		if len(strings.TrimSpace(a.Answer)) < 2 {
			return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "each security answer must be at least 2 characters"}
		}
	}
	return nil
}

func validateQuestionIDsExist(answers []domain.SaveAnswerInput, allQuestions []domain.SecurityQuestion) error {
	validIDs := make(map[int]bool, len(allQuestions))
	for _, q := range allQuestions {
		validIDs[q.ID] = true
	}
	for _, a := range answers {
		if !validIDs[a.QuestionID] {
			return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "invalid security question ID"}
		}
	}
	return nil
}
