package services

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	"github.com/talesmasoero/mybooklist/backend/internal/repositories"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateName(ctx context.Context, userID uuid.UUID, name string) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
	DeleteAccount(ctx context.Context, userID uuid.UUID, currentPassword string) error
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_NOT_FOUND", Message: "user not found", Err: err}
		}
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to retrieve user", Err: err}
	}
	return user, nil
}

func (s *userService) UpdateName(ctx context.Context, userID uuid.UUID, name string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "name must be at least 2 characters"}
	}
	if len(name) > 100 {
		return nil, &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "name must be at most 100 characters"}
	}

	if err := s.repo.UpdateName(ctx, userID, name); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_NOT_FOUND", Message: "user not found", Err: err}
		}
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to update name", Err: err}
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to retrieve updated user", Err: err}
	}
	return user, nil
}

func (s *userService) UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_NOT_FOUND", Message: "user not found", Err: err}
		}
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to retrieve user", Err: err}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return &domain.AppError{Code: http.StatusUnauthorized, ErrorCode: "ERR_INVALID_CREDENTIALS", Message: "current password is incorrect", Err: domain.ErrInvalidCredentials}
	}

	if len(newPassword) < 8 {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "new password must be at least 8 characters"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to process password", Err: err}
	}

	if err := s.repo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to update password", Err: err}
	}
	return nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID uuid.UUID, currentPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_NOT_FOUND", Message: "user not found", Err: err}
		}
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to retrieve user", Err: err}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return &domain.AppError{Code: http.StatusUnauthorized, ErrorCode: "ERR_INVALID_CREDENTIALS", Message: "current password is incorrect", Err: domain.ErrInvalidCredentials}
	}

	if err := s.repo.Delete(ctx, userID); err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to delete account", Err: err}
	}
	return nil
}
