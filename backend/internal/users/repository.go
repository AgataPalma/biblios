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

// scanUser scans all user columns in the canonical column order used by every query.
// Column order: id, email, username, password_hash, role, is_admin, theme, bio, avatar_url, deleted_at, created_at, updated_at
func scanUser(row pgx.Row, user *User) error {
	return row.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.IsAdmin,
		&user.Theme,
		&user.Bio,
		&user.AvatarUrl,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}

const userColumns = `id, email, username, password_hash, role, is_admin, theme, bio, avatar_url, deleted_at, created_at, updated_at`

func (r *Repository) CreateUser(ctx context.Context, email string, username string, passwordHash string) (User, error) {
	var user User
	var query = `
        INSERT INTO users (email, username, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, email, username, password_hash, role, is_admin, theme, bio, avatar_url, deleted_at, created_at, updated_at`

	var err = scanUser(r.db.QueryRow(ctx, query, email, username, passwordHash), &user)
	if err != nil {
		return User{}, fmt.Errorf("failed to insert user: %w", err)
	}

	return user, nil
}

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	var err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email = $1 AND deleted_at IS NULL`, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int
	var err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE username = $1 AND deleted_at IS NULL`, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	var query = `SELECT ` + userColumns + ` FROM users WHERE email = $1 AND deleted_at IS NULL`
	var err = scanUser(r.db.QueryRow(ctx, query, email), &user)
	if err != nil {
		return User{}, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (User, error) {
	var user User
	var query = `SELECT ` + userColumns + ` FROM users WHERE id = $1 AND deleted_at IS NULL`
	var err = scanUser(r.db.QueryRow(ctx, query, id), &user)
	if err != nil {
		return User{}, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// UpdateUser updates any combination of email, username, bio, and avatar_url.
// Passing nil for a field leaves it unchanged (COALESCE pattern).
func (r *Repository) UpdateUser(ctx context.Context, userID string, email, username, bio, avatarUrl *string) (User, error) {
	var user User
	var query = `
        UPDATE users
        SET
            email      = COALESCE($2, email),
            username   = COALESCE($3, username),
            bio        = COALESCE($4, bio),
            avatar_url = COALESCE($5, avatar_url),
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
        RETURNING id, email, username, password_hash, role, is_admin, theme, bio, avatar_url, deleted_at, created_at, updated_at`

	var err = scanUser(r.db.QueryRow(ctx, query, userID, email, username, bio, avatarUrl), &user)
	if err != nil {
		return User{}, fmt.Errorf("failed to update profile: %w", err)
	}
	return user, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID string, newPasswordHash string) error {
	var _, err = r.db.Exec(ctx, `
		UPDATE users SET password_hash = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, newPasswordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *Repository) UpdateTheme(ctx context.Context, userID string, theme string) error {
	var validThemes = map[string]bool{
		"default-light":    true,
		"woody":            true,
		"nordic":           true,
		"metallic":         true,
		"futuristic":       true,
		"post-apocalyptic": true,
		"dark-academia":    true,
		"ocean":            true,
		"space":            true,
	}

	if !validThemes[theme] {
		return fmt.Errorf("invalid theme: %s", theme)
	}

	var _, err = r.db.Exec(ctx, `
        UPDATE users SET theme = $2, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `, userID, theme)
	if err != nil {
		return fmt.Errorf("failed to update theme: %w", err)
	}
	return nil
}

// SoftDeleteWithCascade soft-deletes the user and all their personal data in a single transaction.
// Books in the catalogue are preserved. Reviews are anonymised (user_id set to NULL).
func (r *Repository) SoftDeleteWithCascade(ctx context.Context, userID string) error {
	var tx pgx.Tx
	var err error

	tx, err = r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// 1. Soft delete the user's book copies
	_, err = tx.Exec(ctx, `
		UPDATE book_copies SET deleted_at = NOW()
		WHERE owner_id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to soft delete book copies: %w", err)
	}

	// 2. Remove from library_members (hard delete — membership rows have no deleted_at)
	_, err = tx.Exec(ctx, `DELETE FROM library_members WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to remove library memberships: %w", err)
	}

	// 3. Soft delete the user
	_, err = tx.Exec(ctx, `
		UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	return tx.Commit(ctx)
}
