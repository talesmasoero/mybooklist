package handlers

import (
	"encoding/json"
	"net/http"

	appmiddleware "github.com/talesmasoero/mybooklist/backend/internal/middleware"
	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	"github.com/talesmasoero/mybooklist/backend/internal/services"
)

type SecurityQuestionHandler struct {
	sqSvc services.SecurityQuestionService
}

func NewSecurityQuestionHandler(sqSvc services.SecurityQuestionService) *SecurityQuestionHandler {
	return &SecurityQuestionHandler{sqSvc: sqSvc}
}

func (h *SecurityQuestionHandler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	questions, err := h.sqSvc.ListAvailableQuestions(r.Context())
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, questions)
}

func (h *SecurityQuestionHandler) GetMyAnswerIDs(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "unauthorized")
		return
	}

	ids, err := h.sqSvc.GetUserQuestionIDs(r.Context(), userID)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	if ids == nil {
		ids = []int{}
	}
	writeJSON(w, http.StatusOK, ids)
}

type saveAnswersRequest struct {
	Answers []struct {
		QuestionID int    `json:"question_id"`
		Answer     string `json:"answer"`
	} `json:"answers"`
}

func (h *SecurityQuestionHandler) SaveAnswers(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "unauthorized")
		return
	}

	var req saveAnswersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	answers := make([]domain.SaveAnswerInput, len(req.Answers))
	for i, a := range req.Answers {
		answers[i] = domain.SaveAnswerInput{QuestionID: a.QuestionID, Answer: a.Answer}
	}

	if err := h.sqSvc.SetUserAnswers(r.Context(), userID, answers); err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
