//go:build integration

package library

import (
	"context"
	"errors"
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

func TestLibraryInvitationSecurityPostgres(t *testing.T) {
	databaseURL := os.Getenv("BIBLIOS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BIBLIOS_TEST_DATABASE_URL is not set")
	}
	applyInvitationTestMigrations(t, databaseURL)

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
	ownerID, ownerEmail := insertInvitationTestUser(t, ctx, pool, "invite-owner-"+suffix)
	recipientID, recipientEmail := insertInvitationTestUser(t, ctx, pool, "invite-recipient-"+suffix)
	attackerID, _ := insertInvitationTestUser(t, ctx, pool, "invite-attacker-"+suffix)

	repo := NewRepository(pool)
	service := NewService(repo)
	library, err := repo.CreateLibrary(ctx, ownerID, "Invitation security "+suffix, nil, true, "private")
	if err != nil {
		t.Fatalf("create library: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `
			DELETE FROM users WHERE id=$1 OR id=$2 OR id=$3`,
			ownerID, recipientID, attackerID)
	})

	if _, err := service.InviteMember(ctx, library.ID, ownerID, "not-an-email"); !errors.Is(err, ErrInvalidInvitationEmail) {
		t.Fatalf("invalid email error = %v, want %v", err, ErrInvalidInvitationEmail)
	}
	if _, err := service.InviteMember(ctx, library.ID, ownerID, ownerEmail); !errors.Is(err, ErrInvitationRecipientMember) {
		t.Fatalf("existing member error = %v, want %v", err, ErrInvitationRecipientMember)
	}
	if _, err := repo.CreateInvitation(ctx, library.ID, attackerID, recipientEmail, time.Now().Add(time.Hour)); !errors.Is(err, ErrInvitationCreateForbidden) {
		t.Fatalf("repository authorization error = %v, want %v", err, ErrInvitationCreateForbidden)
	}

	invitation, err := service.InviteMember(ctx, library.ID, ownerID, "  "+recipientEmail+"  ")
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	if invitation.InvitedEmail != recipientEmail {
		t.Fatalf("normalized invitation email = %q, want %q", invitation.InvitedEmail, recipientEmail)
	}
	if invitation.InvitedUserID == nil || *invitation.InvitedUserID != recipientID {
		t.Fatalf("invitation did not resolve intended recipient")
	}
	if _, err := service.InviteMember(ctx, library.ID, ownerID, recipientEmail); !errors.Is(err, ErrInvitationAlreadyPending) {
		t.Fatalf("duplicate invitation error = %v, want %v", err, ErrInvitationAlreadyPending)
	}

	if err := service.AcceptInvitation(ctx, invitation.Token, attackerID); !errors.Is(err, ErrInvitationNotForUser) {
		t.Fatalf("attacker accept error = %v, want %v", err, ErrInvitationNotForUser)
	}
	if err := service.DeclineInvitation(ctx, invitation.Token, attackerID); !errors.Is(err, ErrInvitationNotForUser) {
		t.Fatalf("attacker decline error = %v, want %v", err, ErrInvitationNotForUser)
	}
	assertInvitationState(t, ctx, pool, invitation.ID, "pending")
	assertLibraryMembershipCount(t, ctx, pool, library.ID, attackerID, 0)

	recipientInvitations, err := service.ListMyInvitations(ctx, recipientID)
	if err != nil {
		t.Fatalf("list recipient invitations: %v", err)
	}
	if len(recipientInvitations) != 1 || recipientInvitations[0].ID != invitation.ID {
		t.Fatalf("recipient invitations = %#v", recipientInvitations)
	}
	if recipientInvitations[0].Token != "" {
		t.Fatalf("invitation list exposed bearer token")
	}
	attackerInvitations, err := service.ListMyInvitations(ctx, attackerID)
	if err != nil {
		t.Fatalf("list attacker invitations: %v", err)
	}
	if len(attackerInvitations) != 0 {
		t.Fatalf("attacker could list another user's invitation")
	}

	results := make(chan error, 2)
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	for range 2 {
		go func() {
			defer waitGroup.Done()
			results <- service.AcceptInvitation(ctx, invitation.Token, recipientID)
		}()
	}
	waitGroup.Wait()
	close(results)

	var accepted, alreadyConsumed int
	for result := range results {
		switch {
		case result == nil:
			accepted++
		case errors.Is(result, ErrInvitationNotPending):
			alreadyConsumed++
		default:
			t.Fatalf("concurrent accept returned unexpected error: %v", result)
		}
	}
	if accepted != 1 || alreadyConsumed != 1 {
		t.Fatalf("concurrent accepts: success=%d consumed=%d, want 1/1", accepted, alreadyConsumed)
	}
	assertInvitationState(t, ctx, pool, invitation.ID, "accepted")
	assertLibraryMembershipCount(t, ctx, pool, library.ID, recipientID, 1)

	// A later invitation can be declined only by its intended recipient.
	if err := repo.RemoveMember(ctx, library.ID, recipientID); err != nil {
		t.Fatalf("remove recipient before decline test: %v", err)
	}
	declineInvitation, err := service.InviteMember(ctx, library.ID, ownerID, recipientEmail)
	if err != nil {
		t.Fatalf("create decline invitation: %v", err)
	}
	if err := service.DeclineInvitation(ctx, declineInvitation.Token, recipientID); err != nil {
		t.Fatalf("decline intended invitation: %v", err)
	}
	assertInvitationState(t, ctx, pool, declineInvitation.ID, "declined")
	assertLibraryMembershipCount(t, ctx, pool, library.ID, recipientID, 0)

	// Expired invitations are transitioned and never appear in pending lists.
	expiredInvitation, err := repo.CreateInvitation(ctx, library.ID, ownerID, recipientEmail, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("create expired invitation: %v", err)
	}
	if err := service.AcceptInvitation(ctx, expiredInvitation.Token, recipientID); !errors.Is(err, ErrInvitationExpired) {
		t.Fatalf("expired accept error = %v, want %v", err, ErrInvitationExpired)
	}
	assertInvitationState(t, ctx, pool, expiredInvitation.ID, "expired")
	recipientInvitations, err = service.ListMyInvitations(ctx, recipientID)
	if err != nil {
		t.Fatalf("list after expiration: %v", err)
	}
	if len(recipientInvitations) != 0 {
		t.Fatalf("expired invitation appeared in pending list")
	}
}

func applyInvitationTestMigrations(t *testing.T, databaseURL string) {
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

func insertInvitationTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, username string) (string, string) {
	t.Helper()
	email := username + "@example.test"
	var id string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, 'integration-test-only')
		RETURNING id`, email, username).Scan(&id); err != nil {
		t.Fatalf("insert user %q: %v", username, err)
	}
	return id, email
}

func assertInvitationState(t *testing.T, ctx context.Context, pool *pgxpool.Pool, invitationID, want string) {
	t.Helper()
	var status string
	if err := pool.QueryRow(ctx, `
		SELECT status FROM library_invitations WHERE id=$1`, invitationID).Scan(&status); err != nil {
		t.Fatalf("read invitation state: %v", err)
	}
	if status != want {
		t.Fatalf("invitation state = %q, want %q", status, want)
	}
}

func assertLibraryMembershipCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, libraryID, userID string, want int) {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM library_members WHERE library_id=$1 AND user_id=$2`,
		libraryID, userID).Scan(&count); err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	if count != want {
		t.Fatalf("membership count = %d, want %d", count, want)
	}
}
