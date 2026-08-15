//go:build integration

package library

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLibraryAuthorizationAndPrivacyPostgres(t *testing.T) {
	databaseURL := os.Getenv("BIBLIOS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BIBLIOS_TEST_DATABASE_URL is not set")
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	suffix := time.Now().UTC().Format("20060102150405.000000000")
	ownerID := insertLibraryTestUser(t, ctx, pool, "owner-"+suffix)
	memberID := insertLibraryTestUser(t, ctx, pool, "member-"+suffix)
	visitorID := insertLibraryTestUser(t, ctx, pool, "visitor-"+suffix)

	var bookID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO books (title, status)
		VALUES ($1, 'approved')
		RETURNING id`, "Privacy test "+suffix).Scan(&bookID); err != nil {
		t.Fatalf("insert book: %v", err)
	}
	var editionID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO book_editions (book_id, title, format, language, status)
		VALUES ($1, $2, 'paperback', 'en', 'approved')
		RETURNING id`, bookID, "Privacy test "+suffix).Scan(&editionID); err != nil {
		t.Fatalf("insert edition: %v", err)
	}
	var contributorID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO contributors (name, status)
		VALUES ($1, 'approved')
		RETURNING id`, "Test Author "+suffix).Scan(&contributorID); err != nil {
		t.Fatalf("insert contributor: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO book_contributors (book_id, contributor_id, role)
		VALUES ($1, $2, 'author')`, bookID, contributorID); err != nil {
		t.Fatalf("link contributor: %v", err)
	}

	var ownerCopyID, memberCopyID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO book_copies (edition_id, owner_id, reading_status, current_page, location, personal_notes)
		VALUES ($1, $2, 'reading', 87, 'Owner bedroom', 'Owner private note')
		RETURNING id`, editionID, ownerID).Scan(&ownerCopyID); err != nil {
		t.Fatalf("insert owner copy: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO book_copies (edition_id, owner_id, reading_status, current_page, location, personal_notes)
		VALUES ($1, $2, 'reading', 42, 'Member office', 'Member private note')
		RETURNING id`, editionID, memberID).Scan(&memberCopyID); err != nil {
		t.Fatalf("insert member copy: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE id=$1 OR id=$2 OR id=$3`, ownerID, memberID, visitorID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM books WHERE id=$1`, bookID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM contributors WHERE id=$1`, contributorID)
	})

	repo := NewRepository(pool)
	service := NewService(repo)
	library, err := repo.CreateLibrary(ctx, ownerID, "Authorization test "+suffix, nil, true, "private")
	if err != nil {
		t.Fatalf("create library: %v", err)
	}
	if err := repo.AddMember(ctx, library.ID, memberID, LibraryMember{CanAdd: true}); err != nil {
		t.Fatalf("add member: %v", err)
	}

	if _, err := service.GetLibrary(ctx, library.ID, memberID); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("private library without can_view error = %v, want %v", err, ErrAccessDenied)
	}
	if err := service.AddBookToLibrary(ctx, library.ID, memberID, ownerCopyID); !errors.Is(err, ErrCopyNotOwned) {
		t.Fatalf("adding another user's copy error = %v, want %v", err, ErrCopyNotOwned)
	}
	if err := service.AddBookToLibrary(ctx, library.ID, memberID, memberCopyID); err != nil {
		t.Fatalf("add owned copy: %v", err)
	}

	visibility := "public"
	if _, err := repo.UpdateLibrary(ctx, library.ID, nil, nil, &visibility); err != nil {
		t.Fatalf("make library public: %v", err)
	}
	publicBooks, total, err := service.ListLibraryBooks(ctx, library.ID, visitorID, 1, 20)
	if err != nil {
		t.Fatalf("list public library books: %v", err)
	}
	if total != 1 || len(publicBooks) != 1 {
		t.Fatalf("public library books count = %d/%d, want 1/1", len(publicBooks), total)
	}
	if publicBooks[0].CopyID != "" {
		t.Fatalf("public payload exposed copy ID %q", publicBooks[0].CopyID)
	}
	if len(publicBooks[0].Authors) != 1 || !strings.HasPrefix(publicBooks[0].Authors[0], "Test Author ") {
		t.Fatalf("public catalogue authors = %#v", publicBooks[0].Authors)
	}
	assertNoPrivateLibraryFields(t, publicBooks[0])

	privateBooks, privateTotal, err := service.ListMyLibraryBooks(ctx, library.ID, memberID, 1, 20)
	if !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("private endpoint without can_view error = %v, want %v", err, ErrAccessDenied)
	}
	if len(privateBooks) != 0 || privateTotal != 0 {
		t.Fatalf("denied private endpoint returned books")
	}

	if err := repo.UpdateMemberPermissions(ctx, library.ID, memberID, LibraryMember{CanView: true, CanAdd: true}); err != nil {
		t.Fatalf("grant can_view: %v", err)
	}
	memberBooks, _, err := service.ListLibraryBooks(ctx, library.ID, memberID, 1, 20)
	if err != nil {
		t.Fatalf("list member library books: %v", err)
	}
	if memberBooks[0].CopyID != memberCopyID {
		t.Fatalf("member copy ID = %q, want %q", memberBooks[0].CopyID, memberCopyID)
	}
	privateBooks, privateTotal, err = service.ListMyLibraryBooks(ctx, library.ID, memberID, 1, 20)
	if err != nil {
		t.Fatalf("list member private books: %v", err)
	}
	if privateTotal != 1 || len(privateBooks) != 1 || privateBooks[0].CopyID != memberCopyID {
		t.Fatalf("private books were not scoped to the member's copy")
	}

	if err := service.RemoveMember(ctx, library.ID, ownerID, memberID); err != nil {
		t.Fatalf("remove member: %v", err)
	}
	var remainingMemberships, remainingLibraryCopies int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM library_members WHERE library_id=$1 AND user_id=$2`,
		library.ID, memberID).Scan(&remainingMemberships); err != nil {
		t.Fatalf("count removed memberships: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM library_book_copies WHERE library_id=$1 AND book_copy_id=$2`,
		library.ID, memberCopyID).Scan(&remainingLibraryCopies); err != nil {
		t.Fatalf("count removed member copies: %v", err)
	}
	if remainingMemberships != 0 || remainingLibraryCopies != 0 {
		t.Fatalf("removed member left membership=%d library copies=%d", remainingMemberships, remainingLibraryCopies)
	}
}

func insertLibraryTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, username string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, 'integration-test-only')
		RETURNING id`, username+"@example.test", username).Scan(&id); err != nil {
		t.Fatalf("insert user %q: %v", username, err)
	}
	return id
}

func assertNoPrivateLibraryFields(t *testing.T, book LibraryBook) {
	t.Helper()
	payload, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("marshal library book: %v", err)
	}
	for _, field := range []string{
		"reading_status", "current_page", "started_reading_at", "finished_reading_at",
		"borrowed_from", "location", "condition", "personal_notes", "reread_count",
	} {
		if strings.Contains(string(payload), field) {
			t.Fatalf("public library payload leaked %q: %s", field, payload)
		}
	}
}
