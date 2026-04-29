package books

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) withDB(db DB) *txRepository {
	return &txRepository{db: db}
}

type txRepository struct {
	db DB
}

// FindEditionByISBN searches by isbn10 or isbn13
func (r *Repository) FindEditionByISBN(ctx context.Context, isbn string) (*Edition, error) {
	var e Edition
	err := r.db.QueryRow(ctx, `
		SELECT id, book_id, title, original_title, format, description, cover_url,
		       isbn10, isbn13, asin, language, publisher, edition, published_at,
		       page_count, file_format, duration_minutes, audio_format, status,
		       rejection_reason, deleted_at, created_at, updated_at
		FROM book_editions
		WHERE (isbn10=$1 OR isbn13=$1) AND deleted_at IS NULL
		LIMIT 1`, isbn).Scan(
		&e.ID, &e.BookID, &e.Title, &e.OriginalTitle, &e.Format, &e.Description, &e.CoverURL,
		&e.ISBN10, &e.ISBN13, &e.ASIN, &e.Language, &e.Publisher, &e.Edition, &e.PublishedAt,
		&e.PageCount, &e.FileFormat, &e.DurationMinutes, &e.AudioFormat, &e.Status,
		&e.RejectionReason, &e.DeletedAt, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find edition by isbn: %w", err)
	}
	return &e, nil
}

func (r *Repository) FindEditionByID(ctx context.Context, id string) (*Edition, error) {
	var e Edition
	err := r.db.QueryRow(ctx, `
		SELECT id, book_id, title, original_title, format, description, cover_url,
		       isbn10, isbn13, asin, language, publisher, edition, published_at,
		       page_count, file_format, duration_minutes, audio_format, status,
		       rejection_reason, deleted_at, created_at, updated_at
		FROM book_editions
		WHERE id=$1 AND deleted_at IS NULL`, id).Scan(
		&e.ID, &e.BookID, &e.Title, &e.OriginalTitle, &e.Format, &e.Description, &e.CoverURL,
		&e.ISBN10, &e.ISBN13, &e.ASIN, &e.Language, &e.Publisher, &e.Edition, &e.PublishedAt,
		&e.PageCount, &e.FileFormat, &e.DurationMinutes, &e.AudioFormat, &e.Status,
		&e.RejectionReason, &e.DeletedAt, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find edition by id: %w", err)
	}
	return &e, nil
}

func (r *Repository) FindBookByID(ctx context.Context, id string) (*Book, error) {
	var b Book
	err := r.db.QueryRow(ctx, `
		SELECT id, title, description, series_id, series_position, status,
		       rejection_reason, deleted_at, created_at, updated_at
		FROM books WHERE id=$1 AND deleted_at IS NULL`, id).Scan(
		&b.ID, &b.Title, &b.Description, &b.SeriesID, &b.SeriesPosition,
		&b.Status, &b.RejectionReason, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find book by id: %w", err)
	}
	return &b, nil
}

