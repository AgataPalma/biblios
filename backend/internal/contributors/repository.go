package contributors

import (
	"context"
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

const contributorColumns = `id, name, bio, born_date, died_date, photo_url, website,
	nationality, status, rejection_reason, deleted_at, created_at, updated_at`

func scanContributor(row pgx.Row, c *Contributor) error {
	return row.Scan(
		&c.ID, &c.Name, &c.Bio, &c.BornDate, &c.DiedDate, &c.PhotoURL, &c.Website,
		&c.Nationality, &c.Status, &c.RejectionReason, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt,
	)
}

func (r *Repository) Create(ctx context.Context, name string, bio, photoURL, website, nationality *string, autoApprove bool) (Contributor, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var c Contributor
	err := scanContributor(r.db.QueryRow(ctx, `
		INSERT INTO contributors (name, bio, photo_url, website, nationality, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+contributorColumns,
		name, bio, photoURL, website, nationality, status), &c)
	if err != nil {
		return Contributor{}, fmt.Errorf("create contributor: %w", err)
	}
	return c, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Contributor, error) {
	var c Contributor
	err := scanContributor(r.db.QueryRow(ctx, `
		SELECT `+contributorColumns+` FROM contributors WHERE id=$1 AND deleted_at IS NULL`, id), &c)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find contributor: %w", err)
	}
	return &c, nil
}

func (r *Repository) Search(ctx context.Context, query string, page, limit int) ([]Contributor, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	args := []any{}
	where := "deleted_at IS NULL AND status='approved'"
	if strings.TrimSpace(query) != "" {
		where += " AND (to_tsvector('english', name) @@ plainto_tsquery('english', $1) OR lower(name) LIKE lower('%'||$1||'%'))"
		args = append(args, query)
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contributors WHERE `+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count contributors: %w", err)
	}

	argN := len(args) + 1
	queryArgs := make([]any, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT `+contributorColumns+` FROM contributors
		WHERE %s ORDER BY name ASC
		LIMIT $%d OFFSET $%d`, where, argN, argN+1), queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("search contributors: %w", err)
	}
	defer rows.Close()

	var contribs []Contributor
	for rows.Next() {
		var c Contributor
		if err := scanContributor(rows, &c); err != nil {
			return nil, 0, err
		}
		contribs = append(contribs, c)
	}
	return contribs, total, rows.Err()
}

func (r *Repository) GetAwards(ctx context.Context, contributorID string) ([]ContributorAward, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.name, ca.year, ca.category, ca.result
		FROM contributor_awards ca
		JOIN awards a ON a.id=ca.award_id
		WHERE ca.contributor_id=$1
		ORDER BY ca.year DESC`, contributorID)
	if err != nil {
		return nil, fmt.Errorf("get contributor awards: %w", err)
	}
	defer rows.Close()

	var awards []ContributorAward
	for rows.Next() {
		var a ContributorAward
		if err := rows.Scan(&a.AwardID, &a.AwardName, &a.Year, &a.Category, &a.Result); err != nil {
			return nil, err
		}
		awards = append(awards, a)
	}
	return awards, rows.Err()
}
