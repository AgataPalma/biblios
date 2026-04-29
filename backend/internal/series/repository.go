package series

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

const seriesColumns = `id, name, description, status, rejection_reason, deleted_at, created_at, updated_at`

func scanSeries(row pgx.Row, s *Series) error {
	return row.Scan(&s.ID, &s.Name, &s.Description, &s.Status, &s.RejectionReason,
		&s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
}

func (r *Repository) Create(ctx context.Context, name string, description *string, autoApprove bool) (Series, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var s Series
	err := scanSeries(r.db.QueryRow(ctx, `
		INSERT INTO series (name, description, status) VALUES ($1, $2, $3)
		RETURNING `+seriesColumns, name, description, status), &s)
	if err != nil {
		return Series{}, fmt.Errorf("create series: %w", err)
	}
	return s, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Series, error) {
	var s Series
	err := scanSeries(r.db.QueryRow(ctx, `
		SELECT `+seriesColumns+` FROM series WHERE id=$1 AND deleted_at IS NULL`, id), &s)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find series: %w", err)
	}
	return &s, nil
}

func (r *Repository) Search(ctx context.Context, query string, page, limit int) ([]Series, int, error) {
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
		where += " AND lower(name) LIKE lower('%'||$1||'%')"
		args = append(args, query)
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM series WHERE `+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count series: %w", err)
	}

	argN := len(args) + 1
	queryArgs := make([]any, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT `+seriesColumns+` FROM series WHERE %s
		ORDER BY name ASC LIMIT $%d OFFSET $%d`, where, argN, argN+1), queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("search series: %w", err)
	}
	defer rows.Close()

	var results []Series
	for rows.Next() {
		var s Series
		if err := scanSeries(rows, &s); err != nil {
			return nil, 0, err
		}
		results = append(results, s)
	}
	return results, total, rows.Err()
}

// GetBooks returns all approved books in a series ordered by series_position.
func (r *Repository) GetBooks(ctx context.Context, seriesID string) ([]SeriesBook, error) {
	rows, err := r.db.Query(ctx, `
		SELECT b.id, b.title, b.series_position,
		       (SELECT be.cover_url FROM book_editions be
		        WHERE be.book_id=b.id AND be.deleted_at IS NULL
		        ORDER BY be.published_at DESC NULLS LAST LIMIT 1) AS cover_url
		FROM books b
		WHERE b.series_id=$1 AND b.deleted_at IS NULL AND b.status='approved'
		ORDER BY b.series_position ASC NULLS LAST, b.title ASC`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("get series books: %w", err)
	}
	defer rows.Close()

	var seriesBooks []SeriesBook
	for rows.Next() {
		var sb SeriesBook
		if err := rows.Scan(&sb.BookID, &sb.Title, &sb.SeriesPosition, &sb.CoverURL); err != nil {
			return nil, err
		}
		seriesBooks = append(seriesBooks, sb)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Batch-load authors for each book
	if len(seriesBooks) > 0 {
		ids := make([]string, len(seriesBooks))
		idx := map[string]int{}
		for i, sb := range seriesBooks {
			ids[i] = sb.BookID
			idx[sb.BookID] = i
		}
		aRows, err := r.db.Query(ctx, `
			SELECT bc.book_id, c.name
			FROM contributors c JOIN book_contributors bc ON bc.contributor_id=c.id
			WHERE bc.book_id=ANY($1) AND bc.role IN ('author','co_author') AND c.deleted_at IS NULL
			ORDER BY c.name`, ids)
		if err == nil {
			for aRows.Next() {
				var bookID, name string
				if err := aRows.Scan(&bookID, &name); err == nil {
					if i, ok := idx[bookID]; ok {
						seriesBooks[i].Authors = append(seriesBooks[i].Authors, name)
					}
				}
			}
			aRows.Close()
		}
	}

	return seriesBooks, nil
}