func (r *Repository) FindBookWithDetails(ctx context.Context, id string) (*Book, error) {
	b, err := r.FindBookByID(ctx, id)
	if err != nil || b == nil {
		return nil, err
	}

	// Load series
	if b.SeriesID != nil {
		var s Series
		serr := r.db.QueryRow(ctx, `
			SELECT id, name, description, status, rejection_reason, deleted_at, created_at, updated_at
			FROM series WHERE id=$1 AND deleted_at IS NULL`, *b.SeriesID).Scan(
			&s.ID, &s.Name, &s.Description, &s.Status, &s.RejectionReason,
			&s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if serr == nil {
			b.Series = &s
		}
	}

	// Load authors
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.name, c.bio, c.born_date, c.died_date, c.photo_url, c.website,
		       c.nationality, c.status, c.rejection_reason, c.deleted_at, c.created_at, c.updated_at
		FROM contributors c
		JOIN book_contributors bc ON bc.contributor_id=c.id
		WHERE bc.book_id=$1 AND bc.role IN ('author','co_author') AND c.deleted_at IS NULL
		ORDER BY c.name`, id)
	if err != nil {
		return nil, fmt.Errorf("load authors: %w", err)
	}
	defer rows.Close()
	b.Authors = []Contributor{}
	for rows.Next() {
		var c Contributor
		if err := rows.Scan(&c.ID, &c.Name, &c.Bio, &c.BornDate, &c.DiedDate, &c.PhotoURL,
			&c.Website, &c.Nationality, &c.Status, &c.RejectionReason,
			&c.DeletedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		b.Authors = append(b.Authors, c)
	}
	rows.Close()

	// Load genres
	gRows, err := r.db.Query(ctx, `
		SELECT g.id, g.name, g.status, g.rejection_reason, g.created_at
		FROM genres g JOIN book_genres bg ON bg.genre_id=g.id
		WHERE bg.book_id=$1 ORDER BY g.name`, id)
	if err != nil {
		return nil, fmt.Errorf("load genres: %w", err)
	}
	defer gRows.Close()
	b.Genres = []Genre{}
	for gRows.Next() {
		var g Genre
		if err := gRows.Scan(&g.ID, &g.Name, &g.Status, &g.RejectionReason, &g.CreatedAt); err != nil {
			return nil, err
		}
		b.Genres = append(b.Genres, g)
	}
	gRows.Close()

	// Load moods
	mRows, err := r.db.Query(ctx, `
		SELECT m.id, m.name, m.status, m.rejection_reason
		FROM moods m JOIN book_moods bm ON bm.mood_id=m.id
		WHERE bm.book_id=$1 ORDER BY m.name`, id)
	if err == nil {
		defer mRows.Close()
		b.Moods = []Mood{}
		for mRows.Next() {
			var m Mood
			if err := mRows.Scan(&m.ID, &m.Name, &m.Status, &m.RejectionReason); err == nil {
				b.Moods = append(b.Moods, m)
			}
		}
		mRows.Close()
	}

	// Load editions
	eRows, err := r.db.Query(ctx, `
		SELECT id, book_id, title, original_title, format, description, cover_url,
		       isbn10, isbn13, asin, language, publisher, edition, published_at,
		       page_count, file_format, duration_minutes, audio_format, status,
		       rejection_reason, deleted_at, created_at, updated_at
		FROM book_editions WHERE book_id=$1 AND deleted_at IS NULL ORDER BY published_at DESC`, id)
	if err != nil {
		return nil, fmt.Errorf("load editions: %w", err)
	}
	defer eRows.Close()
	b.Editions = []Edition{}
	for eRows.Next() {
		var e Edition
		if err := eRows.Scan(
			&e.ID, &e.BookID, &e.Title, &e.OriginalTitle, &e.Format, &e.Description, &e.CoverURL,
			&e.ISBN10, &e.ISBN13, &e.ASIN, &e.Language, &e.Publisher, &e.Edition, &e.PublishedAt,
			&e.PageCount, &e.FileFormat, &e.DurationMinutes, &e.AudioFormat, &e.Status,
			&e.RejectionReason, &e.DeletedAt, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		b.Editions = append(b.Editions, e)
	}
	eRows.Close()

	// Load edition contributors (narrators, translators) for each edition
	for i, ed := range b.Editions {
		ecRows, err := r.db.Query(ctx, `
			SELECT c.id, c.name, c.status, ec.role
			FROM contributors c JOIN edition_contributors ec ON ec.contributor_id=c.id
			WHERE ec.edition_id=$1 AND c.deleted_at IS NULL`, ed.ID)
		if err != nil {
			continue
		}
		for ecRows.Next() {
			var c Contributor
			var role string
			if err := ecRows.Scan(&c.ID, &c.Name, &c.Status, &role); err != nil {
				continue
			}
			switch role {
			case "narrator":
				b.Editions[i].Narrators = append(b.Editions[i].Narrators, c)
			case "translator":
				b.Editions[i].Translators = append(b.Editions[i].Translators, c)
			}
		}
		ecRows.Close()
	}

	// Load awards
	aRows, err := r.db.Query(ctx, `
		SELECT a.id, a.name, a.description, ba.year, ba.category, ba.result
		FROM awards a JOIN book_awards ba ON ba.award_id=a.id
		WHERE ba.book_id=$1 ORDER BY ba.year DESC`, id)
	if err == nil {
		defer aRows.Close()
		b.Awards = []BookAward{}
		for aRows.Next() {
			var ba BookAward
			if err := aRows.Scan(&ba.Award.ID, &ba.Award.Name, &ba.Award.Description,
				&ba.Year, &ba.Category, &ba.Result); err == nil {
				b.Awards = append(b.Awards, ba)
			}
		}
		aRows.Close()
	}

	// Load average rating
	var avg *float64
	r.db.QueryRow(ctx, `
		SELECT AVG(rating) FROM reviews
		WHERE book_id=$1 AND is_public=true AND deleted_at IS NULL`, id).Scan(&avg)
	b.AverageRating = avg

	return b, nil
}

func (r *Repository) FindBookByTitleAndAuthors(ctx context.Context, title string, authorNames []string) (*Book, error) {
	if len(authorNames) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, title, status, deleted_at, created_at, updated_at
		FROM books WHERE lower(title)=lower($1) AND deleted_at IS NULL`, title)
	if err != nil {
		return nil, fmt.Errorf("search books by title: %w", err)
	}
	defer rows.Close()

	var candidates []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Status, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		candidates = append(candidates, b)
	}
	rows.Close()

	for _, candidate := range candidates {
		aRows, err := r.db.Query(ctx, `
			SELECT c.name FROM contributors c
			JOIN book_contributors bc ON bc.contributor_id=c.id
			WHERE bc.book_id=$1 AND bc.role IN ('author','co_author') AND c.deleted_at IS NULL
			ORDER BY c.name`, candidate.ID)
		if err != nil {
			continue
		}
		var existing []string
		for aRows.Next() {
			var name string
			if err := aRows.Scan(&name); err == nil {
				existing = append(existing, strings.ToLower(strings.TrimSpace(name)))
			}
		}
		aRows.Close()

		input := make([]string, len(authorNames))
		for i, n := range authorNames {
			input[i] = strings.ToLower(strings.TrimSpace(n))
		}
		sort.Strings(input)
		sort.Strings(existing)

		if len(input) == len(existing) {
			match := true
			for i := range input {
				if input[i] != existing[i] {
					match = false
					break
				}
			}
			if match {
				return &candidate, nil
			}
		}
	}
	return nil, nil
}

