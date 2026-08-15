//go:build integration

package users

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAtomicRegistrationPostgres(t *testing.T) {
	databaseURL := os.Getenv("BIBLIOS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BIBLIOS_TEST_DATABASE_URL is not set")
	}
	applyRegistrationTestMigrations(t, databaseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	suffix := time.Now().UTC().Format("20060102150405.000000000")
	cleanupPattern := "%" + suffix + "%"
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DROP TRIGGER IF EXISTS reject_atomic_registration ON libraries`)
		_, _ = pool.Exec(cleanupCtx, `DROP FUNCTION IF EXISTS reject_atomic_registration_library()`)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE email LIKE $1 OR username LIKE $1`, cleanupPattern)
	})

	repo := NewRepository(pool)
	service := NewService(repo)
	password := "atomic-password"
	primaryEmail := "atomic.primary." + suffix + "@example.test"
	primaryUsername := "AtomicPrimary" + suffix
	primary, err := service.Register(ctx, RegisterInput{
		Email:    "  " + strings.ToUpper(primaryEmail) + "  ",
		Username: "  " + primaryUsername + "  ",
		Password: password,
	})
	if err != nil {
		t.Fatalf("register normalized account: %v", err)
	}
	if primary.Email != primaryEmail || primary.Username != primaryUsername {
		t.Fatalf("normalized identity = %q/%q, want %q/%q", primary.Email, primary.Username, primaryEmail, primaryUsername)
	}
	assertDefaultLibrary(t, ctx, pool, primary.ID)
	if _, err := service.Login(ctx, LoginInput{Email: "  " + strings.ToUpper(primaryEmail) + "  ", Password: password}); err != nil {
		t.Fatalf("login with case-insensitive email: %v", err)
	}

	secondaryEmail := "atomic.secondary." + suffix + "@example.test"
	secondaryUsername := "AtomicSecondary" + suffix
	secondary, err := service.Register(ctx, RegisterInput{Email: secondaryEmail, Username: secondaryUsername, Password: password})
	if err != nil {
		t.Fatalf("register secondary account: %v", err)
	}
	if err := service.UpdateEmail(ctx, secondary.ID, password, strings.ToUpper(primaryEmail)); !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Fatalf("case-insensitive email update error = %v, want %v", err, ErrEmailAlreadyRegistered)
	}
	conflictingUsername := strings.ToLower(primaryUsername)
	if _, err := service.UpdateProfile(ctx, secondary.ID, &conflictingUsername, nil, nil); !errors.Is(err, ErrUsernameAlreadyTaken) {
		t.Fatalf("case-insensitive username update error = %v, want %v", err, ErrUsernameAlreadyTaken)
	}

	t.Run("concurrent duplicate email creates exactly one complete account", func(t *testing.T) {
		email := "atomic.concurrent.email." + suffix + "@example.test"
		const attempts = 6
		results := runConcurrentRegistrations(ctx, service, attempts, func(index int) RegisterInput {
			candidateEmail := email
			if index%2 == 0 {
				candidateEmail = " " + strings.ToUpper(email) + " "
			}
			return RegisterInput{
				Email:    candidateEmail,
				Username: fmt.Sprintf("AtomicEmail%d%s", index, suffix),
				Password: password,
			}
		})
		assertConcurrentRegistrationResults(t, results, ErrEmailAlreadyRegistered)
		assertRegistrationRows(t, ctx, pool,
			`SELECT COUNT(*) FROM users WHERE LOWER(BTRIM(email))=LOWER(BTRIM($1))`, email, 1)
		assertCompleteAccountsByEmail(t, ctx, pool, email, 1)
	})

	t.Run("concurrent duplicate username creates exactly one complete account", func(t *testing.T) {
		username := "AtomicConcurrentUsername" + suffix
		const attempts = 6
		results := runConcurrentRegistrations(ctx, service, attempts, func(index int) RegisterInput {
			candidateUsername := username
			if index%2 == 0 {
				candidateUsername = strings.ToLower(username)
			}
			return RegisterInput{
				Email:    fmt.Sprintf("atomic.concurrent.username%d.%s@example.test", index, suffix),
				Username: candidateUsername,
				Password: password,
			}
		})
		assertConcurrentRegistrationResults(t, results, ErrUsernameAlreadyTaken)
		assertRegistrationRows(t, ctx, pool,
			`SELECT COUNT(*) FROM users WHERE LOWER(BTRIM(username))=LOWER(BTRIM($1))`, username, 1)
		assertCompleteAccountsByUsername(t, ctx, pool, username, 1)
	})

	t.Run("default library failure rolls back user", func(t *testing.T) {
		if _, err := pool.Exec(ctx, `
			CREATE OR REPLACE FUNCTION reject_atomic_registration_library()
			RETURNS TRIGGER AS $$
			BEGIN
				IF EXISTS (
					SELECT 1 FROM users
					WHERE id=NEW.owner_id AND email LIKE 'rollback.atomic.%'
				) THEN
					RAISE EXCEPTION 'forced default library failure';
				END IF;
				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;

			CREATE TRIGGER reject_atomic_registration
			BEFORE INSERT ON libraries
			FOR EACH ROW EXECUTE FUNCTION reject_atomic_registration_library()`); err != nil {
			t.Fatalf("install rollback trigger: %v", err)
		}

		rollbackEmail := "rollback.atomic." + suffix + "@example.test"
		_, registerErr := service.Register(ctx, RegisterInput{
			Email: rollbackEmail, Username: "RollbackAtomic" + suffix, Password: password,
		})
		if registerErr == nil {
			t.Fatal("registration unexpectedly succeeded through failing default-library trigger")
		}
		assertRegistrationRows(t, ctx, pool,
			`SELECT COUNT(*) FROM users WHERE LOWER(email)=LOWER($1)`, rollbackEmail, 0)

		if _, err := pool.Exec(ctx, `DROP TRIGGER reject_atomic_registration ON libraries`); err != nil {
			t.Fatalf("drop rollback trigger: %v", err)
		}
		if _, err := pool.Exec(ctx, `DROP FUNCTION reject_atomic_registration_library()`); err != nil {
			t.Fatalf("drop rollback function: %v", err)
		}
	})
}

