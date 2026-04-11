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

// withDB returns a new repository using the given DB (pool or tx)
func (r *Repository) withDB(db DB) *txRepository {
	return &txRepository{db: db}
}

// txRepository works with either a pool or a transaction
type txRepository struct {
	db DB
}

func (r *Repository) InsertCopyDirect(ctx context.Context, editionID string, ownerID string, condition *string) (Copy, error) {
	var copy Copy
	var query string = `
        INSERT INTO book_copies (edition_id, owner_id, condition)
        VALUES ($1, $2, $3)
        RETURNING id, edition_id, owner_id, condition, reading_status,
                  current_page, started_reading_at, finished_reading_at,
                  owned_by_user, borrowed_from, location,
                  deleted_at, created_at, updated_at
    `
	var err error = r.db.QueryRow(ctx, query, editionID, ownerID, condition).Scan(
		&copy.ID, &copy.EditionID, &copy.OwnerID, &copy.Condition,
		&copy.ReadingStatus, &copy.CurrentPage,
		&copy.StartedReadingAt, &copy.FinishedReadingAt,
		&copy.OwnedByUser, &copy.BorrowedFrom, &copy.Location,
		&copy.DeletedAt, &copy.CreatedAt, &copy.UpdatedAt,
	)
	if err != nil {
		return Copy{}, fmt.Errorf("failed to insert copy: %w", err)
	}
	return copy, nil
}

func (r *Repository) ApproveSubmissionWithCopy(ctx context.Context, id string, reviewerID string, copyID string) error {
	var query string = `
        UPDATE submissions
        SET status = 'approved', reviewed_by = $2, reviewed_at = NOW(),
            copy_id = $3, updated_at = NOW()
        WHERE id = $1
    `
	var _, err = r.db.Exec(ctx, query, id, reviewerID, copyID)
	if err != nil {
		return fmt.Errorf("failed to approve submission with copy: %w", err)
	}
	return nil
}

func (r *Repository) FindEditionByISBN(ctx context.Context, isbn string) (*Edition, error) {
	var edition Edition
	var query string = `
        SELECT id, book_id, format, description, cover_url, isbn10, isbn13, asin, language, publisher,
               edition, published_at, page_count, file_format,
               duration_minutes, audio_format, status, deleted_at, created_at, updated_at
        FROM book_editions
        WHERE (isbn10 = $1 OR isbn13 = $1) AND deleted_at IS NULL
        LIMIT 1
    `

	var err error = r.db.QueryRow(ctx, query, isbn).Scan(
		&edition.ID, &edition.BookID, &edition.Format, &edition.Description, &edition.CoverURL,
		&edition.ISBN10, &edition.ISBN13, &edition.ASIN, &edition.Language,
		&edition.Publisher, &edition.Edition, &edition.PublishedAt,
		&edition.PageCount, &edition.FileFormat, &edition.DurationMinutes,
		&edition.AudioFormat, &edition.Status, &edition.DeletedAt,
		&edition.CreatedAt, &edition.UpdatedAt,
	)
	if err != nil {
		return nil, nil
	}

	return &edition, nil
}