func (r *Repository) SearchBooks(ctx context.Context, f SearchFilters) ([]Book, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	args := []any{}
	argN := 1
	where := []string{"b.deleted_at IS NULL", "b.status='approved'"}

	if f.Query != "" {
		where = append(where, fmt.Sprintf(`(
			b.search_vector @@ plainto_tsquery('english', $%d)
			OR EXISTS (
				SELECT 1 FROM book_editions be2
				WHERE be2.book_id=b.id AND be2.deleted_at IS NULL
				AND (
					(setweight(to_tsvector('portuguese', coalesce(be2.title,'')||' '||coalesce(be2.original_title,'')), 'A') ||
					 setweight(to_tsvector('english', coalesce(be2.title,'')||' '||coalesce(be2.original_title,'')), 'B'))
					@@ plainto_tsquery($%d)
				)
			)
		)`, argN, argN))
		args = append(args, f.Query)
		argN++
	}
	if f.Format != "" {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_editions be3 WHERE be3.book_id=b.id AND be3.format=$%d AND be3.deleted_at IS NULL)`, argN))
		args = append(args, f.Format)
		argN++
	}
	if f.Language != "" {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_editions be4 WHERE be4.book_id=b.id AND be4.language=$%d AND be4.deleted_at IS NULL)`, argN))
		args = append(args, f.Language)
		argN++
	}
	if f.Genre != "" {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_genres bg JOIN genres g ON g.id=bg.genre_id WHERE bg.book_id=b.id AND lower(g.name)=lower($%d))`, argN))
		args = append(args, f.Genre)
		argN++
	}
	if f.Mood != "" {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_moods bm JOIN moods m ON m.id=bm.mood_id WHERE bm.book_id=b.id AND lower(m.name)=lower($%d))`, argN))
		args = append(args, f.Mood)
		argN++
	}
	if f.Series != "" {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM series s WHERE s.id=b.series_id AND lower(s.name)=lower($%d))`, argN))
		args = append(args, f.Series)
		argN++
	}
	if f.Publisher != "" {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_editions be5 WHERE be5.book_id=b.id AND lower(be5.publisher)=lower($%d) AND be5.deleted_at IS NULL)`, argN))
		args = append(args, f.Publisher)
		argN++
	}
	if f.YearFrom > 0 {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_editions be6 WHERE be6.book_id=b.id AND EXTRACT(YEAR FROM be6.published_at)>=$%d AND be6.deleted_at IS NULL)`, argN))
		args = append(args, f.YearFrom)
		argN++
	}
	if f.YearTo > 0 {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_editions be7 WHERE be7.book_id=b.id AND EXTRACT(YEAR FROM be7.published_at)<=$%d AND be7.deleted_at IS NULL)`, argN))
		args = append(args, f.YearTo)
		argN++
	}
	if f.Award != "" {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_awards ba JOIN awards a ON a.id=ba.award_id WHERE ba.book_id=b.id AND lower(a.name)=lower($%d))`, argN))
		args = append(args, f.Award)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM books b WHERE `+whereClause, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search: %w", err)
	}

	orderBy := "b.title ASC"
	switch f.Sort {
	case "newest":
		orderBy = "(SELECT MAX(be.published_at) FROM book_editions be WHERE be.book_id=b.id) DESC NULLS LAST"
	case "oldest":
		orderBy = "(SELECT MIN(be.published_at) FROM book_editions be WHERE be.book_id=b.id) ASC NULLS LAST"
	case "relevance":
		if f.Query != "" {
			orderBy = "ts_rank(b.search_vector, plainto_tsquery('english', $1)) DESC"
		}
	}

	queryArgs := make([]any, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, f.Limit, offset)

	q := fmt.Sprintf(`
		SELECT b.id, b.title, b.description, b.series_id, b.series_position, b.status,
		       b.rejection_reason, b.deleted_at, b.created_at, b.updated_at
		FROM books b
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argN, argN+1)

	rows, err := r.db.Query(ctx, q, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("search books: %w", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Description, &b.SeriesID, &b.SeriesPosition,
			&b.Status, &b.RejectionReason, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, err
		}
		b.Authors = []Contributor{}
		b.Genres = []Genre{}
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Batch-load authors and genres for results
	if len(books) > 0 {
		ids := make([]string, len(books))
		idx := map[string]int{}
		for i, b := range books {
			ids[i] = b.ID
			idx[b.ID] = i
		}
		r.batchLoadAuthors(ctx, ids, idx, books)
		r.batchLoadGenres(ctx, ids, idx, books)
		r.batchLoadEditionCovers(ctx, ids, idx, books)
	}

	return books, total, nil
}

func (r *Repository) ListApprovedBooks(ctx context.Context, page, limit int, sortBy string) ([]Book, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM books WHERE status='approved' AND deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count books: %w", err)
	}

	orderBy := "title ASC"
	switch sortBy {
	case "newest":
		orderBy = "created_at DESC"
	case "oldest":
		orderBy = "created_at ASC"
	}

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, title, description, series_id, series_position, status,
		       rejection_reason, deleted_at, created_at, updated_at
		FROM books WHERE status='approved' AND deleted_at IS NULL
		ORDER BY %s LIMIT $1 OFFSET $2`, orderBy), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Description, &b.SeriesID, &b.SeriesPosition,
			&b.Status, &b.RejectionReason, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, err
		}
		b.Authors = []Contributor{}
		b.Genres = []Genre{}
		books = append(books, b)
	}
	rows.Close()

	if len(books) > 0 {
		ids := make([]string, len(books))
		idx := map[string]int{}
		for i, b := range books {
			ids[i] = b.ID
			idx[b.ID] = i
		}
		r.batchLoadAuthors(ctx, ids, idx, books)
		r.batchLoadGenres(ctx, ids, idx, books)
		r.batchLoadEditionCovers(ctx, ids, idx, books)
	}

	return books, total, nil
}

