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

// Upsert creates or replaces the caller's review for a book (one per user per book).
func (r *Repository) Upsert(ctx context.Context, bookID, userID string, rating float64, body *string, isPublic bool) (Review, error) {
	var rev Review
	err := r.db.QueryRow(ctx, `
		INSERT INTO reviews (book_id, user_id, rating, body, is_public)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (book_id, user_id) DO UPDATE SET
			rating     = EXCLUDED.rating,
			body       = EXCLUDED.body,
			is_public  = EXCLUDED.is_public,
			updated_at = NOW()
		RETURNING id, book_id, user_id, rating, body, is_public, like_count, deleted_at, created_at, updated_at`,
		bookID, userID, rating, body, isPublic,
	).Scan(
		&rev.ID, &rev.BookID, &rev.UserID, &rev.Rating, &rev.Body,
		&rev.IsPublic, &rev.LikeCount, &rev.DeletedAt, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err != nil {
		return Review{}, fmt.Errorf("upsert review: %w", err)
	}
	return rev, nil
}

func (r *Repository) FindByBookAndUser(ctx context.Context, bookID, userID string) (*Review, error) {
	var rev Review
	err := r.db.QueryRow(ctx, `
		SELECT r.id, r.book_id, r.user_id, r.rating, r.body, r.is_public, r.like_count,
		       r.deleted_at, r.created_at, r.updated_at,
		       u.username, u.avatar_url
		FROM reviews r
		LEFT JOIN users u ON u.id=r.user_id
		WHERE r.book_id=$1 AND r.user_id=$2 AND r.deleted_at IS NULL`,
		bookID, userID,
	).Scan(
		&rev.ID, &rev.BookID, &rev.UserID, &rev.Rating, &rev.Body, &rev.IsPublic, &rev.LikeCount,
		&rev.DeletedAt, &rev.CreatedAt, &rev.UpdatedAt,
		&rev.Username, &rev.AvatarURL,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find review: %w", err)
	}
	return &rev, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Review, error) {
	var rev Review
	err := r.db.QueryRow(ctx, `
		SELECT r.id, r.book_id, r.user_id, r.rating, r.body, r.is_public, r.like_count,
		       r.deleted_at, r.created_at, r.updated_at,
		       u.username, u.avatar_url
		FROM reviews r
		LEFT JOIN users u ON u.id=r.user_id
		WHERE r.id=$1 AND r.deleted_at IS NULL`, id,
	).Scan(
		&rev.ID, &rev.BookID, &rev.UserID, &rev.Rating, &rev.Body, &rev.IsPublic, &rev.LikeCount,
		&rev.DeletedAt, &rev.CreatedAt, &rev.UpdatedAt,
		&rev.Username, &rev.AvatarURL,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find review by id: %w", err)
	}
	return &rev, nil
}

// ListPublic returns all public reviews for a book, with a flag indicating if the caller liked each one.
func (r *Repository) ListPublic(ctx context.Context, bookID, callerID string, page, limit int) ([]Review, int, *float64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM reviews
		WHERE book_id=$1 AND is_public=true AND deleted_at IS NULL`, bookID).Scan(&total); err != nil {
		return nil, 0, nil, fmt.Errorf("count reviews: %w", err)
	}

	var avg *float64
	r.db.QueryRow(ctx, `
		SELECT AVG(rating) FROM reviews
		WHERE book_id=$1 AND is_public=true AND deleted_at IS NULL`, bookID).Scan(&avg)

	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.book_id, r.user_id, r.rating, r.body, r.is_public, r.like_count,
		       r.deleted_at, r.created_at, r.updated_at,
		       u.username, u.avatar_url,
		       EXISTS(SELECT 1 FROM review_likes rl WHERE rl.review_id=r.id AND rl.user_id=$2) AS liked_by_me
		FROM reviews r
		LEFT JOIN users u ON u.id=r.user_id
		WHERE r.book_id=$1 AND r.is_public=true AND r.deleted_at IS NULL
		ORDER BY r.like_count DESC, r.created_at DESC
		LIMIT $3 OFFSET $4`,
		bookID, callerID, limit, offset)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var revs []Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(
			&rev.ID, &rev.BookID, &rev.UserID, &rev.Rating, &rev.Body, &rev.IsPublic, &rev.LikeCount,
			&rev.DeletedAt, &rev.CreatedAt, &rev.UpdatedAt,
			&rev.Username, &rev.AvatarURL, &rev.LikedByMe,
		); err != nil {
			return nil, 0, nil, err
		}
		revs = append(revs, rev)
	}
	return revs, total, avg, rows.Err()
}

func (r *Repository) SoftDelete(ctx context.Context, bookID, userID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE reviews SET deleted_at=NOW() WHERE book_id=$1 AND user_id=$2 AND deleted_at IS NULL`,
		bookID, userID)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}
	return nil
}

// AddLike records a like. Returns error if user tries to like their own review or likes twice.
func (r *Repository) AddLike(ctx context.Context, reviewID, userID string) error {
	rev, err := r.FindByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if rev == nil {
		return fmt.Errorf("review not found")
	}
	if rev.UserID != nil && *rev.UserID == userID {
		return fmt.Errorf("cannot like your own review")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	tag, err := tx.Exec(ctx, `
		INSERT INTO review_likes (review_id, user_id) VALUES ($1,$2)
		ON CONFLICT DO NOTHING`, reviewID, userID)
	if err != nil {
		return fmt.Errorf("insert like: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("already liked this review")
	}

	_, err = tx.Exec(ctx, `UPDATE reviews SET like_count=like_count+1 WHERE id=$1`, reviewID)
	if err != nil {
		return fmt.Errorf("increment like count: %w", err)
	}

	return tx.Commit(ctx)
}

// RemoveLike removes a like.
func (r *Repository) RemoveLike(ctx context.Context, reviewID, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	tag, err := tx.Exec(ctx, `DELETE FROM review_likes WHERE review_id=$1 AND user_id=$2`, reviewID, userID)
	if err != nil {
		return fmt.Errorf("remove like: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("like not found")
	}

	_, err = tx.Exec(ctx, `UPDATE reviews SET like_count=GREATEST(0, like_count-1) WHERE id=$1`, reviewID)
	if err != nil {
		return fmt.Errorf("decrement like count: %w", err)
	}

	return tx.Commit(ctx)
}
