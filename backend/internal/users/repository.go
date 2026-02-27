package users

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Insert(ctx context.Context, email string, username string, passwordHash string) (User, error) {
	var user User
	var query string = `
        INSERT INTO users (email, username, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, email, username, password_hash, role, is_admin, theme, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, email, username, passwordHash).Scan(
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
	var query string = `SELECT COUNT(*) FROM users WHERE email = $1 AND deleted_at IS NULL`

	var err error = r.db.QueryRow(ctx, query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	var query string = `
        SELECT id, email, username, password_hash, role, is_admin, theme, deleted_at, created_at, updated_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `

	var err error = r.db.QueryRow(ctx, query, email).Scan(
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
	var validThemes map[string]bool = map[string]bool{
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

	var query string = `
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
	var query string = `
		SELECT id, email, username, password_hash, role, is_admin, theme, deleted_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	var err error = r.db.QueryRow(ctx, query, id).Scan(
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
