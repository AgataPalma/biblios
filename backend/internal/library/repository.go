package library

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertCopy(ctx context.Context, editionID, ownerID string, condition *string, opts CopyOptions) (Copy, error) {
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
        INSERT INTO book_copies (
            edition_id, owner_id, condition,
            reading_status, current_page,
            started_reading_at, finished_reading_at,
            owned_by_user, borrowed_from, location
        )
        VALUES (
            $1, $2, $3,
            $4, $5,
            CASE WHEN $6::text = '' THEN NULL ELSE $6::timestamptz END,
            CASE WHEN $7::text = '' THEN NULL ELSE $7::timestamptz END,
            $8, $9, $10
        )
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

func (r *Repository) UpdateReadingStatus(ctx context.Context, copyID, userID string, input UpdateCopyInput) error {
	var query string = `
        UPDATE book_copies SET
            reading_status      = $1,
            current_page        = $2,
            started_reading_at  = CASE WHEN $3::text = '' THEN NULL ELSE $3::timestamptz END,
            finished_reading_at = CASE WHEN $4::text = '' THEN NULL ELSE $4::timestamptz END,
            owned_by_user       = COALESCE($5, owned_by_user),
            borrowed_from       = $6,
            location            = COALESCE($7, location),
            updated_at          = NOW()
        WHERE id = $8 AND owner_id = $9 AND deleted_at IS NULL
    `
	tag, err := r.db.Exec(ctx, query,
		input.Status,
		input.CurrentPage,
		nullableString(input.StartedReadingAt),
		nullableString(input.FinishedReadingAt),
		input.OwnedByUser,
		input.BorrowedFrom,
		input.Location,
		copyID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update copy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("copy not found")
	}
	return nil
}

func (r *Repository) RemoveCopy(ctx context.Context, copyID, userID string) error {
	var query string = `
        UPDATE book_copies SET deleted_at = NOW(), updated_at = NOW()
        WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
    `
	tag, err := r.db.Exec(ctx, query, copyID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove copy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("copy not found")
	}
	return nil
}

func (r *Repository) FindUserLibrary(ctx context.Context, userID string, page, limit int) ([]UserBook, int, error) {
	var offset int = (page - 1) * limit

	var countQuery string = `
        SELECT COUNT(DISTINCT bc.id)
        FROM book_copies bc
        JOIN book_editions be ON be.id = bc.edition_id
        JOIN books b ON b.id = be.book_id
        WHERE bc.owner_id = $1 AND bc.deleted_at IS NULL AND b.deleted_at IS NULL
    `
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count: %w", err)
	}

	var query string = `
        SELECT
            bc.id,
            bc.reading_status,
            bc.current_page,
            bc.started_reading_at,
            bc.finished_reading_at,
            bc.owned_by_user,
            bc.borrowed_from,
            bc.location,
            bc.condition,
            bc.created_at,

            b.id,
            b.title,
            b.status,
            b.deleted_at,
            b.created_at,
            b.updated_at,

            be.id,
            be.format,
            be.language,
            be.cover_url,

            c.id,
            c.name
        FROM book_copies bc
        JOIN book_editions be ON be.id = bc.edition_id
        JOIN books b ON b.id = be.book_id
        LEFT JOIN book_contributors bc2 ON bc2.book_id = b.id AND bc2.role IN ('author', 'co_author')
        LEFT JOIN contributors c ON c.id = bc2.contributor_id
        WHERE bc.owner_id = $1
          AND bc.deleted_at IS NULL
          AND b.deleted_at IS NULL
        ORDER BY bc.created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	type rowData struct {
		userBook   UserBook
		authorID   *string
		authorName *string
	}

	bookMap := make(map[string]*UserBook)
	order := make([]string, 0)

	for rows.Next() {
		var rd rowData
		err = rows.Scan(
			&rd.userBook.CopyID,
			&rd.userBook.ReadingStatus,
			&rd.userBook.CurrentPage,
			&rd.userBook.StartedReadingAt,
			&rd.userBook.FinishedReadingAt,
			&rd.userBook.OwnedByUser,
			&rd.userBook.BorrowedFrom,
			&rd.userBook.Location,
			&rd.userBook.Condition,
			&rd.userBook.AddedAt,

			&rd.userBook.Book.ID,
			&rd.userBook.Book.Title,
			&rd.userBook.Book.Status,
			&rd.userBook.Book.DeletedAt,
			&rd.userBook.Book.CreatedAt,
			&rd.userBook.Book.UpdatedAt,

			&rd.userBook.Edition.ID,
			&rd.userBook.Edition.Format,
			&rd.userBook.Edition.Language,
			&rd.userBook.Edition.CoverURL,

			&rd.authorID,
			&rd.authorName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan: %w", err)
		}

		existing, ok := bookMap[rd.userBook.CopyID]
		if !ok {
			rd.userBook.Book.Authors = []Author{}
			copy := rd.userBook
			bookMap[rd.userBook.CopyID] = &copy
			order = append(order, rd.userBook.CopyID)
			existing = &copy
		}

		if rd.authorID != nil && rd.authorName != nil {
			existing.Book.Authors = append(existing.Book.Authors, Author{
				ID:   *rd.authorID,
				Name: *rd.authorName,
			})
		}
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("row iteration failed: %w", rows.Err())
	}

	userBooks := make([]UserBook, 0, len(order))
	for _, id := range order {
		userBooks = append(userBooks, *bookMap[id])
	}

	return userBooks, total, nil
}

func nullableString(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