// batchLoadAuthors populates Authors on each book in the slice.
func (r *Repository) batchLoadAuthors(ctx context.Context, ids []string, idx map[string]int, books []Book) {
	rows, err := r.db.Query(ctx, `
		SELECT bc.book_id, c.id, c.name, c.status, c.deleted_at, c.created_at, c.updated_at
		FROM contributors c JOIN book_contributors bc ON bc.contributor_id=c.id
		WHERE bc.book_id=ANY($1) AND bc.role IN ('author','co_author') AND c.deleted_at IS NULL
		ORDER BY c.name`, ids)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var bookID string
		var c Contributor
		if err := rows.Scan(&bookID, &c.ID, &c.Name, &c.Status, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt); err == nil {
			if i, ok := idx[bookID]; ok {
				books[i].Authors = append(books[i].Authors, c)
			}
		}
	}
}

// batchLoadGenres populates Genres on each book in the slice.
func (r *Repository) batchLoadGenres(ctx context.Context, ids []string, idx map[string]int, books []Book) {
	rows, err := r.db.Query(ctx, `
		SELECT bg.book_id, g.id, g.name, g.status, g.rejection_reason, g.created_at
		FROM genres g JOIN book_genres bg ON bg.genre_id=g.id
		WHERE bg.book_id=ANY($1) ORDER BY g.name`, ids)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var bookID string
		var g Genre
		if err := rows.Scan(&bookID, &g.ID, &g.Name, &g.Status, &g.RejectionReason, &g.CreatedAt); err == nil {
			if i, ok := idx[bookID]; ok {
				books[i].Genres = append(books[i].Genres, g)
			}
		}
	}
}

// batchLoadEditionCovers loads the first cover_url for each book into a synthetic edition entry.
func (r *Repository) batchLoadEditionCovers(ctx context.Context, ids []string, idx map[string]int, books []Book) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (book_id) book_id, id, format, language, cover_url, status
		FROM book_editions
		WHERE book_id=ANY($1) AND deleted_at IS NULL
		ORDER BY book_id, published_at DESC NULLS LAST`, ids)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var bookID string
		var e Edition
		if err := rows.Scan(&bookID, &e.ID, &e.Format, &e.Language, &e.CoverURL, &e.Status); err == nil {
			if i, ok := idx[bookID]; ok {
				books[i].Editions = []Edition{e}
			}
		}
	}
}

func (r *Repository) UpdateBook(ctx context.Context, id string, title, description *string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE books SET
			title       = COALESCE($2, title),
			description = COALESCE($3, description),
			updated_at  = NOW()
		WHERE id=$1 AND deleted_at IS NULL`, id, title, description)
	if err != nil {
		return fmt.Errorf("update book: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("book not found")
	}
	return nil
}

func (r *Repository) UpdateEditionDetails(ctx context.Context, editionID string, e Edition) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE book_editions SET
			title            = COALESCE(NULLIF($2,''), title),
			original_title   = COALESCE(NULLIF($3,''), original_title),
			description      = $4,
			cover_url        = $5,
			format           = COALESCE(NULLIF($6,''), format),
			isbn10           = $7,
			isbn13           = $8,
			asin             = $9,
			language         = COALESCE(NULLIF($10,''), language),
			publisher        = $11,
			edition          = $12,
			published_at     = $13,
			page_count       = $14,
			file_format      = NULLIF($15,''),
			duration_minutes = $16,
			audio_format     = NULLIF($17,''),
			updated_at       = NOW()
		WHERE id=$1 AND deleted_at IS NULL`,
		editionID, e.Title, e.OriginalTitle, e.Description, e.CoverURL,
		e.Format, e.ISBN10, e.ISBN13, e.ASIN, e.Language,
		e.Publisher, e.Edition, e.PublishedAt, e.PageCount,
		e.FileFormat, e.DurationMinutes, e.AudioFormat,
	)
	if err != nil {
		return fmt.Errorf("update edition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("edition not found")
	}
	return nil
}

func (r *Repository) UpdateEditionCoverURL(ctx context.Context, editionID, coverURL string) error {
	_, err := r.db.Exec(ctx, `UPDATE book_editions SET cover_url=$2, updated_at=NOW() WHERE id=$1`, editionID, coverURL)
	return err
}

func (r *Repository) DeleteBook(ctx context.Context, id string) error {
	// Prevent deletion if active copies exist
	var n int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM book_copies WHERE edition_id IN (SELECT id FROM book_editions WHERE book_id=$1) AND deleted_at IS NULL`, id).Scan(&n)
	if n > 0 {
		return fmt.Errorf("cannot delete book with existing copies")
	}
	tag, err := r.db.Exec(ctx, `UPDATE books SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("book not found")
	}
	return nil
}

func (r *Repository) DeleteEdition(ctx context.Context, editionID string) error {
	tag, err := r.db.Exec(ctx, `UPDATE book_editions SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, editionID)
	if err != nil {
		return fmt.Errorf("delete edition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("edition not found")
	}
	return nil
}

