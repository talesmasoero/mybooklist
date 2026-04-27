package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	"github.com/talesmasoero/mybooklist/backend/internal/services"
)

type AuthHandler struct {
	authSvc services.AuthService
}

func NewAuthHandler(authSvc services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type authResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	user, accessToken, refreshToken, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		User: userResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	user, accessToken, refreshToken, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		User: userResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		slog.ErrorContext(r.Context(), "unexpected error", "error", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "internal server error")
		return
	}

	switch appErr.Code {
	case http.StatusBadRequest:
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", appErr.Message)
	case http.StatusUnauthorized:
		writeError(w, http.StatusUnauthorized, "ERR_INVALID_CREDENTIALS", appErr.Message)
	case http.StatusConflict:
		writeError(w, http.StatusConflict, "ERR_EMAIL_ALREADY_EXISTS", appErr.Message)
	default:
		slog.ErrorContext(r.Context(), "service error", "error", appErr)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
