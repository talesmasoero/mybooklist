package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"
)

func Health(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if err := db.PingContext(ctx); err != nil {
			slog.ErrorContext(ctx, "health check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"error","message":"database unavailable"}`)) //nolint:errcheck
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	}
}
