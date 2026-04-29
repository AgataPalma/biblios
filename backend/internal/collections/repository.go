package collections

import (
	"context"
	"fmt"

	"github.com/AgataPalma/biblios/internal/books"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const collectionColumns = `id, library_id, created_by, name, description, cover_colour,
	is_public, is_collaborative, deleted_at, created_at, updated_at`

func scanCollection(row pgx.Row, c *Collection) error {
	return row.Scan(
		&c.ID, &c.LibraryID, &c.CreatedBy, &c.Name, &c.Description, &c.CoverColour,
		&c.IsPublic, &c.IsCollaborative, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt,
	)
}

func (r *Repository) Create(ctx context.Context, libraryID, createdBy, name string, description, coverColour *string, isPublic, isCollaborative bool) (Collection, error) {
	var c Collection
	err := scanCollection(r.db.QueryRow(ctx, `
		INSERT INTO collections (library_id, created_by, name, description, cover_colour, is_public, is_collaborative)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING `+collectionColumns,
		libraryID, createdBy, name, description, coverColour, isPublic, isCollaborative), &c)
	if err != nil {
		return Collection{}, fmt.Errorf("create collection: %w", err)
	}
	return c, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Collection, error) {
	var c Collection
	err := scanCollection(r.db.QueryRow(ctx, `
		SELECT `+collectionColumns+` FROM collections WHERE id=$1 AND deleted_at IS NULL`, id), &c)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find collection: %w", err)
	}
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM collection_books WHERE collection_id=$1`, id).Scan(&c.BookCount)
	return &c, nil
}

func (r *Repository) ListByLibrary(ctx context.Context, libraryID string) ([]Collection, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+collectionColumns+` FROM collections
		WHERE library_id=$1 AND deleted_at IS NULL
		ORDER BY name ASC`, libraryID)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()

	var cols []Collection
	for rows.Next() {
		var c Collection
		if err := scanCollection(rows, &c); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load book counts
	for i, col := range cols {
		r.db.QueryRow(ctx, `SELECT COUNT(*) FROM collection_books WHERE collection_id=$1`, col.ID).Scan(&cols[i].BookCount)
	}
	return cols, nil
}

func (r *Repository) Update(ctx context.Context, id string, name, description *string, isPublic *bool) (Collection, error) {
	var c Collection
	err := scanCollection(r.db.QueryRow(ctx, `
		UPDATE collections SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			is_public   = COALESCE($4, is_public),
			updated_at  = NOW()
		WHERE id=$1 AND deleted_at IS NULL
		RETURNING `+collectionColumns, id, name, description, isPublic), &c)
	if err == pgx.ErrNoRows {
		return Collection{}, fmt.Errorf("collection not found")
	}
	if err != nil {
		return Collection{}, fmt.Errorf("update collection: %w", err)
	}
	return c, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `UPDATE collections SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("collection not found")
	}
	return nil
}

func (r *Repository) AddBook(ctx context.Context, collectionID, copyID, addedBy string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO collection_books (collection_id, book_copy_id, added_by)
		VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		collectionID, copyID, addedBy)
	if err != nil {
		return fmt.Errorf("add book to collection: %w", err)
	}
	return nil
}

func (r *Repository) RemoveBook(ctx context.Context, collectionID, copyID string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM collection_books WHERE collection_id=$1 AND book_copy_id=$2`,
		collectionID, copyID)
	if err != nil {
		return fmt.Errorf("remove book from collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("book not found in collection")
	}
	return nil
}

func (r *Repository) ListBooks(ctx context.Context, collectionID string, page, limit int) ([]books.UserBook, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM collection_books WHERE collection_id=$1`, collectionID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count collection books: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT bc.id, bc.reading_status, bc.current_page, bc.started_reading_at, bc.finished_reading_at,
		       bc.owned_by_user, bc.borrowed_from, bc.location, bc.condition, bc.reread_count,
		       bc.personal_notes, cb.added_at,
		       be.id, be.format, be.language, be.cover_url,
		       b.id, b.title, b.description, b.series_id, b.series_position, b.status,
		       b.deleted_at, b.created_at, b.updated_at
		FROM collection_books cb
		JOIN book_copies bc ON bc.id=cb.book_copy_id
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN books b ON b.id=be.book_id
		WHERE cb.collection_id=$1 AND bc.deleted_at IS NULL
		ORDER BY cb.added_at DESC
		LIMIT $2 OFFSET $3`, collectionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list collection books: %w", err)
	}
	defer rows.Close()

	var userBooks []books.UserBook
	for rows.Next() {
		var ub books.UserBook
		if err := rows.Scan(
			&ub.CopyID, &ub.ReadingStatus, &ub.CurrentPage, &ub.StartedReadingAt, &ub.FinishedReadingAt,
			&ub.OwnedByUser, &ub.BorrowedFrom, &ub.Location, &ub.Condition, &ub.RereadCount,
			&ub.PersonalNotes, &ub.AddedAt,
			&ub.EditionID, &ub.Format, &ub.Language, &ub.CoverURL,
			&ub.Book.ID, &ub.Book.Title, &ub.Book.Description, &ub.Book.SeriesID, &ub.Book.SeriesPosition,
			&ub.Book.Status, &ub.Book.DeletedAt, &ub.Book.CreatedAt, &ub.Book.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan collection book: %w", err)
		}
		ub.Book.Authors = []books.Contributor{}
		ub.Book.Genres = []books.Genre{}
		userBooks = append(userBooks, ub)
	}
	return userBooks, total, rows.Err()
}
