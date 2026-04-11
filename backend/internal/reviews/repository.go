package reviews

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

func (r *Repository) FindByBook(ctx context.Context, bookID string, page, limit int) ([]Review, int, error) {
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM reviews
		WHERE book_id = $1 AND deleted_at IS NULL AND is_public = TRUE
	`, bookID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count reviews: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT rv.id, rv.book_id, rv.user_id, u.username, rv.rating, rv.body, rv.is_public,
		       rv.deleted_at, rv.created_at, rv.updated_at
		FROM reviews rv
		LEFT JOIN users u ON u.id = rv.user_id
		WHERE rv.book_id = $1 AND rv.deleted_at IS NULL AND rv.is_public = TRUE
		ORDER BY rv.created_at DESC
		LIMIT $2 OFFSET $3
	`, bookID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	var list []Review
	for rows.Next() {
		var rv Review
		if err := rows.Scan(
			&rv.ID, &rv.BookID, &rv.UserID, &rv.Username, &rv.Rating,
			&rv.Body, &rv.IsPublic, &rv.DeletedAt, &rv.CreatedAt, &rv.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan review: %w", err)
		}
		list = append(list, rv)
	}
	if list == nil {
		list = []Review{}
	}
	return list, total, nil
}

func (r *Repository) FindByUser(ctx context.Context, bookID, userID string) (*Review, error) {
	var rv Review
	err := r.db.QueryRow(ctx, `
		SELECT id, book_id, user_id, rating, body, is_public, deleted_at, created_at, updated_at
		FROM reviews
		WHERE book_id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, bookID, userID).Scan(
		&rv.ID, &rv.BookID, &rv.UserID, &rv.Rating,
		&rv.Body, &rv.IsPublic, &rv.DeletedAt, &rv.CreatedAt, &rv.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find review: %w", err)
	}
	return &rv, nil
}

func (r *Repository) Upsert(ctx context.Context, input UpsertReviewInput) (Review, error) {
	isPublic := true
	if input.IsPublic != nil {
		isPublic = *input.IsPublic
	}

	var rv Review
	err := r.db.QueryRow(ctx, `
		INSERT INTO reviews (book_id, user_id, rating, body, is_public)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (book_id, user_id) DO UPDATE
		SET rating = EXCLUDED.rating,
		    body = EXCLUDED.body,
		    is_public = EXCLUDED.is_public,
		    updated_at = NOW()
		RETURNING id, book_id, user_id, rating, body, is_public, deleted_at, created_at, updated_at
	`, input.BookID, input.UserID, input.Rating, input.Body, isPublic).Scan(
		&rv.ID, &rv.BookID, &rv.UserID, &rv.Rating,
		&rv.Body, &rv.IsPublic, &rv.DeletedAt, &rv.CreatedAt, &rv.UpdatedAt,
	)
	if err != nil {
		return Review{}, fmt.Errorf("failed to upsert review: %w", err)
	}
	return rv, nil
}

func (r *Repository) Delete(ctx context.Context, bookID, userID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE reviews SET deleted_at = NOW(), updated_at = NOW()
		WHERE book_id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, bookID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}
	return nil
}
