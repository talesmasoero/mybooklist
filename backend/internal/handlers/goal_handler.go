package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appmiddleware "github.com/talesmasoero/mybooklist/backend/internal/middleware"
	"github.com/talesmasoero/mybooklist/backend/internal/services"
)

type GoalHandler struct {
	goalSvc services.GoalService
}

func NewGoalHandler(goalSvc services.GoalService) *GoalHandler {
	return &GoalHandler{goalSvc: goalSvc}
}

type createGoalRequest struct {
	Year        int `json:"year"`
	TargetBooks int `json:"target_books"`
}

type updateGoalRequest struct {
	TargetBooks int `json:"target_books"`
}

func (h *GoalHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	goalID, err := uuid.Parse(chi.URLParam(r, "goalId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid goal id")
		return
	}

	var req updateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	progress, err := h.goalSvc.UpdateGoal(r.Context(), userID, goalID, req.TargetBooks)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, progress)
}

func (h *GoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	var req createGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	progress, err := h.goalSvc.CreateGoal(r.Context(), userID, req.Year, req.TargetBooks)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, progress)
}

func (h *GoalHandler) GetCurrentYear(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	progress, err := h.goalSvc.GetCurrentYearGoal(r.Context(), userID)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	if progress == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, progress)
}
