package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/SEU_USUARIO/mybooklist/backend/internal/config"
	"github.com/SEU_USUARIO/mybooklist/backend/internal/handlers"
	appmiddleware "github.com/SEU_USUARIO/mybooklist/backend/internal/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := db.PingContext(pingCtx); err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.SlogLogger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", handlers.Health(db))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}

	db.Close()
	slog.Info("shutdown complete")
}