func (r *Repository) FindBooksWithoutCovers(ctx context.Context) ([]Book, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT b.id, b.title
		FROM books b
		JOIN book_editions be ON be.book_id=b.id
		WHERE b.deleted_at IS NULL AND b.status='approved'
		  AND be.deleted_at IS NULL
		  AND (be.cover_url IS NULL OR be.cover_url='')
		  AND (be.isbn10 IS NOT NULL OR be.isbn13 IS NOT NULL)`)
	if err != nil {
		return nil, fmt.Errorf("find books without covers: %w", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	rows.Close()

	// Load editions for each book
	for i, b := range books {
		eRows, err := r.db.Query(ctx, `
			SELECT id, isbn10, isbn13, cover_url FROM book_editions
			WHERE book_id=$1 AND deleted_at IS NULL AND (isbn10 IS NOT NULL OR isbn13 IS NOT NULL)`, b.ID)
		if err != nil {
			continue
		}
		for eRows.Next() {
			var e Edition
			if err := eRows.Scan(&e.ID, &e.ISBN10, &e.ISBN13, &e.CoverURL); err == nil {
				books[i].Editions = append(books[i].Editions, e)
			}
		}
		eRows.Close()
	}
	return books, nil
}

func (r *Repository) GetUserBooks(ctx context.Context, userID string, page, limit int, status, genre, search, sortBy string) ([]UserBook, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	args := []any{userID}
	argN := 2
	where := []string{"bc.owner_id=$1", "bc.deleted_at IS NULL", "be.deleted_at IS NULL", "b.deleted_at IS NULL"}

	if status != "" {
		where = append(where, fmt.Sprintf("bc.reading_status=$%d", argN))
		args = append(args, status)
		argN++
	}
	if genre != "" {
		where = append(where, fmt.Sprintf(`EXISTS (SELECT 1 FROM book_genres bg JOIN genres g ON g.id=bg.genre_id WHERE bg.book_id=b.id AND lower(g.name)=lower($%d))`, argN))
		args = append(args, genre)
		argN++
	}
	if search != "" {
		where = append(where, fmt.Sprintf(`(lower(b.title) LIKE lower('%%'||$%d||'%%') OR EXISTS (SELECT 1 FROM contributors c JOIN book_contributors bco ON bco.contributor_id=c.id WHERE bco.book_id=b.id AND lower(c.name) LIKE lower('%%'||$%d||'%%')))`, argN, argN))
		args = append(args, search)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN books b ON b.id=be.book_id
		WHERE `+whereClause, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user books: %w", err)
	}

	orderBy := "bc.created_at DESC"
	switch sortBy {
	case "title":
		orderBy = "b.title ASC"
	case "author":
		orderBy = "(SELECT MIN(c.name) FROM contributors c JOIN book_contributors bco ON bco.contributor_id=c.id WHERE bco.book_id=b.id) ASC"
	case "date_read":
		orderBy = "bc.finished_reading_at DESC NULLS LAST"
	}

	queryArgs := make([]any, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT bc.id, bc.reading_status, bc.current_page, bc.started_reading_at, bc.finished_reading_at,
		       bc.owned_by_user, bc.borrowed_from, bc.location, bc.condition, bc.reread_count,
		       bc.personal_notes, bc.created_at,
		       be.id, be.format, be.language, be.cover_url,
		       b.id, b.title, b.description, b.series_id, b.series_position, b.status,
		       b.deleted_at, b.created_at, b.updated_at
		FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN books b ON b.id=be.book_id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argN, argN+1), queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list user books: %w", err)
	}
	defer rows.Close()

	var userBooks []UserBook
	bookIDs := []string{}
	seen := map[string]bool{}
	bookIdx := map[string][]int{} // bookID -> indices in userBooks

	for rows.Next() {
		var ub UserBook
		if err := rows.Scan(
			&ub.CopyID, &ub.ReadingStatus, &ub.CurrentPage, &ub.StartedReadingAt, &ub.FinishedReadingAt,
			&ub.OwnedByUser, &ub.BorrowedFrom, &ub.Location, &ub.Condition, &ub.RereadCount,
			&ub.PersonalNotes, &ub.AddedAt,
			&ub.EditionID, &ub.Format, &ub.Language, &ub.CoverURL,
			&ub.Book.ID, &ub.Book.Title, &ub.Book.Description, &ub.Book.SeriesID, &ub.Book.SeriesPosition,
			&ub.Book.Status, &ub.Book.DeletedAt, &ub.Book.CreatedAt, &ub.Book.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user book: %w", err)
		}
		ub.Book.Authors = []Contributor{}
		ub.Book.Genres = []Genre{}
		i := len(userBooks)
		userBooks = append(userBooks, ub)
		bookIdx[ub.Book.ID] = append(bookIdx[ub.Book.ID], i)
		if !seen[ub.Book.ID] {
			seen[ub.Book.ID] = true
			bookIDs = append(bookIDs, ub.Book.ID)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(bookIDs) > 0 {
		aRows, err := r.db.Query(ctx, `
			SELECT bc.book_id, c.id, c.name, c.status, c.deleted_at, c.created_at, c.updated_at
			FROM contributors c JOIN book_contributors bc ON bc.contributor_id=c.id
			WHERE bc.book_id=ANY($1) AND bc.role IN ('author','co_author') AND c.deleted_at IS NULL`, bookIDs)
		if err == nil {
			for aRows.Next() {
				var bookID string
				var c Contributor
				if err := aRows.Scan(&bookID, &c.ID, &c.Name, &c.Status, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt); err == nil {
					for _, i := range bookIdx[bookID] {
						userBooks[i].Book.Authors = append(userBooks[i].Book.Authors, c)
					}
				}
			}
			aRows.Close()
		}
	}

	return userBooks, total, nil
}

func (r *Repository) UpdateCopyStatus(ctx context.Context, copyID, userID, status string, currentPage *int) error {
	var setClause string
	switch status {
	case "reading":
		setClause = "reading_status=$3, started_reading_at=COALESCE(started_reading_at, NOW()), current_page=COALESCE($4, current_page), updated_at=NOW()"
	case "read":
		setClause = "reading_status=$3, finished_reading_at=NOW(), current_page=COALESCE($4, current_page), updated_at=NOW()"
	default:
		setClause = "reading_status=$3, current_page=COALESCE($4, current_page), updated_at=NOW()"
	}
	tag, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE book_copies SET %s WHERE id=$1 AND owner_id=$2 AND deleted_at IS NULL`, setClause),
		copyID, userID, status, currentPage)
	if err != nil {
		return fmt.Errorf("update copy status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("copy not found")
	}
	return nil
}

func (r *Repository) RemoveCopy(ctx context.Context, copyID, userID string) error {
	tag, err := r.db.Exec(ctx, `UPDATE book_copies SET deleted_at=NOW() WHERE id=$1 AND owner_id=$2 AND deleted_at IS NULL`, copyID, userID)
	if err != nil {
		return fmt.Errorf("remove copy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("copy not found")
	}
	return nil
}

func (r *Repository) UpdateCopyDetails(ctx context.Context, copyID, userID string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	setClauses := []string{}
	args := []any{copyID, userID}
	argN := 3
	for k, v := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s=$%d", k, argN))
		args = append(args, v)
		argN++
	}
	setClauses = append(setClauses, "updated_at=NOW()")
	q := fmt.Sprintf(`UPDATE book_copies SET %s WHERE id=$1 AND owner_id=$2 AND deleted_at IS NULL`, strings.Join(setClauses, ", "))
	_, err := r.db.Exec(ctx, q, args...)
	return err
}

func (r *Repository) InsertModerationLog(ctx context.Context, moderatorID, entityType, entityID, action string, before, after any) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO moderation_log (moderator_id, entity_type, entity_id, action, before, after)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		moderatorID, entityType, entityID, action, before, after)
	if err != nil {
		return fmt.Errorf("insert moderation log: %w", err)
	}
	return nil
}

func (r *Repository) ListPendingSubmissions(ctx context.Context, page, limit int) ([]Submission, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM submissions WHERE status='pending' AND deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count submissions: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, submitted_by, status, catalogue_only, rejection_reason, reviewed_by,
		       reviewed_at, book_id, edition_id, copy_id, contributor_id, deleted_at, created_at, updated_at
		FROM submissions WHERE status='pending' AND deleted_at IS NULL
		ORDER BY created_at ASC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list submissions: %w", err)
	}
	defer rows.Close()

	var subs []Submission
	for rows.Next() {
		var s Submission
		if err := rows.Scan(&s.ID, &s.SubmittedBy, &s.Status, &s.CatalogueOnly, &s.RejectionReason,
			&s.ReviewedBy, &s.ReviewedAt, &s.BookID, &s.EditionID, &s.CopyID, &s.ContributorID,
			&s.DeletedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		subs = append(subs, s)
	}
	return subs, total, rows.Err()
}

func (r *Repository) FindSubmissionByID(ctx context.Context, id string) (*Submission, error) {
	var s Submission
	err := r.db.QueryRow(ctx, `
		SELECT id, submitted_by, status, catalogue_only, rejection_reason, reviewed_by,
		       reviewed_at, book_id, edition_id, copy_id, contributor_id, deleted_at, created_at, updated_at
		FROM submissions WHERE id=$1 AND deleted_at IS NULL`, id).Scan(
		&s.ID, &s.SubmittedBy, &s.Status, &s.CatalogueOnly, &s.RejectionReason,
		&s.ReviewedBy, &s.ReviewedAt, &s.BookID, &s.EditionID, &s.CopyID, &s.ContributorID,
		&s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find submission: %w", err)
	}
	// Enrich with book/edition details
	if s.BookID != nil {
		s.Book, _ = r.FindBookWithDetails(ctx, *s.BookID)
	}
	if s.EditionID != nil {
		s.Edition, _ = r.FindEditionByID(ctx, *s.EditionID)
	}
	return &s, nil
}

func (r *Repository) ApproveSubmission(ctx context.Context, id, reviewerID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE submissions SET status='approved', reviewed_by=$2, reviewed_at=NOW(), updated_at=NOW()
		WHERE id=$1 AND status='pending'`, id, reviewerID)
	if err != nil {
		return fmt.Errorf("approve submission: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("submission not found or already reviewed")
	}
	return nil
}

func (r *Repository) RejectSubmission(ctx context.Context, id, reviewerID, reason string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE submissions SET status='rejected', reviewed_by=$2, reviewed_at=NOW(),
		       rejection_reason=$3, updated_at=NOW()
		WHERE id=$1 AND status='pending'`, id, reviewerID, reason)
	if err != nil {
		return fmt.Errorf("reject submission: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("submission not found or already reviewed")
	}
	return nil
}

func (r *Repository) ApproveBookEntities(ctx context.Context, bookID, editionID string) error {
	if _, err := r.db.Exec(ctx, `UPDATE books SET status='approved', updated_at=NOW() WHERE id=$1`, bookID); err != nil {
		return fmt.Errorf("approve book: %w", err)
	}
	if _, err := r.db.Exec(ctx, `UPDATE book_editions SET status='approved', updated_at=NOW() WHERE id=$1`, editionID); err != nil {
		return fmt.Errorf("approve edition: %w", err)
	}
	if _, err := r.db.Exec(ctx, `
		UPDATE contributors SET status='approved', updated_at=NOW()
		WHERE id IN (SELECT contributor_id FROM book_contributors WHERE book_id=$1) AND status='pending'`, bookID); err != nil {
		return fmt.Errorf("approve book contributors: %w", err)
	}
	if _, err := r.db.Exec(ctx, `
		UPDATE contributors SET status='approved', updated_at=NOW()
		WHERE id IN (SELECT contributor_id FROM edition_contributors WHERE edition_id=$1) AND status='pending'`, editionID); err != nil {
		return fmt.Errorf("approve edition contributors: %w", err)
	}
	if _, err := r.db.Exec(ctx, `
		UPDATE genres SET status='approved'
		WHERE id IN (SELECT genre_id FROM book_genres WHERE book_id=$1) AND status='pending'`, bookID); err != nil {
		return fmt.Errorf("approve genres: %w", err)
	}
	return nil
}

// ─── txRepository methods ────────────────────────────────────────────────────

func (r *txRepository) FindOrCreateContributor(ctx context.Context, name string, autoApprove bool) (Contributor, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var c Contributor
	err := r.db.QueryRow(ctx, `
		SELECT id, name, status, deleted_at, created_at, updated_at
		FROM contributors WHERE name=$1 AND deleted_at IS NULL LIMIT 1`, name).Scan(
		&c.ID, &c.Name, &c.Status, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt)
	if err == nil {
		return c, nil
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO contributors (name, status) VALUES ($1, $2)
		RETURNING id, name, status, deleted_at, created_at, updated_at`, name, status).Scan(
		&c.ID, &c.Name, &c.Status, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return Contributor{}, fmt.Errorf("create contributor: %w", err)
	}
	return c, nil
}