func (r *Repository) FindBookByID(ctx context.Context, bookID string) (*Book, error) {
	var book Book
	var query string = `
        SELECT id, title, description, series_id, series_position, status, deleted_at, created_at, updated_at
        FROM books
        WHERE id = $1 AND deleted_at IS NULL
    `

	var err error = r.db.QueryRow(ctx, query, bookID).Scan(
		&book.ID, &book.Title, &book.Description, &book.SeriesID,
		&book.SeriesPosition, &book.Status, &book.DeletedAt,
		&book.CreatedAt, &book.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	return &book, nil
}

// FindBookByTitleAndAuthors returns a book whose title matches (case-insensitive) and whose
// author set is identical to the provided list. Returns nil (no error) if not found.
func (r *Repository) FindBookByTitleAndAuthors(ctx context.Context, title string, authorNames []string) (*Book, error) {
	if len(authorNames) == 0 {
		return nil, nil
	}
	// Find books with the same title
	rows, err := r.db.Query(ctx, `
        SELECT b.id, b.title, b.status, b.deleted_at, b.created_at, b.updated_at
        FROM books b
        WHERE lower(b.title) = lower($1) AND b.deleted_at IS NULL
    `, title)
	if err != nil {
		return nil, fmt.Errorf("failed to search books: %w", err)
	}
	defer rows.Close()

	var candidates []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title,
			&b.Status, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		candidates = append(candidates, b)
	}

	// For each candidate, compare its authors
	for _, candidate := range candidates {
		authorRows, err := r.db.Query(ctx, `
            SELECT a.name FROM contributors a
            JOIN book_contributors bc ON bc.contributor_id = a.id
           WHERE bc.book_id = $1 AND bc.role IN ('author', 'co_author') AND a.deleted_at IS NULL
            ORDER BY a.name
        `, candidate.ID)
		if err != nil {
			continue
		}
		var existing []string
		for authorRows.Next() {
			var name string
			if err := authorRows.Scan(&name); err == nil {
				existing = append(existing, name)
			}
		}
		authorRows.Close()

		// Normalise input for comparison
		var inputNorm []string
		for _, n := range authorNames {
			inputNorm = append(inputNorm, strings.ToLower(strings.TrimSpace(n)))
		}
		var existNorm []string
		for _, n := range existing {
			existNorm = append(existNorm, strings.ToLower(strings.TrimSpace(n)))
		}
		sort.Strings(inputNorm)
		sort.Strings(existNorm)

		if len(inputNorm) == len(existNorm) {
			match := true
			for i := range inputNorm {
				if inputNorm[i] != existNorm[i] {
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

func (r *Repository) SearchBooks(ctx context.Context, query string, page, limit int) ([]Book, int, error) {
	offset := (page - 1) * limit

	var countQuery string = `
        SELECT COUNT(*)
        FROM books b
        WHERE b.status = 'approved' AND b.deleted_at IS NULL
          AND (
              b.search_vector @@ plainto_tsquery('english', $1)
              OR lower(b.title) LIKE lower('%' || $1 || '%')
              OR lower(coalesce(b.description, '')) LIKE lower('%' || $1 || '%')
          )
    `
	var total int
	if err := r.db.QueryRow(ctx, countQuery, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count: %w", err)
	}

	var queryStr string = `
        SELECT b.id, b.title, b.description, b.series_id, b.series_position, b.status, b.deleted_at, b.created_at, b.updated_at
        FROM books b
        WHERE b.status = 'approved' AND b.deleted_at IS NULL
          AND (
              b.search_vector @@ plainto_tsquery('english', $1)
              OR lower(b.title) LIKE lower('%' || $1 || '%')
              OR lower(coalesce(b.description, '')) LIKE lower('%' || $1 || '%')
          )
        ORDER BY b.title ASC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(ctx, queryStr, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search books: %w", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Description, &b.SeriesID, &b.SeriesPosition, &b.Status, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan book: %w", err)
		}
		b.Authors = []Author{}
		b.Genres = []Genre{}
		books = append(books, b)
	}
	return books, total, rows.Err()
}

func (r *Repository) ListPendingSubmissions(ctx context.Context, page int, limit int) ([]Submission, int, error) {
	var offset int = (page - 1) * limit

	var countQuery string = `
        SELECT COUNT(*) FROM submissions
        WHERE status = 'pending' AND deleted_at IS NULL
    `
	var total int
	var err error = r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count submissions: %w", err)
	}

	var query string = `
        SELECT id, submitted_by, status, catalogue_only, rejection_reason, reviewed_by,
               reviewed_at, book_id, edition_id, copy_id, deleted_at, created_at, updated_at
        FROM submissions
        WHERE status = 'pending' AND deleted_at IS NULL
        ORDER BY created_at ASC
        LIMIT $1 OFFSET $2
    `

	var rows pgx.Rows
	rows, err = r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list submissions: %w", err)
	}
	defer rows.Close()

	var submissions []Submission
	for rows.Next() {
		var s Submission
		err = rows.Scan(
			&s.ID, &s.SubmittedBy, &s.Status, &s.CatalogueOnly, &s.RejectionReason,
			&s.ReviewedBy, &s.ReviewedAt, &s.BookID, &s.EditionID,
			&s.CopyID, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan submission: %w", err)
		}
		submissions = append(submissions, s)
	}

	return submissions, total, nil
}

func (r *Repository) FindSubmissionByID(ctx context.Context, id string) (*Submission, error) {
	var s Submission
	var query string = `
        SELECT id, submitted_by, status, catalogue_only, rejection_reason, reviewed_by,
               reviewed_at, book_id, edition_id, copy_id, deleted_at, created_at, updated_at
        FROM submissions
        WHERE id = $1 AND deleted_at IS NULL
    `

	var err error = r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.SubmittedBy, &s.Status, &s.CatalogueOnly, &s.RejectionReason,
		&s.ReviewedBy, &s.ReviewedAt, &s.BookID, &s.EditionID,
		&s.CopyID, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	return &s, nil
}

func (r *Repository) ApproveSubmission(ctx context.Context, id string, reviewerID string) error {
	var query string = `
        UPDATE submissions
        SET status = 'approved', reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW()
        WHERE id = $1 AND status = 'pending'
    `
	var tag pgconn.CommandTag
	var err error
	tag, err = r.db.Exec(ctx, query, id, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to approve submission: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("submission not found or already reviewed")
	}
	return nil
}

func (r *Repository) RejectSubmission(ctx context.Context, id string, reviewerID string, reason string) error {
	var query string = `
        UPDATE submissions
        SET status = 'rejected', reviewed_by = $2, reviewed_at = NOW(),
            rejection_reason = $3, updated_at = NOW()
        WHERE id = $1 AND status = 'pending'
    `
	var tag pgconn.CommandTag
	var err error
	tag, err = r.db.Exec(ctx, query, id, reviewerID, reason)
	if err != nil {
		return fmt.Errorf("failed to reject submission: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("submission not found or already reviewed")
	}
	return nil
}

func (r *Repository) ApproveBookEntities(ctx context.Context, bookID string, editionID string) error {
	var err error

	var bookQuery string = `UPDATE books SET status = 'approved', updated_at = NOW() WHERE id = $1`
	_, err = r.db.Exec(ctx, bookQuery, bookID)
	if err != nil {
		return fmt.Errorf("failed to approve book: %w", err)
	}

	var editionQuery string = `UPDATE book_editions SET status = 'approved', updated_at = NOW() WHERE id = $1`
	_, err = r.db.Exec(ctx, editionQuery, editionID)
	if err != nil {
		return fmt.Errorf("failed to approve edition: %w", err)
	}

	var authorQuery string = `
        UPDATE contributors SET status = 'approved', updated_at = NOW()
        WHERE id IN (SELECT contributor_id FROM book_contributors WHERE book_id = $1)
          AND status = 'pending'
    `
	_, err = r.db.Exec(ctx, authorQuery, bookID)
	if err != nil {
		return fmt.Errorf("failed to approve authors: %w", err)
	}

	var editionContributors string = `
    UPDATE contributors SET status = 'approved', updated_at = NOW()
    WHERE id IN (SELECT contributor_id FROM edition_contributors WHERE edition_id = $1)
    AND status = 'pending'
`
	_, err = r.db.Exec(ctx, editionContributors, editionID)
	if err != nil {
		return fmt.Errorf("failed to approve edition contributors: %w", err)
	}

	var genreQuery string = `
        UPDATE genres SET status = 'approved'
        WHERE id IN (SELECT genre_id FROM book_genres WHERE book_id = $1)
        AND status = 'pending'
    `
	_, err = r.db.Exec(ctx, genreQuery, bookID)
	if err != nil {
		return fmt.Errorf("failed to approve genres: %w", err)
	}

	return nil
}

func (r *Repository) ReplaceBookAuthors(ctx context.Context, bookID string, authorNames []string, autoApprove bool) error {
	_, err := r.db.Exec(ctx, `DELETE FROM book_contributors WHERE book_id = $1 AND role IN ('author', 'co_author')`, bookID)
	if err != nil {
		return fmt.Errorf("failed to clear book authors: %w", err)
	}
	for _, name := range authorNames {
		if name == "" {
			continue
		}
		var author Author
		var row pgx.Row = r.db.QueryRow(ctx, `SELECT id, name, status FROM contributors WHERE name = $1 AND deleted_at IS NULL LIMIT 1`, name)
		err = row.Scan(&author.ID, &author.Name, &author.Status)
		if err != nil {
			// Not found — create
			err = r.db.QueryRow(ctx,
				`INSERT INTO contributors (name, status) VALUES ($1, $2) RETURNING id, name, status`,
				name, map[bool]string{true: "approved", false: "pending"}[autoApprove],
			).Scan(&author.ID, &author.Name, &author.Status)
			if err != nil {
				return fmt.Errorf("failed to create author: %w", err)
			}
		}
		_, err = r.db.Exec(ctx, `INSERT INTO book_contributors (book_id, contributor_id, role) VALUES ($1, $2, 'author') ON CONFLICT DO NOTHING`, bookID, author.ID)
		if err != nil {
			return fmt.Errorf("failed to link author: %w", err)
		}
	}
	return nil
}

func (r *Repository) ReplaceBookGenres(ctx context.Context, bookID string, genreNames []string, autoApprove bool) error {
	_, err := r.db.Exec(ctx, `DELETE FROM book_genres WHERE book_id = $1`, bookID)
	if err != nil {
		return fmt.Errorf("failed to clear book genres: %w", err)
	}
	for _, name := range genreNames {
		if name == "" {
			continue
		}
		var genre Genre
		var row pgx.Row = r.db.QueryRow(ctx, `SELECT id, name, status FROM genres WHERE name = $1`, name)
		err = row.Scan(&genre.ID, &genre.Name, &genre.Status)
		if err != nil {
			err = r.db.QueryRow(ctx,
				`INSERT INTO genres (name, status) VALUES ($1, $2) RETURNING id, name, status`,
				name, map[bool]string{true: "approved", false: "pending"}[autoApprove],
			).Scan(&genre.ID, &genre.Name, &genre.Status)
			if err != nil {
				return fmt.Errorf("failed to create genre: %w", err)
			}
		}
		_, err = r.db.Exec(ctx, `INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, bookID, genre.ID)
		if err != nil {
			return fmt.Errorf("failed to link genre: %w", err)
		}
	}
	return nil
}

func (r *Repository) ReplaceEditionTranslators(ctx context.Context, editionID string, names []string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM edition_contributors WHERE edition_id = $1 AND role = 'translator'`, editionID)
	if err != nil {
		return fmt.Errorf("failed to clear edition translators: %w", err)
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		var id string
		err = r.db.QueryRow(ctx, `SELECT id FROM contributors WHERE name = $1 AND deleted_at IS NULL LIMIT 1`, name).Scan(&id)
		if err != nil {
			err = r.db.QueryRow(ctx, `INSERT INTO contributors (name, status) VALUES ($1, 'approved') RETURNING id`, name).Scan(&id)
			if err != nil {
				return fmt.Errorf("failed to create translator: %w", err)
			}
		}
		_, err = r.db.Exec(ctx, `INSERT INTO edition_contributors (edition_id, contributor_id, role) VALUES ($1, $2, 'translator') ON CONFLICT DO NOTHING`, editionID, id)
		if err != nil {
			return fmt.Errorf("failed to link translator: %w", err)
		}
	}
	return nil
}

func (r *Repository) UpdateEditionDetails(ctx context.Context, editionID string, e Edition) error {
	// Use COALESCE(NULLIF($n, ''), existing) so an empty string falls back to the
	// existing value instead of violating the NOT NULL / CHECK constraint on format.
	var query string = `
        UPDATE book_editions SET
            description     = $2,
            cover_url       = $3,
            format          = COALESCE(NULLIF($4, ''), format),
            isbn10          = $5,
            isbn13          = $6,
            asin            = $7,
            language        = COALESCE(NULLIF($8, ''), language),
            publisher       = $9,
            edition         = $10,
            published_at    = $11,
            page_count      = $12,
            file_format     = NULLIF($13, ''),
            duration_minutes = $14,
            audio_format    = NULLIF($15, ''),
            updated_at      = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	tag, err := r.db.Exec(ctx, query,
		editionID, e.Description, e.CoverURL, e.Format, e.ISBN10, e.ISBN13, e.ASIN, e.Language,
		e.Publisher, e.Edition, e.PublishedAt, e.PageCount,
		e.FileFormat, e.DurationMinutes, e.AudioFormat,
	)
	if err != nil {
		return fmt.Errorf("failed to update edition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("edition not found")
	}
	return nil
}

func (r *Repository) InsertModerationLog(ctx context.Context, moderatorID string, entityType string, entityID string, action string, before interface{}, after interface{}) error {
	var query string = `
        INSERT INTO moderation_log (moderator_id, entity_type, entity_id, action, before, after)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	var _, err = r.db.Exec(ctx, query, moderatorID, entityType, entityID, action, before, after)
	if err != nil {
		return fmt.Errorf("failed to insert moderation log: %w", err)
	}
	return nil
}

func (r *txRepository) FindOrCreateAuthor(ctx context.Context, name string, autoApprove bool) (Author, error) {
	var author Author
	var status string = "pending"
	if autoApprove {
		status = "approved"
	}

	// Try to find existing contributor first
	var fetchQuery string = `
        SELECT id, name, status, deleted_at, created_at, updated_at
        FROM contributors
        WHERE name = $1 AND deleted_at IS NULL
        LIMIT 1
    `
	var err error = r.db.QueryRow(ctx, fetchQuery, name).Scan(
		&author.ID,
		&author.Name,
		&author.Status,
		&author.DeletedAt,
		&author.CreatedAt,
		&author.UpdatedAt,
	)
	if err == nil {
		return author, nil
	}

	// Not found — insert
	var insertQuery string = `
        INSERT INTO contributors (name, status)
        VALUES ($1, $2)
        RETURNING id, name, status, deleted_at, created_at, updated_at
    `
	err = r.db.QueryRow(ctx, insertQuery, name, status).Scan(
		&author.ID,
		&author.Name,
		&author.Status,
		&author.DeletedAt,
		&author.CreatedAt,
		&author.UpdatedAt,
	)
	if err != nil {
		return Author{}, fmt.Errorf("failed to create author: %w", err)
	}

	return author, nil
}

func (r *txRepository) FindOrCreateGenre(ctx context.Context, name string, autoApprove bool) (Genre, error) {
	var genre Genre
	var status string = "pending"
	if autoApprove {
		status = "approved"
	}

	var query string = `
    INSERT INTO genres (name, status)
    VALUES ($1, $2)
    ON CONFLICT (name) DO NOTHING
    RETURNING id, name, status, created_at
`

	var err error = r.db.QueryRow(ctx, query, name, status).Scan(
		&genre.ID,
		&genre.Name,
		&genre.Status,
		&genre.CreatedAt,
	)
	if err != nil {
		var fetchQuery string = `
            SELECT id, name, status, created_at
            FROM genres WHERE name = $1
        `
		err = r.db.QueryRow(ctx, fetchQuery, name).Scan(
			&genre.ID,
			&genre.Name,
			&genre.Status,
			&genre.CreatedAt,
		)
		if err != nil {
			return Genre{}, fmt.Errorf("failed to find genre: %w", err)
		}
	}

	return genre, nil
}

func (r *txRepository) InsertBook(ctx context.Context, title string, autoApprove bool) (Book, error) {
	var book Book
	var status string = "pending"
	if autoApprove {
		status = "approved"
	}

	var query string = `
        INSERT INTO books (title, description, status)
        VALUES ($1, $2, $3)
        RETURNING id, title, description, series_id, series_position, status, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, title, nil, status).Scan(
		&book.ID,
		&book.Title,
		&book.Description,
		&book.SeriesID,
		&book.SeriesPosition,
		&book.Status,
		&book.DeletedAt,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
		return Book{}, fmt.Errorf("failed to insert book: %w", err)
	}

	return book, nil
}

func (r *txRepository) InsertEdition(ctx context.Context, e Edition, autoApprove bool) (Edition, error) {
	var edition Edition
	var status string = "pending"
	if autoApprove {
		status = "approved"
	}

	var query string = `
        INSERT INTO book_editions (
            book_id, format, description, cover_url, isbn10, isbn13, asin, language, publisher,
            edition, published_at, page_count, file_format,
            duration_minutes, audio_format, status
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
        RETURNING id, book_id, format, description, cover_url, isbn10, isbn13, asin, language, publisher,
            edition, published_at, page_count, file_format,
            duration_minutes, audio_format, status, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query,
		e.BookID, e.Format, e.Description, e.CoverURL, e.ISBN10, e.ISBN13, e.ASIN, e.Language,
		e.Publisher, e.Edition, e.PublishedAt, e.PageCount,
		e.FileFormat, e.DurationMinutes, e.AudioFormat, status,
	).Scan(
		&edition.ID, &edition.BookID, &edition.Format, &edition.Description, &edition.CoverURL,
		&edition.ISBN10, &edition.ISBN13, &edition.ASIN, &edition.Language,
		&edition.Publisher, &edition.Edition, &edition.PublishedAt,
		&edition.PageCount, &edition.FileFormat, &edition.DurationMinutes,
		&edition.AudioFormat, &edition.Status, &edition.DeletedAt,
		&edition.CreatedAt, &edition.UpdatedAt,
	)
	if err != nil {
		return Edition{}, fmt.Errorf("failed to insert edition: %w", err)
	}

	return edition, nil
}

func (r *txRepository) InsertCopy(ctx context.Context, editionID string, ownerID string, condition *string, opts CopyOptions) (Copy, error) {
	var copy Copy
	status := opts.ReadingStatus
	if status == "" {
		status = "want_to_read"
	}
	ownedByUser := true
	if opts.OwnedByUser != nil {
		ownedByUser = *opts.OwnedByUser
	}
	var query string = `
        INSERT INTO book_copies (edition_id, owner_id, condition,
            reading_status, current_page, started_reading_at, finished_reading_at,
            owned_by_user, borrowed_from, location)
        VALUES ($1, $2, $3, $4, $5,
            CASE WHEN $6::text = '' THEN NULL ELSE $6::timestamptz END,
            CASE WHEN $7::text = '' THEN NULL ELSE $7::timestamptz END,
            $8, $9, $10)
        RETURNING id, edition_id, owner_id, condition, reading_status,
                  current_page, started_reading_at, finished_reading_at,
                  owned_by_user, borrowed_from, location,
                  deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query,
		editionID, ownerID, condition,
		status, opts.CurrentPage,
		nullableString(opts.StartedReadingAt),
		nullableString(opts.FinishedReadingAt),
		ownedByUser, opts.BorrowedFrom, opts.Location,
	).Scan(
		&copy.ID, &copy.EditionID, &copy.OwnerID, &copy.Condition,
		&copy.ReadingStatus, &copy.CurrentPage,
		&copy.StartedReadingAt, &copy.FinishedReadingAt,
		&copy.OwnedByUser, &copy.BorrowedFrom, &copy.Location,
		&copy.DeletedAt, &copy.CreatedAt, &copy.UpdatedAt,
	)
	if err != nil {
		return Copy{}, fmt.Errorf("failed to insert copy: %w", err)
	}

	return copy, nil
}

func (r *txRepository) LinkBookAuthor(ctx context.Context, bookID string, authorID string) error {
	var query string = `
        INSERT INTO book_contributors (book_id, contributor_id, role) VALUES ($1, $2, 'author') ON CONFLICT DO NOTHING
    `
	var _, err = r.db.Exec(ctx, query, bookID, authorID)
	if err != nil {
		return fmt.Errorf("failed to link book author: %w", err)
	}
	return nil
}

func (r *txRepository) LinkBookGenre(ctx context.Context, bookID string, genreID string) error {
	var query string = `
        INSERT INTO book_genres (book_id, genre_id)
        VALUES ($1, $2) ON CONFLICT DO NOTHING
    `
	var _, err = r.db.Exec(ctx, query, bookID, genreID)
	if err != nil {
		return fmt.Errorf("failed to link book genre: %w", err)
	}
	return nil
}

func (r *txRepository) FindOrCreateNarrator(ctx context.Context, name string, autoApprove bool) (Narrator, error) {
	var narrator Narrator
	var status string = "pending"
	if autoApprove {
		status = "approved"
	}

	var fetchQuery string = `
		SELECT id, name, status, deleted_at, created_at, updated_at
		FROM contributors
		WHERE name = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	var err error = r.db.QueryRow(ctx, fetchQuery, name).Scan(
		&narrator.ID,
		&narrator.Name,
		&narrator.Status,
		&narrator.DeletedAt,
		&narrator.CreatedAt,
		&narrator.UpdatedAt,
	)
	if err == nil {
		return narrator, nil
	}

	var insertQuery string = `
		INSERT INTO contributors (name, status)
		VALUES ($1, $2)
		RETURNING id, name, status, deleted_at, created_at, updated_at
	`
	err = r.db.QueryRow(ctx, insertQuery, name, status).Scan(
		&narrator.ID,
		&narrator.Name,
		&narrator.Status,
		&narrator.DeletedAt,
		&narrator.CreatedAt,
		&narrator.UpdatedAt,
	)
	if err != nil {
		return Narrator{}, fmt.Errorf("failed to create narrator: %w", err)
	}

	return narrator, nil
}

func (r *txRepository) LinkEditionNarrator(ctx context.Context, editionID string, narratorID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO edition_contributors (edition_id, contributor_id, role) VALUES ($1, $2, 'narrator') ON CONFLICT DO NOTHING`,
		editionID, narratorID,
	)
	if err != nil {
		return fmt.Errorf("failed to link narrator: %w", err)
	}
	return nil
}

func (r *txRepository) FindOrCreateTranslator(ctx context.Context, name string, autoApprove bool) (Translator, error) {
	var translator Translator
	var status string = "pending"
	if autoApprove {
		status = "approved"
	}

	var fetchQuery string = `
		SELECT id, name, status, deleted_at, created_at, updated_at
		FROM contributors
		WHERE name = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	var err error = r.db.QueryRow(ctx, fetchQuery, name).Scan(
		&translator.ID,
		&translator.Name,
		&translator.Status,
		&translator.DeletedAt,
		&translator.CreatedAt,
		&translator.UpdatedAt,
	)
	if err == nil {
		return translator, nil
	}

	var insertQuery string = `
		INSERT INTO contributors (name, status)
		VALUES ($1, $2)
		RETURNING id, name, status, deleted_at, created_at, updated_at
	`
	err = r.db.QueryRow(ctx, insertQuery, name, status).Scan(
		&translator.ID,
		&translator.Name,
		&translator.Status,
		&translator.DeletedAt,
		&translator.CreatedAt,
		&translator.UpdatedAt,
	)
	if err != nil {
		return Translator{}, fmt.Errorf("failed to create translator: %w", err)
	}

	return translator, nil
}

func (r *txRepository) LinkEditionTranslator(ctx context.Context, editionID string, translatorID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO edition_contributors (edition_id, contributor_id, role) VALUES ($1, $2, 'translator') ON CONFLICT DO NOTHING`,
		editionID, translatorID,
	)
	if err != nil {
		return fmt.Errorf("failed to link translator: %w", err)
	}
	return nil
}

