package main

import (
	"context"
	"fmt"
	"github.com/AgataPalma/biblios/internal/auth"
	"github.com/AgataPalma/biblios/internal/books"
	"github.com/AgataPalma/biblios/internal/config"
	"github.com/AgataPalma/biblios/internal/database"
	"github.com/AgataPalma/biblios/internal/lookup"
	"github.com/AgataPalma/biblios/internal/middleware"
	"github.com/AgataPalma/biblios/internal/tokenstore"
	"github.com/AgataPalma/biblios/internal/users"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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

	// Token store
	var tStore *tokenstore.Store = tokenstore.NewStore(rdb)

	// Repositories and services
	var userRepo *users.Repository = users.NewRepository(db)
	var userService *users.Service = users.NewService(userRepo)
	var authHandler *auth.Handler = auth.NewHandler(userService, cfg.JWTSecret, tStore)

	// Books
	var bookRepo *books.Repository = books.NewRepository(db)
	var bookService *books.Service = books.NewService(bookRepo, db)
	var bookHandler *books.Handler = books.NewHandler(bookService)

	// Lookup
	var lookupService *lookup.Service = lookup.NewService(cfg.GoogleBooksAPIKey)
	var lookupHandler *lookup.Handler = lookup.NewHandler(lookupService)

	// Router
	var r *chi.Mux = chi.NewRouter()
	//    r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWTSecret, tStore))
			r.Get("/auth/me", authHandler.Me)
			r.Post("/auth/logout", authHandler.Logout)
			r.Post("/books", bookHandler.SubmitBook)
			r.Get("/books/lookup", lookupHandler.Lookup)
		})
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