func (r *txRepository) FindOrCreateGenre(ctx context.Context, name string, autoApprove bool) (Genre, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var g Genre
	err := r.db.QueryRow(ctx, `
		INSERT INTO genres (name, status) VALUES ($1, $2)
		ON CONFLICT (name) DO NOTHING
		RETURNING id, name, status, rejection_reason, created_at`, name, status).Scan(
		&g.ID, &g.Name, &g.Status, &g.RejectionReason, &g.CreatedAt)
	if err != nil {
		// Already exists — fetch it
		err = r.db.QueryRow(ctx, `SELECT id, name, status, rejection_reason, created_at FROM genres WHERE name=$1`, name).Scan(
			&g.ID, &g.Name, &g.Status, &g.RejectionReason, &g.CreatedAt)
		if err != nil {
			return Genre{}, fmt.Errorf("find genre: %w", err)
		}
	}
	return g, nil
}

func (r *txRepository) FindOrCreateSeries(ctx context.Context, name string, autoApprove bool) (Series, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var s Series
	err := r.db.QueryRow(ctx, `
		INSERT INTO series (name, status) VALUES ($1, $2)
		ON CONFLICT (name) DO NOTHING
		RETURNING id, name, description, status, rejection_reason, deleted_at, created_at, updated_at`, name, status).Scan(
		&s.ID, &s.Name, &s.Description, &s.Status, &s.RejectionReason, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		err = r.db.QueryRow(ctx, `
			SELECT id, name, description, status, rejection_reason, deleted_at, created_at, updated_at
			FROM series WHERE name=$1 AND deleted_at IS NULL`, name).Scan(
			&s.ID, &s.Name, &s.Description, &s.Status, &s.RejectionReason, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return Series{}, fmt.Errorf("find series: %w", err)
		}
	}
	return s, nil
}

