package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appmiddleware "github.com/talesmasoero/mybooklist/backend/internal/middleware"
	"github.com/talesmasoero/mybooklist/backend/internal/services"
)

type SessionHandler struct {
	sessionSvc services.SessionService
}

func NewSessionHandler(sessionSvc services.SessionService) *SessionHandler {
	return &SessionHandler{sessionSvc: sessionSvc}
}

type sessionRequest struct {
	StartPage       int    `json:"start_page"`
	EndPage         int    `json:"end_page"`
	DurationSeconds *int   `json:"duration_seconds"`
	SessionDate     string `json:"session_date"`
}

func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	readingID, err := uuid.Parse(chi.URLParam(r, "readingId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid reading id")
		return
	}

	var body sessionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}
	if body.StartPage < 1 {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "start_page must be at least 1")
		return
	}
	if body.EndPage < 1 {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "end_page is required")
		return
	}

	sessionDate := time.Now().UTC()
	if body.SessionDate != "" {
		parsed, err := time.Parse("2006-01-02", body.SessionDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "session_date must be in YYYY-MM-DD format")
			return
		}
		sessionDate = parsed.UTC()
	}

	payload := services.CreateSessionPayload{
		StartPage:       body.StartPage,
		EndPage:         body.EndPage,
		DurationSeconds: body.DurationSeconds,
		SessionDate:     sessionDate,
	}

	session, err := h.sessionSvc.CreateSession(r.Context(), userID, readingID, payload)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	readingID, err := uuid.Parse(chi.URLParam(r, "readingId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid reading id")
		return
	}

	sessions, err := h.sessionSvc.ListSessions(r.Context(), userID, readingID)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid session id")
		return
	}

	var body sessionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}
	if body.StartPage < 1 {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "start_page must be at least 1")
		return
	}
	if body.EndPage < 1 {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "end_page is required")
		return
	}

	sessionDate := time.Now().UTC()
	if body.SessionDate != "" {
		parsed, err := time.Parse("2006-01-02", body.SessionDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "session_date must be in YYYY-MM-DD format")
			return
		}
		sessionDate = parsed.UTC()
	}

	payload := services.UpdateSessionPayload{
		StartPage:       body.StartPage,
		EndPage:         body.EndPage,
		DurationSeconds: body.DurationSeconds,
		SessionDate:     sessionDate,
	}

	session, err := h.sessionSvc.UpdateSession(r.Context(), userID, sessionID, payload)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid session id")
		return
	}

	if err := h.sessionSvc.DeleteSession(r.Context(), userID, sessionID); err != nil {
		handleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