func (r *txRepository) InsertSubmission(ctx context.Context, userID string, bookID string, editionID string) (Submission, error) {
	var submission Submission
	var query string = `
        INSERT INTO submissions (submitted_by, book_id, edition_id, status, catalogue_only)
        VALUES ($1, $2, $3, 'pending', FALSE)
        RETURNING id, submitted_by, status, catalogue_only, book_id, edition_id, copy_id, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, userID, bookID, editionID).Scan(
		&submission.ID,
		&submission.SubmittedBy,
		&submission.Status,
		&submission.CatalogueOnly,
		&submission.BookID,
		&submission.EditionID,
		&submission.CopyID,
		&submission.DeletedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		return Submission{}, fmt.Errorf("failed to insert submission: %w", err)
	}

	return submission, nil
}

func (r *txRepository) InsertSubmissionApprovedNoCopy(ctx context.Context, userID string, bookID string, editionID string) (Submission, error) {
	var submission Submission
	var query string = `
        INSERT INTO submissions (submitted_by, book_id, edition_id, status, catalogue_only)
        VALUES ($1, $2, $3, 'approved', TRUE)
        RETURNING id, submitted_by, status, catalogue_only, book_id, edition_id, copy_id, deleted_at, created_at, updated_at
    `
	var err error = r.db.QueryRow(ctx, query, userID, bookID, editionID).Scan(
		&submission.ID,
		&submission.SubmittedBy,
		&submission.Status,
		&submission.CatalogueOnly,
		&submission.BookID,
		&submission.EditionID,
		&submission.CopyID,
		&submission.DeletedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		return Submission{}, fmt.Errorf("failed to insert approved submission (no copy): %w", err)
	}
	return submission, nil
}

func (r *txRepository) InsertSubmissionApproved(ctx context.Context, userID string, bookID string, editionID string, copyID string) (Submission, error) {
	var submission Submission
	var query string = `
        INSERT INTO submissions (submitted_by, book_id, edition_id, copy_id, status, catalogue_only)
        VALUES ($1, $2, $3, $4, 'approved', FALSE)
        RETURNING id, submitted_by, status, catalogue_only, book_id, edition_id, copy_id, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, userID, bookID, editionID, copyID).Scan(
		&submission.ID,
		&submission.SubmittedBy,
		&submission.Status,
		&submission.CatalogueOnly,
		&submission.BookID,
		&submission.EditionID,
		&submission.CopyID,
		&submission.DeletedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		return Submission{}, fmt.Errorf("failed to insert approved submission: %w", err)
	}

	return submission, nil
}

