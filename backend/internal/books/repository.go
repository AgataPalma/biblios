package books

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

// withDB returns a new repository using the given DB (pool or tx)
func (r *Repository) withDB(db DB) *txRepository {
	return &txRepository{db: db}
}

// txRepository works with either a pool or a transaction
type txRepository struct {
	db DB
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
