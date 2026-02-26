package books

import (
	"context"
	"fmt"

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

func (r *Repository) InsertCopyDirect(ctx context.Context, editionID string, ownerID string) (Copy, error) {
	var copy Copy
	var query string = `
        INSERT INTO book_copies (edition_id, owner_id)
        VALUES ($1, $2)
        RETURNING id, edition_id, owner_id, condition, deleted_at, created_at, updated_at
    `
	var err error = r.db.QueryRow(ctx, query, editionID, ownerID).Scan(
		&copy.ID, &copy.EditionID, &copy.OwnerID, &copy.Condition,
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
        SELECT id, book_id, format, isbn, asin, language, publisher,
               edition, published_at, page_count, file_format,
               duration_minutes, audio_format, status, deleted_at, created_at, updated_at
        FROM book_editions
        WHERE isbn = $1 AND deleted_at IS NULL
        LIMIT 1
    `

	var err error = r.db.QueryRow(ctx, query, isbn).Scan(
		&edition.ID, &edition.BookID, &edition.Format,
		&edition.ISBN, &edition.ASIN, &edition.Language,
		&edition.Publisher, &edition.Edition, &edition.PublishedAt,
		&edition.PageCount, &edition.FileFormat, &edition.DurationMinutes,
		&edition.AudioFormat, &edition.Status, &edition.DeletedAt,
		&edition.CreatedAt, &edition.UpdatedAt,
	)
	if err != nil {
		// Not found is not an error here
		return nil, nil
	}

	return &edition, nil
}

func (r *Repository) FindBookByID(ctx context.Context, bookID string) (*Book, error) {
	var book Book
	var query string = `
        SELECT id, title, description, cover_url, status, deleted_at, created_at, updated_at
        FROM books
        WHERE id = $1 AND deleted_at IS NULL
    `

	var err error = r.db.QueryRow(ctx, query, bookID).Scan(
		&book.ID, &book.Title, &book.Description,
		&book.CoverURL, &book.Status, &book.DeletedAt,
		&book.CreatedAt, &book.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	return &book, nil
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
        SELECT id, submitted_by, status, rejection_reason, reviewed_by,
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
			&s.ID, &s.SubmittedBy, &s.Status, &s.RejectionReason,
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
        SELECT id, submitted_by, status, rejection_reason, reviewed_by,
               reviewed_at, book_id, edition_id, copy_id, deleted_at, created_at, updated_at
        FROM submissions
        WHERE id = $1 AND deleted_at IS NULL
    `

	var err error = r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.SubmittedBy, &s.Status, &s.RejectionReason,
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
        UPDATE authors SET status = 'approved', updated_at = NOW()
        WHERE id IN (SELECT author_id FROM book_authors WHERE book_id = $1)
        AND status = 'pending'
    `
	_, err = r.db.Exec(ctx, authorQuery, bookID)
	if err != nil {
		return fmt.Errorf("failed to approve authors: %w", err)
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

func (r *Repository) UpdateBookDetails(ctx context.Context, bookID string, title string, description *string, coverURL *string) error {
	var query string = `
        UPDATE books SET title = $2, description = $3, cover_url = $4, updated_at = NOW()
        WHERE id = $1
    `
	var _, err = r.db.Exec(ctx, query, bookID, title, description, coverURL)
	if err != nil {
		return fmt.Errorf("failed to update book: %w", err)
	}
	return nil
}

func (r *Repository) UpdateEditionDetails(ctx context.Context, editionID string, e Edition) error {
	var query string = `
        UPDATE book_editions SET
            format = $2, isbn = $3, language = $4, publisher = $5,
            edition = $6, page_count = $7, updated_at = NOW()
        WHERE id = $1
    `
	var _, err = r.db.Exec(ctx, query,
		editionID, e.Format, e.ISBN, e.Language,
		e.Publisher, e.Edition, e.PageCount,
	)
	if err != nil {
		return fmt.Errorf("failed to update edition: %w", err)
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

	var query string = `
        INSERT INTO authors (name, status)
        VALUES ($1, $2)
        ON CONFLICT (name) DO NOTHING
        RETURNING id, name, status, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, name, status).Scan(
		&author.ID,
		&author.Name,
		&author.Status,
		&author.DeletedAt,
		&author.CreatedAt,
		&author.UpdatedAt,
	)
	if err != nil {
		var fetchQuery string = `
            SELECT id, name, status, deleted_at, created_at, updated_at
            FROM authors WHERE name = $1 AND deleted_at IS NULL
        `
		err = r.db.QueryRow(ctx, fetchQuery, name).Scan(
			&author.ID,
			&author.Name,
			&author.Status,
			&author.DeletedAt,
			&author.CreatedAt,
			&author.UpdatedAt,
		)
		if err != nil {
			return Author{}, fmt.Errorf("failed to find author: %w", err)
		}
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

func (r *txRepository) InsertBook(ctx context.Context, title string, description *string, coverURL *string, autoApprove bool) (Book, error) {
	var book Book
	var status string = "pending"
	if autoApprove {
		status = "approved"
	}

	var query string = `
        INSERT INTO books (title, description, cover_url, status)
        VALUES ($1, $2, $3, $4)
        RETURNING id, title, description, cover_url, status, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, title, description, coverURL, status).Scan(
		&book.ID,
		&book.Title,
		&book.Description,
		&book.CoverURL,
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
            book_id, format, isbn, asin, language, publisher,
            edition, published_at, page_count, file_format,
            duration_minutes, audio_format, status
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
        RETURNING id, book_id, format, isbn, asin, language, publisher,
            edition, published_at, page_count, file_format,
            duration_minutes, audio_format, status, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query,
		e.BookID, e.Format, e.ISBN, e.ASIN, e.Language,
		e.Publisher, e.Edition, e.PublishedAt, e.PageCount,
		e.FileFormat, e.DurationMinutes, e.AudioFormat, status,
	).Scan(
		&edition.ID, &edition.BookID, &edition.Format,
		&edition.ISBN, &edition.ASIN, &edition.Language,
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

func (r *txRepository) InsertCopy(ctx context.Context, editionID string, ownerID string, condition *string) (Copy, error) {
	var copy Copy
	var query string = `
        INSERT INTO book_copies (edition_id, owner_id, condition)
        VALUES ($1, $2, $3)
        RETURNING id, edition_id, owner_id, condition, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, editionID, ownerID, condition).Scan(
		&copy.ID,
		&copy.EditionID,
		&copy.OwnerID,
		&copy.Condition,
		&copy.DeletedAt,
		&copy.CreatedAt,
		&copy.UpdatedAt,
	)
	if err != nil {
		return Copy{}, fmt.Errorf("failed to insert copy: %w", err)
	}

	return copy, nil
}

func (r *txRepository) LinkBookAuthor(ctx context.Context, bookID string, authorID string) error {
	var query string = `
        INSERT INTO book_authors (book_id, author_id)
        VALUES ($1, $2) ON CONFLICT DO NOTHING
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

func (r *txRepository) InsertSubmission(ctx context.Context, userID string, bookID string, editionID string) (Submission, error) {
	var submission Submission
	var query string = `
        INSERT INTO submissions (submitted_by, book_id, edition_id, status)
        VALUES ($1, $2, $3, 'pending')
        RETURNING id, submitted_by, status, book_id, edition_id, copy_id, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, userID, bookID, editionID).Scan(
		&submission.ID,
		&submission.SubmittedBy,
		&submission.Status,
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

func (r *txRepository) InsertSubmissionApproved(ctx context.Context, userID string, bookID string, editionID string, copyID string) (Submission, error) {
	var submission Submission
	var query string = `
        INSERT INTO submissions (submitted_by, book_id, edition_id, copy_id, status)
        VALUES ($1, $2, $3, $4, 'approved')
        RETURNING id, submitted_by, status, book_id, edition_id, copy_id, deleted_at, created_at, updated_at
    `

	var err error = r.db.QueryRow(ctx, query, userID, bookID, editionID, copyID).Scan(
		&submission.ID,
		&submission.SubmittedBy,
		&submission.Status,
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

func (r *Repository) ListApprovedBooks(ctx context.Context, page int, limit int) ([]Book, int, error) {
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

	var query string = `
        SELECT id, title, description, cover_url, status, deleted_at, created_at, updated_at
        FROM books
        WHERE status = 'approved' AND deleted_at IS NULL
        ORDER BY title ASC
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
		err = rows.Scan(
			&b.ID, &b.Title, &b.Description, &b.CoverURL,
			&b.Status, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan book: %w", err)
		}
		bookList = append(bookList, b)
	}

	return bookList, total, nil
}

func (r *Repository) FindBookWithDetails(ctx context.Context, id string) (*Book, error) {
	// Get book
	var book Book
	var query string = `
        SELECT id, title, description, cover_url, status, deleted_at, created_at, updated_at
        FROM books
        WHERE id = $1 AND deleted_at IS NULL
    `
	var err error = r.db.QueryRow(ctx, query, id).Scan(
		&book.ID, &book.Title, &book.Description, &book.CoverURL,
		&book.Status, &book.DeletedAt, &book.CreatedAt, &book.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	// Get authors
	var authorQuery string = `
        SELECT a.id, a.name, a.status, a.deleted_at, a.created_at, a.updated_at
        FROM authors a
        JOIN book_authors ba ON ba.author_id = a.id
        WHERE ba.book_id = $1 AND a.deleted_at IS NULL
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
        SELECT id, book_id, format, isbn, asin, language, publisher,
               edition, published_at, page_count, file_format,
               duration_minutes, audio_format, status, deleted_at, created_at, updated_at
        FROM book_editions
        WHERE book_id = $1 AND deleted_at IS NULL
        ORDER BY created_at ASC
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
			&e.ID, &e.BookID, &e.Format, &e.ISBN, &e.ASIN,
			&e.Language, &e.Publisher, &e.Edition, &e.PublishedAt,
			&e.PageCount, &e.FileFormat, &e.DurationMinutes,
			&e.AudioFormat, &e.Status, &e.DeletedAt,
			&e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edition: %w", err)
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

func (r *Repository) UpdateBook(ctx context.Context, id string, title string, description *string, coverURL *string) error {
	var query string = `
        UPDATE books
        SET title = $2, description = $3, cover_url = $4, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	var tag pgconn.CommandTag
	var err error
	tag, err = r.db.Exec(ctx, query, id, title, description, coverURL)
	if err != nil {
		return fmt.Errorf("failed to update book: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("book not found")
	}
	return nil
}

func (r *Repository) DeleteBook(ctx context.Context, id string) error {
	// Only delete if no approved copies exist
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

func (r *Repository) FindUserBooks(ctx context.Context, userID string, page int, limit int) ([]Book, int, error) {
	var offset int = (page - 1) * limit

	var countQuery string = `
        SELECT COUNT(DISTINCT b.id)
        FROM books b
        JOIN book_editions be ON be.book_id = b.id
        JOIN book_copies bc ON bc.edition_id = be.id
        WHERE bc.owner_id = $1 AND bc.deleted_at IS NULL AND b.deleted_at IS NULL
    `
	var total int
	var err error = r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user books: %w", err)
	}

	var query string = `
        SELECT DISTINCT b.id, b.title, b.description, b.cover_url, b.status,
               b.deleted_at, b.created_at, b.updated_at
        FROM books b
        JOIN book_editions be ON be.book_id = b.id
        JOIN book_copies bc ON bc.edition_id = be.id
        WHERE bc.owner_id = $1 AND bc.deleted_at IS NULL AND b.deleted_at IS NULL
        ORDER BY b.title ASC
        LIMIT $2 OFFSET $3
    `
	var rows pgx.Rows
	rows, err = r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user books: %w", err)
	}
	defer rows.Close()

	var bookList []Book
	for rows.Next() {
		var b Book
		err = rows.Scan(
			&b.ID, &b.Title, &b.Description, &b.CoverURL,
			&b.Status, &b.DeletedAt, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan book: %w", err)
		}
		bookList = append(bookList, b)
	}

	return bookList, total, nil
}