func runConcurrentRegistrations(
	ctx context.Context,
	service *Service,
	attempts int,
	inputFor func(index int) RegisterInput,
) []error {
	results := make([]error, attempts)
	var waitGroup sync.WaitGroup
	waitGroup.Add(attempts)
	for index := range attempts {
		go func() {
			defer waitGroup.Done()
			_, results[index] = service.Register(ctx, inputFor(index))
		}()
	}
	waitGroup.Wait()
	return results
}

func assertConcurrentRegistrationResults(t *testing.T, results []error, conflict error) {
	t.Helper()
	var successes, conflicts int
	for _, result := range results {
		switch {
		case result == nil:
			successes++
		case errors.Is(result, conflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent registration error: %v", result)
		}
	}
	if successes != 1 || conflicts != len(results)-1 {
		t.Fatalf("concurrent registrations: success=%d conflict=%d, want 1/%d", successes, conflicts, len(results)-1)
	}
}

func assertDefaultLibrary(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID string) {
	t.Helper()
	var libraries, ownerMemberships int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM libraries
		WHERE owner_id=$1 AND name='My Library' AND visibility='private' AND deleted_at IS NULL`, userID).Scan(&libraries); err != nil {
		t.Fatalf("count default libraries: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM library_members lm
		JOIN libraries l ON l.id=lm.library_id
		WHERE lm.user_id=$1 AND l.owner_id=$1
		  AND lm.is_owner AND lm.can_view AND lm.can_add AND lm.can_remove
		  AND lm.can_edit AND lm.can_invite AND lm.can_manage_members`, userID).Scan(&ownerMemberships); err != nil {
		t.Fatalf("count default owner memberships: %v", err)
	}
	if libraries != 1 || ownerMemberships != 1 {
		t.Fatalf("default inventory: libraries=%d owner memberships=%d, want 1/1", libraries, ownerMemberships)
	}
}

func assertCompleteAccountsByEmail(t *testing.T, ctx context.Context, pool *pgxpool.Pool, email string, want int) {
	t.Helper()
	assertRegistrationRows(t, ctx, pool, `
		SELECT COUNT(*)
		FROM users u
		JOIN libraries l ON l.owner_id=u.id AND l.deleted_at IS NULL
		JOIN library_members lm ON lm.library_id=l.id AND lm.user_id=u.id AND lm.is_owner
		WHERE LOWER(BTRIM(u.email))=LOWER(BTRIM($1))`, email, want)
}

func assertCompleteAccountsByUsername(t *testing.T, ctx context.Context, pool *pgxpool.Pool, username string, want int) {
	t.Helper()
	assertRegistrationRows(t, ctx, pool, `
		SELECT COUNT(*)
		FROM users u
		JOIN libraries l ON l.owner_id=u.id AND l.deleted_at IS NULL
		JOIN library_members lm ON lm.library_id=l.id AND lm.user_id=u.id AND lm.is_owner
		WHERE LOWER(BTRIM(u.username))=LOWER(BTRIM($1))`, username, want)
}

func assertRegistrationRows(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query, value string, want int) {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, query, value).Scan(&count); err != nil {
		t.Fatalf("count registration rows: %v", err)
	}
	if count != want {
		t.Fatalf("registration rows = %d, want %d for %s", count, want, query)
	}
}

func applyRegistrationTestMigrations(t *testing.T, databaseURL string) {
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
