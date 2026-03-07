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

func (r *Repository) CreateUser(ctx context.Context, email string, username string, passwordHash string) (User, error) {
	var user User
	var query = `
        INSERT INTO users (email, username, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, email, username, password_hash, role, is_admin, theme, deleted_at, created_at, updated_at
    `

	var err = r.db.QueryRow(ctx, query, email, username, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.IsAdmin,
		&user.Theme,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("failed to insert user: %w", err)
	}

	return user, nil
}

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	var query = `SELECT COUNT(*) FROM users WHERE email = $1 AND deleted_at IS NULL`

	var err = r.db.QueryRow(ctx, query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

func (r *Repository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int
	var query = `SELECT COUNT(*) FROM users WHERE username = $1 AND deleted_at IS NULL`

	var err = r.db.QueryRow(ctx, query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	var query = `
        SELECT id, email, username, password_hash, role, is_admin, theme, deleted_at, created_at, updated_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `

	var err = r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.IsAdmin,
		&user.Theme,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
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

	var query = `
        UPDATE users SET theme = $2, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	var _, err = r.db.Exec(ctx, query, userID, theme)
	if err != nil {
		return fmt.Errorf("failed to update theme: %w", err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (User, error) {
	var user User
	var query = `
		SELECT id, email, username, password_hash, role, is_admin, theme, deleted_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	var err = r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.IsAdmin,
		&user.Theme,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, userID string, email *string, username *string) (User, error) {
	var user User
	var query = `
        UPDATE users
        SET
            email    = COALESCE($2, email),
            username = COALESCE($3, username),
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
        RETURNING id, email, username, password_hash, role, is_admin, theme, deleted_at, created_at, updated_at
    `
	// passing nil for a parameter makes pgx send NULL, which COALESCE ignores
	var err = r.db.QueryRow(ctx, query, userID, email, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.IsAdmin,
		&user.Theme,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("failed to update profile: %w", err)
	}
	return user, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID string, newPasswordHash string) error {
	var query = `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	var _, err = r.db.Exec(ctx, query, userID, newPasswordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
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
		UPDATE book_copies
		SET deleted_at = NOW()
		WHERE owner_id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to soft delete book copies: %w", err)
	}

	// 2. Remove from library_members (hard delete — membership rows have no deleted_at)
	_, err = tx.Exec(ctx, `
		DELETE FROM library_members WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to remove library memberships: %w", err)
	}

	// 3. Remove library_book_copies for this user's copies
	// (ON DELETE CASCADE from book_copies handles this automatically once copies are soft-deleted,
	//  but since we're using soft deletes we need to clean up explicitly)
	_, err = tx.Exec(ctx, `
		DELETE FROM library_book_copies
		WHERE book_copy_id IN (
			SELECT id FROM book_copies WHERE owner_id = $1
		)
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to remove library book copies: %w", err)
	}

	// 4. Anonymise reviews — keep the rating data, remove the link to the user
	_, err = tx.Exec(ctx, `
		UPDATE reviews
		SET user_id = NULL, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to anonymise reviews: %w", err)
	}

	// 5. Soft delete the user account
	_, err = tx.Exec(ctx, `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit user deletion: %w", err)
	}

	return nil
}
