package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

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

type securityAnswerRequest struct {
	QuestionID int    `json:"question_id"`
	Answer     string `json:"answer"`
}

type registerRequest struct {
	Email           string                  `json:"email"`
	Password        string                  `json:"password"`
	Name            string                  `json:"name"`
	SecurityAnswers []securityAnswerRequest `json:"security_answers"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type forgotPasswordResponse struct {
	UserID    string                   `json:"user_id"`
	Questions []securityQuestionJSON   `json:"questions"`
}

type securityQuestionJSON struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type verifyAnswersRequest struct {
	UserID  string                  `json:"user_id"`
	Answers []securityAnswerRequest `json:"answers"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	answers := make([]domain.SaveAnswerInput, len(req.SecurityAnswers))
	for i, a := range req.SecurityAnswers {
		answers[i] = domain.SaveAnswerInput{QuestionID: a.QuestionID, Answer: a.Answer}
	}

	input := services.RegisterInput{
		Email:           req.Email,
		Password:        req.Password,
		Name:            req.Name,
		SecurityAnswers: answers,
	}

	user, accessToken, refreshToken, err := h.authSvc.Register(r.Context(), input)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		User:         toUserResponse(user),
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
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	session, err := h.authSvc.InitiatePasswordReset(r.Context(), req.Email)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	questions := make([]securityQuestionJSON, len(session.Questions))
	for i, q := range session.Questions {
		questions[i] = securityQuestionJSON{ID: q.ID, Text: q.Text}
	}

	userID := ""
	if len(questions) >= 2 {
		userID = session.UserID.String()
	}

	writeJSON(w, http.StatusOK, forgotPasswordResponse{
		UserID:    userID,
		Questions: questions,
	})
}

func (h *AuthHandler) VerifySecurityAnswers(w http.ResponseWriter, r *http.Request) {
	var req verifyAnswersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid user_id")
		return
	}

	answers := make([]domain.AnswerInput, len(req.Answers))
	for i, a := range req.Answers {
		answers[i] = domain.AnswerInput{QuestionID: a.QuestionID, Answer: a.Answer}
	}

	token, err := h.authSvc.ValidatePasswordResetAnswers(r.Context(), userID, answers)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"reset_token": token})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	if err := h.authSvc.ResetPasswordWithToken(r.Context(), req.Token, req.NewPassword); err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toUserResponse(user *domain.User) userResponse {
	return userResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}
}

func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		slog.ErrorContext(r.Context(), "unexpected error", "error", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "internal server error")
		return
	}
	if appErr.Code >= 500 {
		slog.ErrorContext(r.Context(), "service error", "error", appErr)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "internal server error")
		return
	}
	writeError(w, appErr.Code, appErr.ErrorCode, appErr.Message)
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
