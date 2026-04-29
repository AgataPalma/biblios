package main

import (
	"context"
	"fmt"
	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/auth"
	"github.com/AgataPalma/biblios/internal/books"
	"github.com/AgataPalma/biblios/internal/collections"
	"github.com/AgataPalma/biblios/internal/config"
	"github.com/AgataPalma/biblios/internal/contributors"
	"github.com/AgataPalma/biblios/internal/database"
	"github.com/AgataPalma/biblios/internal/genres"
	"github.com/AgataPalma/biblios/internal/library"
	"github.com/AgataPalma/biblios/internal/lookup"
	"github.com/AgataPalma/biblios/internal/middleware"
	"github.com/AgataPalma/biblios/internal/moderation"
	"github.com/AgataPalma/biblios/internal/notifications"
	"github.com/AgataPalma/biblios/internal/reading"
	"github.com/AgataPalma/biblios/internal/reviews"
	"github.com/AgataPalma/biblios/internal/series"
	"github.com/AgataPalma/biblios/internal/shelves"
	"github.com/AgataPalma/biblios/internal/tokenstore"
	"github.com/AgataPalma/biblios/internal/users"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	// ── Structured logging ────────────────────────────────────────────────────
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// ── Config ────────────────────────────────────────────────────────────────
	cfg := config.Load()

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
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

	// ── Migrations ────────────────────────────────────────────────────────────
	if err = database.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse redis URL", "error", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()
	if err = rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("redis ping failed", "error", err)
		os.Exit(1)
	}
	slog.Info("redis connected")

	// ── Token store ───────────────────────────────────────────────────────────
	tStore := tokenstore.NewStore(rdb)

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := users.NewRepository(db)
	bookRepo := books.NewRepository(db)
	libraryRepo := library.NewRepository(db)
	collectionRepo := collections.NewRepository(db)
	reviewRepo := reviews.NewRepository(db)
	notifRepo := notifications.NewRepository(db)
	readingRepo := reading.NewRepository(db)
	shelfRepo := shelves.NewRepository(db)
	contributorRepo := contributors.NewRepository(db)
	seriesRepo := series.NewRepository(db)
	genreRepo := genres.NewRepository(db)

	// ── Services ──────────────────────────────────────────────────────────────
	userSvc := users.NewService(userRepo)
	bookSvc := books.NewService(bookRepo, db)
	librarySvc := library.NewService(libraryRepo)
	collectionSvc := collections.NewService(collectionRepo)
	reviewSvc := reviews.NewService(reviewRepo)
	notifSvc := notifications.NewService(notifRepo)
	readingSvc := reading.NewService(readingRepo)
	shelfSvc := shelves.NewService(shelfRepo)
	contributorSvc := contributors.NewService(contributorRepo)
	seriesSvc := series.NewService(seriesRepo)
	genreSvc := genres.NewService(genreRepo)
	lookupSvc := lookup.NewService(cfg.GoogleBooksAPIKey)
	moderationSvc := moderation.NewService(bookRepo, db)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := auth.NewHandler(userSvc, cfg.JWTSecret, tStore)
	userHandler := users.NewHandler(userSvc)
	bookHandler := books.NewHandler(bookSvc, lookupSvc, cfg.CoversDir)
	lookupHandler := lookup.NewHandler(lookupSvc)
	libraryHandler := library.NewHandler(librarySvc)
	collectionHandler := collections.NewHandler(collectionSvc)
	reviewHandler := reviews.NewHandler(reviewSvc)
	notifHandler := notifications.NewHandler(notifSvc)
	readingHandler := reading.NewHandler(readingSvc)
	shelfHandler := shelves.NewHandler(shelfSvc)
	contributorHandler := contributors.NewHandler(contributorSvc)
	seriesHandler := series.NewHandler(seriesSvc)
	genreHandler := genres.NewHandler(genreSvc)
	moderationHandler := moderation.NewHandler(moderationSvc)

	// Suppress unused variable warnings for services not directly used in routes
	_ = notifSvc

	// ── Router ────────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://biblioslibrary.app", "https://biblioslibrary.dev"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// ── Health check ──────────────────────────────────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// ── Static covers ─────────────────────────────────────────────────────────
	if err = os.MkdirAll(cfg.CoversDir, 0755); err != nil {
		slog.Error("failed to create covers directory", "error", err)
		os.Exit(1)
	}
	r.Handle("/covers/*", http.StripPrefix("/covers/", http.FileServer(http.Dir(cfg.CoversDir))))

	// ── API v1 ────────────────────────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		registerRoutes(r, cfg, tStore,
			authHandler, userHandler, bookHandler, lookupHandler,
			libraryHandler, collectionHandler, reviewHandler,
			notifHandler, readingHandler, shelfHandler,
			contributorHandler, seriesHandler, genreHandler,
			moderationHandler,
		)
	})

	slog.Info("server starting", "port", cfg.Port)
	if err = http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func registerRoutes(
	r chi.Router,
	cfg config.Config,
	tStore *tokenstore.Store,
	authH *auth.Handler,
	userH *users.Handler,
	bookH *books.Handler,
	lookupH *lookup.Handler,
	libraryH *library.Handler,
	collectionH *collections.Handler,
	reviewH *reviews.Handler,
	notifH *notifications.Handler,
	readingH *reading.Handler,
	shelfH *shelves.Handler,
	contributorH *contributors.Handler,
	seriesH *series.Handler,
	genreH *genres.Handler,
	moderationH *moderation.Handler,
) {
	// ── Public ────────────────────────────────────────────────────────────────
	r.Post("/auth/register", authH.Register)
	r.Post("/auth/login", authH.Login)

	// ── Authenticated ─────────────────────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWTSecret, tStore))

		// Auth
		r.Get("/auth/me", userH.Me)
		r.Post("/auth/logout", authH.Logout)

		// Users
		r.Put("/users/me", userH.UpdateProfile)
		r.Put("/users/me/email", userH.UpdateEmail)
		r.Put("/users/me/password", userH.UpdatePassword)
		r.Put("/users/me/theme", userH.UpdateTheme)
		r.Delete("/users/me", userH.DeleteUser)

		// Books — public catalogue
		r.Get("/books", bookH.ListBooks)
		r.Get("/books/lookup", lookupH.Lookup)
		r.Get("/books/check", bookH.CheckDuplicate)
		r.Post("/books", bookH.SubmitBook)
		r.Get("/books/{id}", bookH.GetBook)
		r.Post("/books/copies", bookH.AddCopy)
		r.Get("/users/me/books", bookH.GetMyBooks)
		r.Put("/books/copies/{id}/status", bookH.UpdateCopyStatus)
		r.Delete("/books/copies/{id}", bookH.RemoveCopy)

		// Libraries
		r.Get("/libraries", libraryH.ListMyLibraries)
		r.Post("/libraries", libraryH.CreateLibrary)
		r.Get("/libraries/public", libraryH.ListPublicLibraries)
		r.Get("/libraries/{id}", libraryH.GetLibrary)
		r.Put("/libraries/{id}", libraryH.UpdateLibrary)
		r.Delete("/libraries/{id}", libraryH.DeleteLibrary)
		r.Get("/libraries/{id}/members", libraryH.ListMembers)
		r.Put("/libraries/{id}/members/{userId}", libraryH.UpdateMemberPermissions)
		r.Delete("/libraries/{id}/members/{userId}", libraryH.RemoveMember)
		r.Post("/libraries/{id}/invite", libraryH.InviteMember)
		r.Get("/libraries/{id}/books", libraryH.ListLibraryBooks)
		r.Post("/libraries/{id}/books", libraryH.AddBookToLibrary)
		r.Delete("/libraries/{id}/books/{copyId}", libraryH.RemoveBookFromLibrary)
		r.Get("/users/me/library", libraryH.GetMyLibrary)

		// Invitations
		r.Get("/invitations", libraryH.ListMyInvitations)
		r.Post("/invitations/{token}/accept", libraryH.AcceptInvitation)
		r.Post("/invitations/{token}/decline", libraryH.DeclineInvitation)

		// Collections
		r.Get("/libraries/{id}/collections", collectionH.List)
		r.Post("/libraries/{id}/collections", collectionH.Create)
		r.Get("/libraries/{id}/collections/{collectionId}", collectionH.Get)
		r.Put("/libraries/{id}/collections/{collectionId}", collectionH.Update)
		r.Delete("/libraries/{id}/collections/{collectionId}", collectionH.Delete)
		r.Get("/libraries/{id}/collections/{collectionId}/books", collectionH.ListBooks)
		r.Post("/libraries/{id}/collections/{collectionId}/books", collectionH.AddBook)
		r.Delete("/libraries/{id}/collections/{collectionId}/books/{copyId}", collectionH.RemoveBook)

		// Reviews
		r.Get("/books/{id}/reviews", reviewH.ListPublicReviews)
		r.Get("/books/{id}/reviews/me", reviewH.GetMyReview)
		r.Post("/books/{id}/reviews", reviewH.UpsertReview)
		r.Delete("/books/{id}/reviews/me", reviewH.DeleteMyReview)
		r.Post("/reviews/{id}/like", reviewH.LikeReview)
		r.Delete("/reviews/{id}/like", reviewH.UnlikeReview)

		// Notifications
		r.Get("/notifications", notifH.List)
		r.Put("/notifications/read-all", notifH.MarkAllRead)
		r.Put("/notifications/{id}/read", notifH.MarkRead)

		// Reading
		r.Get("/reading/challenges", readingH.ListChallenges)
		r.Post("/reading/challenges", readingH.CreateChallenge)
		r.Delete("/reading/challenges/{id}", readingH.DeleteChallenge)
		r.Get("/reading/challenges/{id}/progress", readingH.GetProgress)
		r.Post("/reading/sessions", readingH.LogSession)
		r.Get("/reading/sessions", readingH.ListSessions)
		r.Get("/reading/stats", readingH.GetOverallStats)
		r.Get("/reading/stats/year/{year}", readingH.GetYearStats)
		r.Get("/reading/stats/month/{year}/{month}", readingH.GetMonthStats)

		// Shelves
		r.Get("/shelves", shelfH.List)
		r.Post("/shelves", shelfH.Create)
		r.Put("/shelves/{id}", shelfH.Rename)
		r.Delete("/shelves/{id}", shelfH.Delete)
		r.Get("/shelves/{id}/books", shelfH.ListBooks)
		r.Post("/shelves/{id}/books", shelfH.AddBook)
		r.Delete("/shelves/{id}/books/{copyId}", shelfH.RemoveBook)

		// Contributors
		r.Get("/contributors", contributorH.List)
		r.Get("/contributors/{id}", contributorH.Get)
		r.Post("/contributors", contributorH.Create)

		// Series
		r.Get("/series", seriesH.List)
		r.Get("/series/{id}", seriesH.Get)
		r.Post("/series", seriesH.Create)

		// Genres & Moods
		r.Get("/genres", genreH.ListGenres)
		r.Post("/genres", genreH.CreateGenre)
		r.Get("/moods", genreH.ListMoods)
		r.Post("/moods", genreH.CreateMood)

		// ── Moderator / Admin ─────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(apictx.RoleModerator, apictx.RoleAdmin))

			r.Put("/books/{id}", bookH.UpdateBook)
			r.Delete("/books/{id}", bookH.DeleteBook)
			r.Post("/books/{id}/cover", bookH.UploadCover)

			r.Get("/moderation/submissions", moderationH.ListPending)
			r.Get("/moderation/submissions/{id}", moderationH.GetSubmission)
			r.Put("/moderation/submissions/{id}/approve", moderationH.Approve)
			r.Put("/moderation/submissions/{id}/reject", moderationH.Reject)
			r.Put("/moderation/submissions/{id}/edit", moderationH.EditAndApprove)
			r.Get("/moderation/logs", moderationH.ListLogs)

			// ── Admin only ────────────────────────────────────────────────────
			r.With(middleware.RequireRole(apictx.RoleAdmin)).
				Post("/admin/backfill-covers", bookH.BackfillCovers)
		})
	})
}
