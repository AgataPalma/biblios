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
        RETURNING id, email, username, password_hash, is_admin, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, email, username, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.IsAdmin,
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
        SELECT id, email, username, password_hash, is_admin, deleted_at, created_at, updated_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `

	var err error = r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