func (r *Repository) ListApprovedBooks(ctx context.Context, page int, limit int, sort string) ([]Book, int, error) {
	var offset int = (page - 1) * limit

	var countQuery string = `
        SELECT COUNT(*) FROM books
        WHERE status = 'approved' AND deleted_at IS NULL
    `
	var total int
	var err error = r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count books: %w", err)
	}

	var orderClause string
	switch sort {
	case "newest":
		orderClause = "created_at DESC"
	case "oldest":
		orderClause = "created_at ASC"
	default:
		orderClause = "title ASC"
	}

	var query string = `
        SELECT b.id, b.title, b.status, b.deleted_at, b.created_at, b.updated_at,
               (SELECT be.cover_url FROM book_editions be
                WHERE be.book_id = b.id AND be.deleted_at IS NULL
                ORDER BY be.published_at DESC NULLS LAST, be.created_at ASC
                LIMIT 1) AS cover_url
        FROM books b
        WHERE b.status = 'approved' AND b.deleted_at IS NULL
        ORDER BY ` + orderClause + `
        LIMIT $1 OFFSET $2
    `

	var rows pgx.Rows
	rows, err = r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list books: %w", err)
	}
	defer rows.Close()

	var bookList []Book
	for rows.Next() {
		var b Book
		var coverURL *string
		err = rows.Scan(
			&b.ID, &b.Title,
			&b.Status, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt,
			&coverURL,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan book: %w", err)
		}
		// Surface cover on a synthetic first edition so the frontend can read it
		// as book.editions[0].cover_url without requiring a full details load.
		if coverURL != nil {
			b.Editions = []Edition{{CoverURL: coverURL}}
		}
		b.Authors = []Author{}
		b.Genres = []Genre{}
		bookList = append(bookList, b)
	}
	rows.Close()

	// Batch-load authors and genres for all returned books
	if len(bookList) > 0 {
		ids := make([]string, len(bookList))
		idx := make(map[string]int, len(bookList))
		for i, b := range bookList {
			ids[i] = b.ID
			idx[b.ID] = i
		}

		// Authors
		aRows, aErr := r.db.Query(ctx, `
			SELECT bc.book_id, a.id, a.name, a.status, a.deleted_at, a.created_at, a.updated_at
			FROM contributors a
			JOIN book_contributors bc ON bc.contributor_id = a.id
			WHERE bc.book_id = ANY($1) AND bc.role IN ('author', 'co_author') AND a.deleted_at IS NULL
		`, ids)
		if aErr == nil {
			for aRows.Next() {
				var bookID string
				var a Author
				if sErr := aRows.Scan(&bookID, &a.ID, &a.Name, &a.Status, &a.DeletedAt, &a.CreatedAt, &a.UpdatedAt); sErr == nil {
					if i, ok := idx[bookID]; ok {
						bookList[i].Authors = append(bookList[i].Authors, a)
					}
				}
			}
			aRows.Close()
		}

		// Genres
		gRows, gErr := r.db.Query(ctx, `
			SELECT bg.book_id, g.id, g.name, g.status, g.created_at
			FROM genres g
			JOIN book_genres bg ON bg.genre_id = g.id
			WHERE bg.book_id = ANY($1)
		`, ids)
		if gErr == nil {
			for gRows.Next() {
				var bookID string
				var g Genre
				if sErr := gRows.Scan(&bookID, &g.ID, &g.Name, &g.Status, &g.CreatedAt); sErr == nil {
					if i, ok := idx[bookID]; ok {
						bookList[i].Genres = append(bookList[i].Genres, g)
					}
				}
			}
			gRows.Close()
		}
	}

	return bookList, total, nil
}

