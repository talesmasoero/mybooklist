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

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/talesmasoero/mybooklist/backend/internal/config"
	"github.com/talesmasoero/mybooklist/backend/internal/googlebooks"
	"github.com/talesmasoero/mybooklist/backend/internal/handlers"
	appmiddleware "github.com/talesmasoero/mybooklist/backend/internal/middleware"
	"github.com/talesmasoero/mybooklist/backend/internal/repositories"
	"github.com/talesmasoero/mybooklist/backend/internal/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := godotenv.Load("../.env"); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}
	if err := godotenv.Overload("../.env.local"); err != nil {
		slog.Warn("no .env.local file found, skipping local overrides")
	}

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

	userRepo := repositories.NewPostgresUserRepository(db)
	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(authSvc)
	userSvc := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	bookRepo := repositories.NewPostgresBookRepository(db)
	readingRepo := repositories.NewPostgresReadingRepository(db)
	googleBooksClient := googlebooks.NewClient(cfg.GoogleBooksAPIKey)
	bookSvc := services.NewBookService(bookRepo, readingRepo, googleBooksClient)
	bookHandler := handlers.NewBookHandler(bookSvc)

	sessionRepo := repositories.NewPostgresSessionRepository(db)
	sessionSvc := services.NewSessionService(sessionRepo, readingRepo)
	sessionHandler := handlers.NewSessionHandler(sessionSvc)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.SlogLogger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", handlers.Health(db))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.JWTAuth(cfg.JWTSecret))

			r.Get("/books/search", bookHandler.Search)
			r.Post("/library", bookHandler.AddToLibrary)
			r.Get("/library", bookHandler.ListLibrary)
			r.Patch("/library/{id}/status", bookHandler.UpdateLibraryStatus)

			r.Post("/readings/{readingId}/sessions", sessionHandler.Create)
			r.Get("/readings/{readingId}/sessions", sessionHandler.List)
			r.Patch("/sessions/{sessionId}", sessionHandler.Update)
			r.Delete("/sessions/{sessionId}", sessionHandler.Delete)

			r.Get("/me", userHandler.GetProfile)
			r.Patch("/me", userHandler.UpdateName)
			r.Patch("/me/password", userHandler.UpdatePassword)
			r.Delete("/me", userHandler.DeleteAccount)
		})
	})

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
