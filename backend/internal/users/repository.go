package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const userColumns = `id, email, username, password_hash, role, is_admin, theme, bio, avatar_url, deleted_at, created_at, updated_at`

func scanUser(row pgx.Row, u *User) error {
	return row.Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash,
		&u.Role, &u.IsAdmin, &u.Theme, &u.Bio, &u.AvatarURL,
		&u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
}

// CreateUserWithDefaultLibrary creates the account, its default library, and
// the owner membership as a single transaction. No caller can observe a user
// without the default inventory required by the rest of the application.
func (r *Repository) CreateUserWithDefaultLibrary(ctx context.Context, email, username, passwordHash string) (User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin registration: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var user User
	if err := scanUser(tx.QueryRow(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING `+userColumns, email, username, passwordHash), &user); err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	var libraryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO libraries (owner_id, name, description, visibility, is_cooperative)
		VALUES ($1, 'My Library', 'My personal library', 'private', false)
		RETURNING id`, user.ID).Scan(&libraryID); err != nil {
		return User{}, fmt.Errorf("create default library: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO library_members
			(library_id, user_id, is_owner, can_view, can_add, can_remove, can_edit, can_invite, can_manage_members)
		VALUES ($1, $2, true, true, true, true, true, true, true)`, libraryID, user.ID); err != nil {
		return User{}, fmt.Errorf("create default library membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit registration: %w", err)
	}
	return user, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := scanUser(r.db.QueryRow(ctx, `
		SELECT `+userColumns+` FROM users
		WHERE LOWER(BTRIM(email))=LOWER(BTRIM($1)) AND deleted_at IS NULL`, email), &u)
	if err != nil {
		return User{}, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (User, error) {
	var u User
	err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id=$1 AND deleted_at IS NULL`, id), &u)
	if err != nil {
		return User{}, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, userID string, username, bio, avatarURL *string) (User, error) {
	var u User
	err := scanUser(r.db.QueryRow(ctx, `
		UPDATE users SET
			username   = COALESCE($2, username),
			bio        = COALESCE($3, bio),
			avatar_url = COALESCE($4, avatar_url),
			updated_at = NOW()
		WHERE id=$1 AND deleted_at IS NULL
		RETURNING `+userColumns, userID, username, bio, avatarURL), &u)
	if err != nil {
		return User{}, fmt.Errorf("update profile: %w", err)
	}
	return u, nil
}

func (r *Repository) UpdateEmail(ctx context.Context, userID, newEmail string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET email=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, userID, newEmail)
	return err
}

func (r *Repository) UpdatePassword(ctx context.Context, userID, hash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password_hash=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, userID, hash)
	return err
}

func (r *Repository) UpdateTheme(ctx context.Context, userID, theme string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET theme=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, userID, theme)
	return err
}

func (r *Repository) SoftDelete(ctx context.Context, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var currentEmail string
	if err := tx.QueryRow(ctx, `
		SELECT email FROM users
		WHERE id=$1 AND deleted_at IS NULL
		FOR UPDATE`, userID).Scan(&currentEmail); errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	} else if err != nil {
		return fmt.Errorf("lock user for deletion: %w", err)
	}

	anonymizedID := strings.ReplaceAll(userID, "-", "")
	anonymizedEmail := "deleted+" + anonymizedID + "@invalid.local"
	anonymizedUsername := "deleted_" + anonymizedID

	// Revoke invitations involving this account and remove the invited email,
	// which is personal data even when invited_user_id was not resolved.
	if _, err = tx.Exec(ctx, `
		UPDATE library_invitations
		SET status = CASE WHEN status='pending' THEN 'revoked' ELSE status END,
		    invited_user_id = CASE WHEN invited_user_id=$1 THEN NULL ELSE invited_user_id END,
		    invited_email = CASE
		        WHEN invited_user_id=$1 OR LOWER(invited_email)=LOWER($2) THEN $3
		        ELSE invited_email
		    END
		WHERE invited_by=$1
		   OR invited_user_id=$1
		   OR LOWER(invited_email)=LOWER($2)`, userID, currentEmail, anonymizedEmail); err != nil {
		return fmt.Errorf("revoke invitations: %w", err)
	}

	// Libraries cannot retain a deleted account as an active owner.
	if _, err = tx.Exec(ctx, `
		UPDATE libraries
		SET deleted_at=NOW(), updated_at=NOW()
		WHERE owner_id=$1 AND deleted_at IS NULL`, userID); err != nil {
		return fmt.Errorf("soft delete owned libraries: %w", err)
	}

	// Remove shared inventory links before making the physical copies private
	// and inaccessible. Personal copy state is erased rather than retained.
	if _, err = tx.Exec(ctx, `
		DELETE FROM library_book_copies lbc
		USING book_copies bc
		WHERE lbc.book_copy_id=bc.id AND bc.owner_id=$1`, userID); err != nil {
		return fmt.Errorf("remove library copy links: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		DELETE FROM shelf_books sb
		USING book_copies bc
		WHERE sb.copy_id=bc.id AND bc.owner_id=$1`, userID); err != nil {
		return fmt.Errorf("remove shelf copy links: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		DELETE FROM collection_books cb
		USING book_copies bc
		WHERE cb.book_copy_id=bc.id AND bc.owner_id=$1`, userID); err != nil {
		return fmt.Errorf("remove collection copy links: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM collection_books WHERE added_by=$1`, userID); err != nil {
		return fmt.Errorf("remove contributed collection links: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		UPDATE collections SET deleted_at=NOW(), updated_at=NOW()
		WHERE created_by=$1 AND deleted_at IS NULL`, userID); err != nil {
		return fmt.Errorf("soft delete created collections: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM shelves WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("delete shelves: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		DELETE FROM user_copy_states WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("delete personal copy states: %w", err)
	}

	if _, err = tx.Exec(ctx, `
		UPDATE book_copies
		SET deleted_at=NOW(), personal_notes=NULL, location=NULL, borrowed_from=NULL, updated_at=NOW()
		WHERE owner_id=$1 AND deleted_at IS NULL`, userID); err != nil {
		return fmt.Errorf("soft delete copies: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		UPDATE book_copies SET borrowed_from=NULL, updated_at=NOW()
		WHERE borrowed_from=$1`, userID); err != nil {
		return fmt.Errorf("anonymise borrowed copies: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM library_members WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("remove memberships: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE reviews SET user_id=NULL WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("anonymise reviews: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM notifications WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("delete notifications: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM email_queue WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("delete queued email: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM reading_sessions WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("delete reading sessions: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM reading_challenges WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("delete reading challenges: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM import_jobs WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("delete import jobs: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM review_likes WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("delete review likes: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM user_follows WHERE follower_id=$1 OR following_id=$1`, userID); err != nil {
		return fmt.Errorf("delete user follows: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE quotes SET user_id=NULL WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("anonymise quotes: %w", err)
	}

	if _, err = tx.Exec(ctx, `
		UPDATE users
		SET email=$2,
		    username=$3,
		    password_hash='deleted-account',
		    role='user',
		    is_admin=false,
		    theme='default-light',
		    bio=NULL,
		    avatar_url=NULL,
		    deleted_at=NOW(),
		    updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`, userID, anonymizedEmail, anonymizedUsername); err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit account deletion: %w", err)
	}
	return nil
}