func (r *Repository) FindEditionByID(ctx context.Context, id string) (*Edition, error) {
	var edition Edition
	var query string = `
		SELECT id, book_id, format, description, cover_url, isbn10, isbn13, asin, language, publisher,
		       edition, published_at, page_count, file_format,
		       duration_minutes, audio_format, status, deleted_at, created_at, updated_at
		FROM book_editions
		WHERE id = $1 AND deleted_at IS NULL
	`
	var err error = r.db.QueryRow(ctx, query, id).Scan(
		&edition.ID, &edition.BookID, &edition.Format, &edition.Description, &edition.CoverURL,
		&edition.ISBN10, &edition.ISBN13, &edition.ASIN, &edition.Language,
		&edition.Publisher, &edition.Edition, &edition.PublishedAt,
		&edition.PageCount, &edition.FileFormat, &edition.DurationMinutes,
		&edition.AudioFormat, &edition.Status, &edition.DeletedAt,
		&edition.CreatedAt, &edition.UpdatedAt,
	)
	if err != nil {
		return nil, nil // not found is not an error
	}
	return &edition, nil
}

func (r *Repository) FindBookWithDetails(ctx context.Context, id string) (*Book, error) {
	// Get book
	var book Book
	var query string = `
        SELECT id, title, description, series_id, series_position, status, deleted_at, created_at, updated_at
FROM books WHERE id = $1 AND deleted_at IS NULL
    `
	var err error = r.db.QueryRow(ctx, query, id).Scan(
		&book.ID, &book.Title, &book.Description, &book.SeriesID, &book.SeriesPosition,
		&book.Status, &book.DeletedAt, &book.CreatedAt, &book.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	// Load series if present
	if book.SeriesID != nil {
		var series Series
		var seriesQuery string = `
            SELECT id, name, description, status, deleted_at, created_at, updated_at
            FROM series
            WHERE id = $1 AND deleted_at IS NULL
        `
		err = r.db.QueryRow(ctx, seriesQuery, *book.SeriesID).Scan(
			&series.ID,
			&series.Name,
			&series.Description,
			&series.Status,
			&series.DeletedAt,
			&series.CreatedAt,
			&series.UpdatedAt,
		)
		if err == nil {
			book.Series = &series
		}
	}

	// Get authors
	var authorQuery string = `
        SELECT a.id, a.name, a.status, a.deleted_at, a.created_at, a.updated_at
		FROM contributors a
		JOIN book_contributors bc ON bc.contributor_id = a.id
		WHERE bc.book_id = $1
 		 AND bc.role IN ('author', 'co_author')
  		 AND a.deleted_at IS NULL
    `
	var authorRows pgx.Rows
	authorRows, err = r.db.Query(ctx, authorQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get authors: %w", err)
	}
	defer authorRows.Close()

	for authorRows.Next() {
		var a Author
		err = authorRows.Scan(
			&a.ID, &a.Name, &a.Status,
			&a.DeletedAt, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan author: %w", err)
		}
		book.Authors = append(book.Authors, a)
	}

	// Get genres
	var genreQuery string = `
        SELECT g.id, g.name, g.status, g.created_at
        FROM genres g
        JOIN book_genres bg ON bg.genre_id = g.id
        WHERE bg.book_id = $1
    `
	var genreRows pgx.Rows
	genreRows, err = r.db.Query(ctx, genreQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres: %w", err)
	}
	defer genreRows.Close()

	for genreRows.Next() {
		var g Genre
		err = genreRows.Scan(&g.ID, &g.Name, &g.Status, &g.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan genre: %w", err)
		}
		book.Genres = append(book.Genres, g)
	}

	// Get editions
	var editionQuery string = `
        SELECT id, book_id, format, description, cover_url, isbn10, isbn13, asin, language, publisher,
               edition, published_at, page_count, file_format,
               duration_minutes, audio_format, status, deleted_at, created_at, updated_at
        FROM book_editions
        WHERE book_id = $1 AND deleted_at IS NULL
        ORDER BY published_at DESC NULLS LAST, created_at ASC
    `
	var editionRows pgx.Rows
	editionRows, err = r.db.Query(ctx, editionQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get editions: %w", err)
	}
	defer editionRows.Close()

	for editionRows.Next() {
		var e Edition
		err = editionRows.Scan(
			&e.ID, &e.BookID, &e.Format, &e.Description, &e.CoverURL, &e.ISBN10, &e.ISBN13, &e.ASIN,
			&e.Language, &e.Publisher, &e.Edition, &e.PublishedAt,
			&e.PageCount, &e.FileFormat, &e.DurationMinutes,
			&e.AudioFormat, &e.Status, &e.DeletedAt, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edition: %w", err)
		}

		// Load narrators for this edition
		narratorRows, nErr := r.db.Query(ctx, `SELECT c.id, c.name, c.status, c.deleted_at, c.created_at, c.updated_at
			FROM contributors c
			JOIN edition_contributors ec ON ec.contributor_id = c.id
			WHERE ec.edition_id = $1 AND ec.role = 'narrator' AND c.deleted_at IS NULL`, e.ID)
		if nErr == nil {
			for narratorRows.Next() {
				var n Narrator
				if scanErr := narratorRows.Scan(
					&n.ID, &n.Name, &n.Status,
					&n.DeletedAt, &n.CreatedAt, &n.UpdatedAt,
				); scanErr == nil {
					e.Narrators = append(e.Narrators, n)
				}
			}
			narratorRows.Close()
		}

		// Load translators for this edition
		translatorRows, tErr := r.db.Query(ctx, `
			SELECT c.id, c.name, c.status, c.deleted_at, c.created_at, c.updated_at
			FROM contributors c
			JOIN edition_contributors ec ON ec.contributor_id = c.id
			WHERE ec.edition_id = $1 AND ec.role = 'translator' AND c.deleted_at IS NULL
		`, e.ID)
		if tErr == nil {
			for translatorRows.Next() {
				var t Translator
				if scanErr := translatorRows.Scan(
					&t.ID, &t.Name, &t.Status,
					&t.DeletedAt, &t.CreatedAt, &t.UpdatedAt,
				); scanErr == nil {
					e.Translators = append(e.Translators, t)
				}
			}
			translatorRows.Close()
		}

		book.Editions = append(book.Editions, e)
	}

	if book.Authors == nil {
		book.Authors = []Author{}
	}
	if book.Genres == nil {
		book.Genres = []Genre{}
	}
	if book.Editions == nil {
		book.Editions = []Edition{}
	}

	return &book, nil
}

func (r *Repository) UpdateBook(ctx context.Context, id string, title *string, description *string) error {
	var query string = `
        UPDATE books
        SET title       = COALESCE($2, title),
            description = COALESCE($3, description),
            updated_at  = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	var tag pgconn.CommandTag
	var err error
	tag, err = r.db.Exec(ctx, query, id, title, description)
	if err != nil {
		return fmt.Errorf("failed to update book: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("book not found")
	}
	return nil
}

func (r *Repository) DeleteBook(ctx context.Context, id string) error {
	// Only delete if no active copies exist
	var countQuery string = `
        SELECT COUNT(*) FROM book_copies bc
        JOIN book_editions be ON be.id = bc.edition_id
        WHERE be.book_id = $1 AND bc.deleted_at IS NULL
    `
	var count int
	var err error = r.db.QueryRow(ctx, countQuery, id).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check copies: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot delete book with existing copies")
	}

	var query string = `
        UPDATE books SET deleted_at = NOW(), updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	var tag pgconn.CommandTag
	tag, err = r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("book not found")
	}
	return nil
}

