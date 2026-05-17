package services

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	"github.com/talesmasoero/mybooklist/backend/internal/repositories"
)

type CreateSessionPayload struct {
	StartPage       int
	EndPage         int
	DurationSeconds *int
	SessionDate     time.Time
}

type UpdateSessionPayload struct {
	StartPage       int
	EndPage         int
	DurationSeconds *int
	SessionDate     time.Time
}

type SessionService interface {
	CreateSession(ctx context.Context, userID, readingID uuid.UUID, payload CreateSessionPayload) (*domain.Session, error)
	ListSessions(ctx context.Context, userID, readingID uuid.UUID) ([]domain.Session, error)
	UpdateSession(ctx context.Context, userID, sessionID uuid.UUID, payload UpdateSessionPayload) (*domain.Session, error)
	DeleteSession(ctx context.Context, userID, sessionID uuid.UUID) error
}

type sessionService struct {
	sessions repositories.SessionRepository
	readings repositories.ReadingRepository
}

func NewSessionService(sessions repositories.SessionRepository, readings repositories.ReadingRepository) SessionService {
	return &sessionService{sessions: sessions, readings: readings}
}

func (s *sessionService) CreateSession(ctx context.Context, userID, readingID uuid.UUID, payload CreateSessionPayload) (*domain.Session, error) {
	reading, err := s.readings.GetByID(ctx, readingID)
	if err != nil {
		if errors.Is(err, domain.ErrReadingNotFound) {
			return nil, &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_NOT_FOUND", Message: "reading not found"}
		}
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to get reading", Err: err}
	}
	if reading.UserID != userID {
		return nil, &domain.AppError{Code: http.StatusForbidden, ErrorCode: "ERR_FORBIDDEN", Message: "reading does not belong to user"}
	}
	if payload.EndPage < payload.StartPage {
		return nil, &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_INVALID_PAGE_RANGE", Message: "end_page must be greater than or equal to start_page", Err: domain.ErrInvalidPageRange}
	}

	if reading.Status != domain.StatusReading {
		if err := s.readings.UpdateStatus(ctx, readingID, domain.StatusReading); err != nil {
			return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to update reading status", Err: err}
		}
	}

	now := time.Now().UTC()
	session := &domain.Session{
		ID:              uuid.New(),
		ReadingID:       readingID,
		StartPage:       payload.StartPage,
		EndPage:         payload.EndPage,
		DurationSeconds: payload.DurationSeconds,
		SessionDate:     payload.SessionDate,
		CreatedAt:       now,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to create session", Err: err}
	}
	return session, nil
}

func (s *sessionService) ListSessions(ctx context.Context, userID, readingID uuid.UUID) ([]domain.Session, error) {
	reading, err := s.readings.GetByID(ctx, readingID)
	if err != nil {
		if errors.Is(err, domain.ErrReadingNotFound) {
			return nil, &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_NOT_FOUND", Message: "reading not found"}
		}
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to get reading", Err: err}
	}
	if reading.UserID != userID {
		return nil, &domain.AppError{Code: http.StatusForbidden, ErrorCode: "ERR_FORBIDDEN", Message: "reading does not belong to user"}
	}

	sessions, err := s.sessions.ListByReading(ctx, readingID)
	if err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to list sessions", Err: err}
	}
	return sessions, nil
}

func (s *sessionService) UpdateSession(ctx context.Context, userID, sessionID uuid.UUID, payload UpdateSessionPayload) (*domain.Session, error) {
	session, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil, &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_SESSION_NOT_FOUND", Message: "session not found"}
		}
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to get session", Err: err}
	}

	reading, err := s.readings.GetByID(ctx, session.ReadingID)
	if err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to get reading", Err: err}
	}
	if reading.UserID != userID {
		return nil, &domain.AppError{Code: http.StatusForbidden, ErrorCode: "ERR_FORBIDDEN", Message: "reading does not belong to user"}
	}

	if payload.EndPage < payload.StartPage {
		return nil, &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_INVALID_PAGE_RANGE", Message: "end_page must be greater than or equal to start_page", Err: domain.ErrInvalidPageRange}
	}

	session.StartPage = payload.StartPage
	session.EndPage = payload.EndPage
	session.DurationSeconds = payload.DurationSeconds
	session.SessionDate = payload.SessionDate

	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to update session", Err: err}
	}
	return session, nil
}

func (s *sessionService) DeleteSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	session, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return &domain.AppError{Code: http.StatusNotFound, ErrorCode: "ERR_SESSION_NOT_FOUND", Message: "session not found"}
		}
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to get session", Err: err}
	}

	reading, err := s.readings.GetByID(ctx, session.ReadingID)
	if err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to get reading", Err: err}
	}
	if reading.UserID != userID {
		return &domain.AppError{Code: http.StatusForbidden, ErrorCode: "ERR_FORBIDDEN", Message: "reading does not belong to user"}
	}

	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to delete session", Err: err}
	}
	return nil
}
