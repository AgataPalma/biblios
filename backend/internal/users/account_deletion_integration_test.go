//go:build integration

package users

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AgataPalma/biblios/internal/tokenstore"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func TestAccountDeletionRevokesSessionsAndAnonymizesData(t *testing.T) {
	databaseURL := os.Getenv("BIBLIOS_TEST_DATABASE_URL")
	redisURL := os.Getenv("BIBLIOS_TEST_REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		t.Skip("BIBLIOS_TEST_DATABASE_URL and BIBLIOS_TEST_REDIS_URL must be set")
	}
	applyAccountDeletionTestMigrations(t, databaseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOptions)
	t.Cleanup(func() { _ = redisClient.Close() })
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping Redis: %v", err)
	}

	suffix := time.Now().UTC().Format("20060102150405.000000000")
	username := "delete-user-" + suffix
	email := username + "@example.test"
	repo := NewRepository(pool)
	user, err := repo.CreateUser(ctx, email, username, "integration-test-hash")
	if err != nil {
		t.Fatalf("create deletion test user: %v", err)
	}
	inviter, err := repo.CreateUser(ctx, "delete-inviter-"+suffix+"@example.test", "delete-inviter-"+suffix, "integration-test-hash")
	if err != nil {
		t.Fatalf("create inviter: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE users SET bio='private bio', avatar_url='https://private.example/avatar.png'
		WHERE id=$1`, user.ID); err != nil {
		t.Fatalf("add private profile data: %v", err)
	}
	if err := repo.CreateDefaultLibrary(ctx, user.ID); err != nil {
		t.Fatalf("create default library: %v", err)
	}
	if err := repo.CreateDefaultLibrary(ctx, inviter.ID); err != nil {
		t.Fatalf("create inviter library: %v", err)
	}

	var libraryID, inviterLibraryID string
	if err := pool.QueryRow(ctx, `SELECT id FROM libraries WHERE owner_id=$1`, user.ID).Scan(&libraryID); err != nil {
		t.Fatalf("find default library: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM libraries WHERE owner_id=$1`, inviter.ID).Scan(&inviterLibraryID); err != nil {
		t.Fatalf("find inviter library: %v", err)
	}

	var bookID, editionID, copyID, borrowedCopyID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO books (title, status) VALUES ($1, 'approved') RETURNING id`,
		"Deletion privacy "+suffix).Scan(&bookID); err != nil {
		t.Fatalf("insert book: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO book_editions (book_id, title, format, language, status)
		VALUES ($1, $2, 'paperback', 'en', 'approved') RETURNING id`,
		bookID, "Deletion privacy "+suffix).Scan(&editionID); err != nil {
		t.Fatalf("insert edition: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO book_copies
			(edition_id, owner_id, reading_status, current_page, location, personal_notes)
		VALUES ($1, $2, 'reading', 25, 'Private room', 'Private copy note')
		RETURNING id`, editionID, user.ID).Scan(&copyID); err != nil {
		t.Fatalf("insert copy: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO book_copies (edition_id, owner_id, borrowed_from)
		VALUES ($1, $2, $3) RETURNING id`, editionID, inviter.ID, user.ID).Scan(&borrowedCopyID); err != nil {
		t.Fatalf("insert borrowed copy: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_copy_states (user_id, copy_id, reading_status, current_page, personal_notes)
		VALUES ($1, $2, 'reading', 25, 'Private state note')`, user.ID, copyID); err != nil {
		t.Fatalf("insert private copy state: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO library_book_copies (library_id, book_copy_id) VALUES ($1, $2)`, libraryID, copyID); err != nil {
		t.Fatalf("link library copy: %v", err)
	}

	var shelfID, collectionID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO shelves (user_id, name) VALUES ($1, $2) RETURNING id`,
		user.ID, "Private shelf "+suffix).Scan(&shelfID); err != nil {
		t.Fatalf("insert shelf: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO shelf_books (shelf_id, copy_id) VALUES ($1, $2)`, shelfID, copyID); err != nil {
		t.Fatalf("link shelf copy: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO collections (library_id, created_by, name)
		VALUES ($1, $2, $3) RETURNING id`, libraryID, user.ID, "Private collection "+suffix).Scan(&collectionID); err != nil {
		t.Fatalf("insert collection: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO collection_books (collection_id, book_copy_id, added_by)
		VALUES ($1, $2, $3)`, collectionID, copyID, user.ID); err != nil {
		t.Fatalf("link collection copy: %v", err)
	}

	var reviewID, invitationID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO reviews (book_id, user_id, rating, body)
		VALUES ($1, $2, 4, 'Review retained without attribution') RETURNING id`,
		bookID, user.ID).Scan(&reviewID); err != nil {
		t.Fatalf("insert review: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO notifications (user_id, type, title, body)
		VALUES ($1, 'account-test', 'Private notification', 'Private body')`, user.ID); err != nil {
		t.Fatalf("insert notification: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO email_queue (user_id, to_email, subject, body)
		VALUES ($1, $2, 'Private email', 'Private queued body')`, user.ID, email); err != nil {
		t.Fatalf("insert queued email: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO reading_sessions (user_id, copy_id, logged_date, pages_read, note)
		VALUES ($1, $2, CURRENT_DATE, 10, 'Private session note')`, user.ID, copyID); err != nil {
		t.Fatalf("insert reading session: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO reading_challenges (user_id, year, goal)
		VALUES ($1, 2099, 12)`, user.ID); err != nil {
		t.Fatalf("insert reading challenge: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO import_jobs (user_id, status, error_message)
		VALUES ($1, 'failed', 'Private import error')`, user.ID); err != nil {
		t.Fatalf("insert import job: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO review_likes (review_id, user_id) VALUES ($1, $2)`, reviewID, user.ID); err != nil {
		t.Fatalf("insert review like: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_follows (follower_id, following_id) VALUES ($1, $2), ($2, $1)`,
		user.ID, inviter.ID); err != nil {
		t.Fatalf("insert user follows: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO quotes (book_id, user_id, body) VALUES ($1, $2, 'Retained anonymous quote')`,
		bookID, user.ID); err != nil {
		t.Fatalf("insert quote: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO library_invitations
			(library_id, invited_by, invited_user_id, invited_email, token, expires_at)
		VALUES ($1, $2, $3, $4, $5, NOW()+INTERVAL '1 day')
		RETURNING id`, inviterLibraryID, inviter.ID, user.ID, email, "delete-token-"+suffix).Scan(&invitationID); err != nil {
		t.Fatalf("insert invitation: %v", err)
	}

	sessionStore := tokenstore.NewStore(redisClient)
	for _, tokenID := range []string{"session-a-" + suffix, "session-b-" + suffix} {
		if err := sessionStore.StoreToken(ctx, user.ID, tokenID, time.Hour); err != nil {
			t.Fatalf("store session %q: %v", tokenID, err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_ = sessionStore.EnableUser(cleanupCtx, user.ID)
		_ = sessionStore.DeleteAllUserSessions(cleanupCtx, user.ID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE id=$1 OR id=$2`, user.ID, inviter.ID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE email=$1`, email)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM books WHERE id=$1`, bookID)
	})

	if exists, err := sessionStore.SessionExists(ctx, user.ID, "session-a-"+suffix); err != nil || !exists {
		t.Fatalf("session should initially exist: exists=%v err=%v", exists, err)
	}
	if err := sessionStore.DisableUser(ctx, user.ID); err != nil {
		t.Fatalf("disable user sessions: %v", err)
	}
	if exists, err := sessionStore.SessionExists(ctx, user.ID, "session-a-"+suffix); err != nil || exists {
		t.Fatalf("disabled session authenticated: exists=%v err=%v", exists, err)
	}
	if err := sessionStore.StoreToken(ctx, user.ID, "racing-login-"+suffix, time.Hour); !errors.Is(err, tokenstore.ErrUserDisabled) {
		t.Fatalf("racing login error = %v, want %v", err, tokenstore.ErrUserDisabled)
	}

	if err := repo.SoftDelete(ctx, user.ID); err != nil {
		t.Fatalf("soft delete account: %v", err)
	}
	if err := sessionStore.DeleteAllUserSessions(ctx, user.ID); err != nil {
		t.Fatalf("clean account sessions: %v", err)
	}
	for _, tokenID := range []string{"session-a-" + suffix, "session-b-" + suffix} {
		if count := redisClient.Exists(ctx, "session:"+user.ID+":"+tokenID).Val(); count != 0 {
			t.Fatalf("session key %q was not cleaned", tokenID)
		}
	}

	var deletedAt *time.Time
	var storedEmail, storedUsername, passwordHash, role, theme string
	var isAdmin bool
	var bio, avatarURL *string
	if err := pool.QueryRow(ctx, `
		SELECT email, username, password_hash, role, is_admin, theme, bio, avatar_url, deleted_at
		FROM users WHERE id=$1`, user.ID).Scan(
		&storedEmail, &storedUsername, &passwordHash, &role, &isAdmin, &theme, &bio, &avatarURL, &deletedAt,
	); err != nil {
		t.Fatalf("read deleted user: %v", err)
	}
	if deletedAt == nil || storedEmail == email || storedUsername == username || bio != nil || avatarURL != nil {
		t.Fatalf("deleted account retained profile data")
	}
	if !strings.HasSuffix(storedEmail, "@invalid.local") || !strings.HasPrefix(storedUsername, "deleted_") {
		t.Fatalf("unexpected anonymized identity: %q / %q", storedEmail, storedUsername)
	}
	if passwordHash != "deleted-account" || role != "user" || isAdmin || theme != "default-light" {
		t.Fatalf("deleted account retained authentication or privilege state")
	}

	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM libraries WHERE id=$1 AND deleted_at IS NOT NULL`, libraryID, 1)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM library_members WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM user_copy_states WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM library_book_copies WHERE book_copy_id=$1`, copyID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM shelf_books WHERE copy_id=$1`, copyID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM collection_books WHERE book_copy_id=$1`, copyID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM shelves WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM collections WHERE id=$1 AND deleted_at IS NOT NULL`, collectionID, 1)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM notifications WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM email_queue WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM reading_sessions WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM reading_challenges WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM import_jobs WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM review_likes WHERE user_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM user_follows WHERE follower_id=$1 OR following_id=$1`, user.ID, 0)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM reviews WHERE id=$1 AND user_id IS NULL`, reviewID, 1)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM quotes WHERE book_id=$1 AND user_id IS NULL`, bookID, 1)
	assertAccountDeletionCount(t, ctx, pool, `SELECT COUNT(*) FROM book_copies WHERE id=$1 AND borrowed_from IS NULL`, borrowedCopyID, 1)

	var copyDeletedAt *time.Time
	var location, personalNotes *string
	if err := pool.QueryRow(ctx, `
		SELECT deleted_at, location, personal_notes FROM book_copies WHERE id=$1`, copyID).Scan(
		&copyDeletedAt, &location, &personalNotes,
	); err != nil {
		t.Fatalf("read deleted copy: %v", err)
	}
	if copyDeletedAt == nil || location != nil || personalNotes != nil {
		t.Fatalf("deleted copy retained personal state")
	}

	var invitationStatus string
	var invitedUserID, invitedEmail *string
	if err := pool.QueryRow(ctx, `
		SELECT status, invited_user_id, invited_email
		FROM library_invitations WHERE id=$1`, invitationID).Scan(
		&invitationStatus, &invitedUserID, &invitedEmail,
	); err != nil {
		t.Fatalf("read revoked invitation: %v", err)
	}
	if invitationStatus != "revoked" || invitedUserID != nil || invitedEmail == nil || *invitedEmail == email {
		t.Fatalf("invitation retained deleted recipient data")
	}

	// Anonymizing the unique fields allows a future account to use the former
	// email and username without resurrecting the deleted identity.
	if _, err := repo.CreateUser(ctx, email, username, "new-account-hash"); err != nil {
		t.Fatalf("reuse anonymized email and username: %v", err)
	}
}

func applyAccountDeletionTestMigrations(t *testing.T, databaseURL string) {
	t.Helper()
	migrationsPath, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatalf("resolve migrations path: %v", err)
	}
	migrateURL := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	migrator, err := migrate.New("file://"+migrationsPath, migrateURL)
	if err != nil {
		t.Fatalf("create migrator: %v", err)
	}
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("apply migrations: %v", err)
	}
	if sourceErr, databaseErr := migrator.Close(); sourceErr != nil || databaseErr != nil {
		t.Fatalf("close migrator: source=%v database=%v", sourceErr, databaseErr)
	}
}

func assertAccountDeletionCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, id string, want int) {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, query, id).Scan(&count); err != nil {
		t.Fatalf("count account deletion rows: %v", err)
	}
	if count != want {
		t.Fatalf("account deletion count = %d, want %d for %s", count, want, query)
	}
}