// ForceDeleteBook soft-deletes a book and all its copies/editions regardless of ownership.
func (r *Repository) ForceDeleteBook(ctx context.Context, id string) error {
	// Soft-delete all copies belonging to this book's editions
	if _, err := r.db.Exec(ctx, `
        UPDATE book_copies SET deleted_at = NOW(), updated_at = NOW()
        WHERE edition_id IN (
            SELECT id FROM book_editions WHERE book_id = $1
        ) AND deleted_at IS NULL`, id); err != nil {
		return fmt.Errorf("failed to delete copies: %w", err)
	}
	// Soft-delete all editions
	if _, err := r.db.Exec(ctx, `
        UPDATE book_editions SET deleted_at = NOW(), updated_at = NOW()
        WHERE book_id = $1 AND deleted_at IS NULL`, id); err != nil {
		return fmt.Errorf("failed to delete editions: %w", err)
	}
	// Soft-delete the book itself
	tag, err := r.db.Exec(ctx, `
        UPDATE books SET deleted_at = NOW(), updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("book not found")
	}
	return nil
}

// DeleteEdition soft-deletes a single edition and all its copies.
func (r *Repository) DeleteEdition(ctx context.Context, editionID string) error {
	if _, err := r.db.Exec(ctx, `
        UPDATE book_copies SET deleted_at = NOW(), updated_at = NOW()
        WHERE edition_id = $1 AND deleted_at IS NULL`, editionID); err != nil {
		return fmt.Errorf("failed to delete copies: %w", err)
	}
	tag, err := r.db.Exec(ctx, `
        UPDATE book_editions SET deleted_at = NOW(), updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL`, editionID)
	if err != nil {
		return fmt.Errorf("failed to delete edition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("edition not found")
	}
	return nil
}

func (r *txRepository) FindOrCreateSeries(ctx context.Context, name string, autoApprove bool) (Series, error) {
	var series Series
	status := "pending"
	if autoApprove {
		status = "approved"
	}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, status, deleted_at, created_at, updated_at
         FROM series WHERE name = $1 AND deleted_at IS NULL LIMIT 1`, name,
	).Scan(&series.ID, &series.Name, &series.Status, &series.DeletedAt, &series.CreatedAt, &series.UpdatedAt)
	if err == nil {
		return series, nil
	}
	err = r.db.QueryRow(ctx,
		`INSERT INTO series (name, status) VALUES ($1, $2)
         RETURNING id, name, status, deleted_at, created_at, updated_at`,
		name, status,
	).Scan(&series.ID, &series.Name, &series.Status, &series.DeletedAt, &series.CreatedAt, &series.UpdatedAt)
	if err != nil {
		return Series{}, fmt.Errorf("failed to create series: %w", err)
	}
	return series, nil
}

