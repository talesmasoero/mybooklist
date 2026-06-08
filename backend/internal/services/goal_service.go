package services

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	"github.com/talesmasoero/mybooklist/backend/internal/repositories"
)

type GoalService interface {
	CreateGoal(ctx context.Context, userID uuid.UUID, year, targetBooks int) (*domain.GoalProgress, error)
	GetCurrentYearGoal(ctx context.Context, userID uuid.UUID) (*domain.GoalProgress, error)
	UpdateGoal(ctx context.Context, userID, goalID uuid.UUID, targetBooks int) (*domain.GoalProgress, error)
}

type goalService struct {
	repo repositories.GoalRepository
}

func NewGoalService(repo repositories.GoalRepository) GoalService {
	return &goalService{repo: repo}
}

func (s *goalService) CreateGoal(ctx context.Context, userID uuid.UUID, year, targetBooks int) (*domain.GoalProgress, error) {
	currentYear := time.Now().Year()
	if year < currentYear {
		return nil, &domain.AppError{
			Code:      http.StatusBadRequest,
			ErrorCode: "ERR_VALIDATION",
			Message:   "year must be current year or future",
		}
	}
	if targetBooks < 1 || targetBooks > 1000 {
		return nil, &domain.AppError{
			Code:      http.StatusBadRequest,
			ErrorCode: "ERR_VALIDATION",
			Message:   "target_books must be between 1 and 1000",
		}
	}

	now := time.Now()
	goal := &domain.Goal{
		ID:          uuid.New(),
		UserID:      userID,
		Year:        year,
		TargetBooks: targetBooks,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, goal); err != nil {
		if errors.Is(err, domain.ErrGoalAlreadyExists) {
			return nil, &domain.AppError{
				Code:      http.StatusConflict,
				ErrorCode: "ERR_CONFLICT",
				Message:   "a goal already exists for this year",
				Err:       err,
			}
		}
		return nil, &domain.AppError{
			Code:      http.StatusInternalServerError,
			ErrorCode: "ERR_INTERNAL",
			Message:   "failed to create goal",
			Err:       err,
		}
	}

	return s.buildProgress(ctx, goal)
}

func (s *goalService) GetCurrentYearGoal(ctx context.Context, userID uuid.UUID) (*domain.GoalProgress, error) {
	year := time.Now().Year()
	goal, err := s.repo.GetByUserAndYear(ctx, userID, year)
	if err != nil {
		if errors.Is(err, domain.ErrGoalNotFound) {
			return nil, nil
		}
		return nil, &domain.AppError{
			Code:      http.StatusInternalServerError,
			ErrorCode: "ERR_INTERNAL",
			Message:   "failed to retrieve goal",
			Err:       err,
		}
	}
	return s.buildProgress(ctx, goal)
}

func (s *goalService) UpdateGoal(ctx context.Context, userID, goalID uuid.UUID, targetBooks int) (*domain.GoalProgress, error) {
	if targetBooks < 1 || targetBooks > 1000 {
		return nil, &domain.AppError{
			Code:      http.StatusBadRequest,
			ErrorCode: "ERR_VALIDATION",
			Message:   "target_books must be between 1 and 1000",
		}
	}

	goal, err := s.repo.UpdateTarget(ctx, goalID, userID, targetBooks)
	if err != nil {
		if errors.Is(err, domain.ErrGoalNotFound) {
			return nil, &domain.AppError{
				Code:      http.StatusNotFound,
				ErrorCode: "ERR_NOT_FOUND",
				Message:   "goal not found",
				Err:       err,
			}
		}
		return nil, &domain.AppError{
			Code:      http.StatusInternalServerError,
			ErrorCode: "ERR_INTERNAL",
			Message:   "failed to update goal",
			Err:       err,
		}
	}

	return s.buildProgress(ctx, goal)
}

func (s *goalService) buildProgress(ctx context.Context, goal *domain.Goal) (*domain.GoalProgress, error) {
	booksRead, err := s.repo.CountReadBooksByUserAndYear(ctx, goal.UserID, goal.Year)
	if err != nil {
		return nil, &domain.AppError{
			Code:      http.StatusInternalServerError,
			ErrorCode: "ERR_INTERNAL",
			Message:   "failed to count read books",
			Err:       err,
		}
	}

	percentage := 0.0
	if goal.TargetBooks > 0 {
		percentage = math.Min(100, float64(booksRead)/float64(goal.TargetBooks)*100)
	}

	return &domain.GoalProgress{
		Goal:       *goal,
		BooksRead:  booksRead,
		Percentage: math.Round(percentage*10) / 10,
	}, nil
}
