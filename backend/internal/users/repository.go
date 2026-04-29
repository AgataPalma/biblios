package users

import (
	"context"
	"fmt"

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

func (r *Repository) CreateUser(ctx context.Context, email, username, passwordHash string) (User, error) {
	var u User
	err := scanUser(r.db.QueryRow(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING `+userColumns, email, username, passwordHash), &u)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email=$1 AND deleted_at IS NULL`, email).Scan(&n)
	return n > 0, err
}

func (r *Repository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE username=$1 AND deleted_at IS NULL`, username).Scan(&n)
	return n > 0, err
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email=$1 AND deleted_at IS NULL`, email), &u)
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
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Soft-delete copies
	if _, err = tx.Exec(ctx, `UPDATE book_copies SET deleted_at=NOW() WHERE owner_id=$1 AND deleted_at IS NULL`, userID); err != nil {
		return fmt.Errorf("soft delete copies: %w", err)
	}
	// Remove library memberships
	if _, err = tx.Exec(ctx, `DELETE FROM library_members WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("remove memberships: %w", err)
	}
	// Anonymise reviews
	if _, err = tx.Exec(ctx, `UPDATE reviews SET user_id=NULL WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("anonymise reviews: %w", err)
	}
	// Soft-delete user
	if _, err = tx.Exec(ctx, `UPDATE users SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, userID); err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	return tx.Commit(ctx)
}

// CreateDefaultLibrary creates the default private library for a new user.
func (r *Repository) CreateDefaultLibrary(ctx context.Context, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var libID string
	err = tx.QueryRow(ctx, `
		INSERT INTO libraries (owner_id, name, description, visibility, is_cooperative)
		VALUES ($1, 'My Library', 'My personal library', 'private', false)
		RETURNING id`, userID).Scan(&libID)
	if err != nil {
		return fmt.Errorf("create default library: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO library_members (library_id, user_id, is_owner, can_view, can_add, can_remove, can_edit, can_invite, can_manage_members)
		VALUES ($1, $2, true, true, true, true, true, true, true)`, libID, userID)
	if err != nil {
		return fmt.Errorf("add owner to library: %w", err)
	}

	return tx.Commit(ctx)
}