func (r *txRepository) UpdateBookSeries(ctx context.Context, bookID, seriesID string, position *float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE books SET series_id = $2, series_position = $3, updated_at = NOW() WHERE id = $1`,
		bookID, seriesID, position,
	)
	return err
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func (r *Repository) FindBooksWithoutCovers(ctx context.Context) ([]BookWithDetails, error) {
	query := `
        SELECT DISTINCT b.id, b.title
        FROM books b
        JOIN book_editions e ON e.book_id = b.id
        WHERE (e.cover_url IS NULL OR e.cover_url = '')
          AND ((e.isbn13 IS NOT NULL AND e.isbn13 != '') OR (e.isbn10 IS NOT NULL AND e.isbn10 != ''))
          AND b.deleted_at IS NULL
          AND e.deleted_at IS NULL
    `
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []BookWithDetails
	for rows.Next() {
		var b BookWithDetails
		if err := rows.Scan(&b.ID, &b.Title); err != nil {
			continue
		}
		// Load editions for this book
		b.Editions, _ = r.findEditionsByBookID(ctx, b.ID)
		books = append(books, b)
	}
	return books, nil
}

func (r *Repository) UpdateEditionCoverURL(ctx context.Context, editionID, coverURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE book_editions SET cover_url = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`,
		coverURL, editionID,
	)
	return err
}

func (r *Repository) findEditionsByBookID(ctx context.Context, bookID string) ([]Edition, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, book_id, isbn10, isbn13, format, language
         FROM book_editions WHERE book_id = $1 AND deleted_at IS NULL`,
		bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var editions []Edition
	for rows.Next() {
		var e Edition
		if err := rows.Scan(&e.ID, &e.BookID, &e.ISBN10, &e.ISBN13, &e.Format, &e.Language); err != nil {
			continue
		}
		editions = append(editions, e)
	}
	return editions, nil
}

// nullableString returns the string value or nil if the pointer is nil or empty.
func nullableString(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