func (r *txRepository) InsertBook(ctx context.Context, title string, description *string, autoApprove bool) (Book, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var b Book
	err := r.db.QueryRow(ctx, `
		INSERT INTO books (title, description, status) VALUES ($1, $2, $3)
		RETURNING id, title, description, series_id, series_position, status,
		          rejection_reason, deleted_at, created_at, updated_at`,
		title, description, status).Scan(
		&b.ID, &b.Title, &b.Description, &b.SeriesID, &b.SeriesPosition,
		&b.Status, &b.RejectionReason, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return Book{}, fmt.Errorf("insert book: %w", err)
	}
	return b, nil
}

func (r *txRepository) InsertEdition(ctx context.Context, e Edition, autoApprove bool) (Edition, error) {
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	var out Edition
	err := r.db.QueryRow(ctx, `
		INSERT INTO book_editions (
			book_id, title, original_title, format, description, cover_url,
			isbn10, isbn13, asin, language, publisher, edition, published_at,
			page_count, file_format, duration_minutes, audio_format, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING id, book_id, title, original_title, format, description, cover_url,
		          isbn10, isbn13, asin, language, publisher, edition, published_at,
		          page_count, file_format, duration_minutes, audio_format, status,
		          rejection_reason, deleted_at, created_at, updated_at`,
		e.BookID, e.Title, e.OriginalTitle, e.Format, e.Description, e.CoverURL,
		e.ISBN10, e.ISBN13, e.ASIN, e.Language, e.Publisher, e.Edition, e.PublishedAt,
		e.PageCount, e.FileFormat, e.DurationMinutes, e.AudioFormat, status,
	).Scan(
		&out.ID, &out.BookID, &out.Title, &out.OriginalTitle, &out.Format, &out.Description, &out.CoverURL,
		&out.ISBN10, &out.ISBN13, &out.ASIN, &out.Language, &out.Publisher, &out.Edition, &out.PublishedAt,
		&out.PageCount, &out.FileFormat, &out.DurationMinutes, &out.AudioFormat, &out.Status,
		&out.RejectionReason, &out.DeletedAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return Edition{}, fmt.Errorf("insert edition: %w", err)
	}
	return out, nil
}

