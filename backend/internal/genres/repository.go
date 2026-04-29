package genres

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

// ─── Genres ───────────────────────────────────────────────────────────────────

func (r *Repository) ListGenres(ctx context.Context) ([]Genre, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, status, rejection_reason, created_at
		FROM genres WHERE status='approved'
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list genres: %w", err)
	}
	defer rows.Close()

	var genres []Genre
	for rows.Next() {
		var g Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.Status, &g.RejectionReason, &g.CreatedAt); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, rows.Err()
}

func (r *Repository) CreateGenre(ctx context.Context, name string, autoApprove bool) (Genre, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var g Genre
	err := r.db.QueryRow(ctx, `
		INSERT INTO genres (name, status) VALUES ($1, $2)
		ON CONFLICT (name) DO NOTHING
		RETURNING id, name, status, rejection_reason, created_at`,
		name, status).Scan(&g.ID, &g.Name, &g.Status, &g.RejectionReason, &g.CreatedAt)
	if err != nil {
		// Already exists — return existing
		err2 := r.db.QueryRow(ctx, `
			SELECT id, name, status, rejection_reason, created_at FROM genres WHERE name=$1`, name).Scan(
			&g.ID, &g.Name, &g.Status, &g.RejectionReason, &g.CreatedAt)
		if err2 != nil {
			return Genre{}, fmt.Errorf("create genre: %w", err)
		}
	}
	return g, nil
}

// ─── Moods ────────────────────────────────────────────────────────────────────

func (r *Repository) ListMoods(ctx context.Context) ([]Mood, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, status, rejection_reason
		FROM moods WHERE status='approved'
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list moods: %w", err)
	}
	defer rows.Close()

	var moods []Mood
	for rows.Next() {
		var m Mood
		if err := rows.Scan(&m.ID, &m.Name, &m.Status, &m.RejectionReason); err != nil {
			return nil, err
		}
		moods = append(moods, m)
	}
	return moods, rows.Err()
}

func (r *Repository) CreateMood(ctx context.Context, name string, autoApprove bool) (Mood, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var m Mood
	err := r.db.QueryRow(ctx, `
		INSERT INTO moods (name, status) VALUES ($1, $2)
		ON CONFLICT (name) DO NOTHING
		RETURNING id, name, status, rejection_reason`,
		name, status).Scan(&m.ID, &m.Name, &m.Status, &m.RejectionReason)
	if err != nil {
		err2 := r.db.QueryRow(ctx, `
			SELECT id, name, status, rejection_reason FROM moods WHERE name=$1`, name).Scan(
			&m.ID, &m.Name, &m.Status, &m.RejectionReason)
		if err2 != nil {
			return Mood{}, fmt.Errorf("create mood: %w", err)
		}
	}
	return m, nil
}
