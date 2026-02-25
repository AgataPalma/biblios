package main

import (
	"context"
	"fmt"
	"github.com/AgataPalma/biblios/internal/auth"
	"github.com/AgataPalma/biblios/internal/config"
	"github.com/AgataPalma/biblios/internal/database"
	"github.com/AgataPalma/biblios/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	//logs
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, nil)
	var logger *slog.Logger = slog.New(handler)
	slog.SetDefault(logger)

	//or
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// slog.SetDefault(logger)

	// Config
	var cfg config.Config = config.Load()
	//cfg := config.Load()

	// Postgres
	var db *pgxpool.Pool
	var err error
	db, err = pgxpool.New(context.Background(), cfg.DatabaseURL)
	//db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		slog.Error("postgres ping failed", "error", err)
		os.Exit(1)
	}
	slog.Info("postgres connected")

	// Run migrations
	var dbURL string = cfg.DatabaseURL
	if err = database.RunMigrations(dbURL); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	// Redis
	var opts *redis.Options
	opts, err = redis.ParseURL(cfg.RedisURL)
	// opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse redis URL", "error", err)
		os.Exit(1)
	}
	var rdb *redis.Client = redis.NewClient(opts)
	//    rdb := redis.NewClient(opts)
	defer rdb.Close()

	if err = rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("redis ping failed", "error", err)
		os.Exit(1)
	}
	slog.Info("redis connected")

	// Repositories and services
	var userRepo *users.Repository = users.NewRepository(db)
	var userService *users.Service = users.NewService(userRepo)
	var authHandler *auth.Handler = auth.NewHandler(userService, cfg.JWTSecret)
	// Router
	var r *chi.Mux = chi.NewRouter()
	//    r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	slog.Info("server starting", "port", cfg.Port)
	if err = http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