func (r *txRepository) InsertCopy(ctx context.Context, editionID, ownerID string, condition *string, opts CopyOptions) (Copy, error) {
	status := opts.ReadingStatus
	if status == "" {
		status = "want_to_read"
	}
	ownedByUser := true
	if opts.OwnedByUser != nil {
		ownedByUser = *opts.OwnedByUser
	}
	var c Copy
	err := r.db.QueryRow(ctx, `
		INSERT INTO book_copies (edition_id, owner_id, condition, reading_status, current_page,
			owned_by_user, borrowed_from, location)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, edition_id, owner_id, condition, reading_status, current_page,
		          started_reading_at, finished_reading_at, owned_by_user, borrowed_from,
		          location, reread_count, personal_notes, deleted_at, created_at, updated_at`,
		editionID, ownerID, condition, status, opts.CurrentPage,
		ownedByUser, opts.BorrowedFrom, opts.Location,
	).Scan(
		&c.ID, &c.EditionID, &c.OwnerID, &c.Condition, &c.ReadingStatus, &c.CurrentPage,
		&c.StartedReadingAt, &c.FinishedReadingAt, &c.OwnedByUser, &c.BorrowedFrom,
		&c.Location, &c.RereadCount, &c.PersonalNotes, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return Copy{}, fmt.Errorf("insert copy: %w", err)
	}
	return c, nil
}

func (r *txRepository) LinkBookContributor(ctx context.Context, bookID, contributorID, role string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO book_contributors (book_id, contributor_id, role) VALUES ($1,$2,$3)
		ON CONFLICT DO NOTHING`, bookID, contributorID, role)
	return err
}

func (r *txRepository) LinkBookGenre(ctx context.Context, bookID, genreID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO book_genres (book_id, genre_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, bookID, genreID)
	return err
}

func (r *txRepository) LinkEditionContributor(ctx context.Context, editionID, contributorID, role string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO edition_contributors (edition_id, contributor_id, role) VALUES ($1,$2,$3)
		ON CONFLICT DO NOTHING`, editionID, contributorID, role)
	return err
}

func (r *txRepository) UpdateBookSeries(ctx context.Context, bookID, seriesID string, position *float64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE books SET series_id=$2, series_position=$3, updated_at=NOW() WHERE id=$1`,
		bookID, seriesID, position)
	return err
}

func (r *txRepository) InsertSubmission(ctx context.Context, userID, bookID, editionID string) (Submission, error) {
	var s Submission
	err := r.db.QueryRow(ctx, `
		INSERT INTO submissions (submitted_by, book_id, edition_id, status, catalogue_only)
		VALUES ($1,$2,$3,'pending',false)
		RETURNING id, submitted_by, status, catalogue_only, book_id, edition_id, copy_id,
		          deleted_at, created_at, updated_at`,
		userID, bookID, editionID).Scan(
		&s.ID, &s.SubmittedBy, &s.Status, &s.CatalogueOnly, &s.BookID, &s.EditionID, &s.CopyID,
		&s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return Submission{}, fmt.Errorf("insert submission: %w", err)
	}
	return s, nil
}

func (r *txRepository) InsertSubmissionApproved(ctx context.Context, userID, bookID, editionID, copyID string) (Submission, error) {
	var s Submission
	err := r.db.QueryRow(ctx, `
		INSERT INTO submissions (submitted_by, book_id, edition_id, copy_id, status, catalogue_only)
		VALUES ($1,$2,$3,$4,'approved',false)
		RETURNING id, submitted_by, status, catalogue_only, book_id, edition_id, copy_id,
		          deleted_at, created_at, updated_at`,
		userID, bookID, editionID, copyID).Scan(
		&s.ID, &s.SubmittedBy, &s.Status, &s.CatalogueOnly, &s.BookID, &s.EditionID, &s.CopyID,
		&s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return Submission{}, fmt.Errorf("insert approved submission: %w", err)
	}
	return s, nil
}

func (r *txRepository) InsertSubmissionApprovedNoCopy(ctx context.Context, userID, bookID, editionID string) (Submission, error) {
	var s Submission
	err := r.db.QueryRow(ctx, `
		INSERT INTO submissions (submitted_by, book_id, edition_id, status, catalogue_only)
		VALUES ($1,$2,$3,'approved',true)
		RETURNING id, submitted_by, status, catalogue_only, book_id, edition_id, copy_id,
		          deleted_at, created_at, updated_at`,
		userID, bookID, editionID).Scan(
		&s.ID, &s.SubmittedBy, &s.Status, &s.CatalogueOnly, &s.BookID, &s.EditionID, &s.CopyID,
		&s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return Submission{}, fmt.Errorf("insert approved no-copy submission: %w", err)
	}
	return s, nil
}

// nullableTime converts a *time.Time to a value safe for pgx (nil stays nil).
func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

// pgxTx is a local alias so service.go can use pgx.Tx without importing pgx directly.
type pgxTx interface {
	DB
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// pgconn import alias to avoid unused import errors.
var _ pgconn.CommandTag

// WithDB is the exported version of withDB, used by the moderation service.
func (r *Repository) WithDB(db DB) *txRepository {
	return &txRepository{db: db}
}
