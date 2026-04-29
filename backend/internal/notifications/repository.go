package notifications

import (
	"context"
	"encoding/json"
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

func scanNotification(row pgx.Row, n *Notification) error {
	return row.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.ReadAt, &n.CreatedAt)
}

// Create inserts a new notification. data may be nil.
func (r *Repository) Create(ctx context.Context, userID, notifType, title, body string, data json.RawMessage) (Notification, error) {
	var n Notification
	err := scanNotification(r.db.QueryRow(ctx, `
		INSERT INTO notifications (user_id, type, title, body, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, type, title, body, data, read_at, created_at`,
		userID, notifType, title, body, data), &n)
	if err != nil {
		return Notification{}, fmt.Errorf("create notification: %w", err)
	}
	return n, nil
}

// List returns notifications for a user with optional type and read-status filters.
func (r *Repository) List(ctx context.Context, userID, filterType, filterRead string, page, limit int) ([]Notification, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	args := []any{userID}
	argN := 2
	where := []string{"user_id=$1"}

	if filterType != "" {
		where = append(where, fmt.Sprintf("type=$%d", argN))
		args = append(args, filterType)
		argN++
	}
	switch strings.ToLower(filterRead) {
	case "true", "1", "yes":
		where = append(where, "read_at IS NOT NULL")
	case "false", "0", "no":
		where = append(where, "read_at IS NULL")
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE `+whereClause, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	queryArgs := make([]any, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, user_id, type, title, body, data, read_at, created_at
		FROM notifications
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argN, argN+1), queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifs []Notification
	for rows.Next() {
		var n Notification
		if err := scanNotification(rows, &n); err != nil {
			return nil, 0, err
		}
		notifs = append(notifs, n)
	}
	return notifs, total, rows.Err()
}

// MarkRead marks a single notification as read, only if it belongs to the user.
func (r *Repository) MarkRead(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE notifications SET read_at=NOW()
		WHERE id=$1 AND user_id=$2 AND read_at IS NULL`, id, userID)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("notification not found or already read")
	}
	return nil
}

// MarkAllRead marks all unread notifications for a user as read.
func (r *Repository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE notifications SET read_at=NOW()
		WHERE user_id=$1 AND read_at IS NULL`, userID)
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}
