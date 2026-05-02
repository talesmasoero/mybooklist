package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	appmiddleware "github.com/talesmasoero/mybooklist/backend/internal/middleware"
	"github.com/talesmasoero/mybooklist/backend/internal/services"
)

type BookHandler struct {
	bookSvc services.BookService
}

func NewBookHandler(bookSvc services.BookService) *BookHandler {
	return &BookHandler{bookSvc: bookSvc}
}

func (h *BookHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	maxResults := 10
	if m := r.URL.Query().Get("max"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	results, err := h.bookSvc.SearchExternal(r.Context(), q, maxResults)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *BookHandler) AddToLibrary(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	var payload services.AddToLibraryPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_VALIDATION", "invalid request body")
		return
	}

	reading, err := h.bookSvc.AddToLibrary(r.Context(), userID, payload)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, reading)
}

func (h *BookHandler) ListLibrary(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "missing user context")
		return
	}

	var statusFilter *string
	if s := r.URL.Query().Get("status"); s != "" {
		statusFilter = &s
	}

	readings, err := h.bookSvc.ListLibrary(r.Context(), userID, statusFilter)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	if readings == nil {
		readings = []domain.Reading{}
	}
	writeJSON(w, http.StatusOK, readings)
}
