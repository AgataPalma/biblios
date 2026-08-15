package collections

import (
	"context"
	"fmt"

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

func (r *Repository) GetLibraryAccess(ctx context.Context, libraryID, userID string) (*LibraryAccess, error) {
	var access LibraryAccess
	err := r.db.QueryRow(ctx, `
		SELECT l.visibility,
		       lm.user_id IS NOT NULL,
		       COALESCE(lm.is_owner, false),
		       COALESCE(lm.can_view, false),
		       COALESCE(lm.can_add, false),
		       COALESCE(lm.can_remove, false),
		       COALESCE(lm.can_edit, false)
		FROM libraries l
		LEFT JOIN library_members lm
		       ON lm.library_id=l.id AND lm.user_id=$2
		WHERE l.id=$1 AND l.deleted_at IS NULL`, libraryID, userID).Scan(
		&access.Visibility,
		&access.IsMember,
		&access.IsOwner,
		&access.CanView,
		&access.CanAdd,
		&access.CanRemove,
		&access.CanEdit,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get library access: %w", err)
	}
	return &access, nil
}

func (r *Repository) FindByID(ctx context.Context, libraryID, id string) (*Collection, error) {
	var c Collection
	err := scanCollection(r.db.QueryRow(ctx, `
			SELECT `+collectionColumns+` FROM collections
			WHERE id=$1 AND library_id=$2 AND deleted_at IS NULL`, id, libraryID), &c)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find collection: %w", err)
	}
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM collection_books cb
		JOIN book_copies bc ON bc.id=cb.book_copy_id
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN books b ON b.id=be.book_id
		WHERE cb.collection_id=$1
		  AND bc.deleted_at IS NULL
		  AND be.deleted_at IS NULL AND be.status='approved'
		  AND b.deleted_at IS NULL AND b.status='approved'`, id).Scan(&c.BookCount); err != nil {
		return nil, fmt.Errorf("count collection books: %w", err)
	}
	return &c, nil
}

func (r *Repository) ListByLibrary(ctx context.Context, libraryID string, publicOnly bool) ([]Collection, error) {
	rows, err := r.db.Query(ctx, `
			SELECT c.id, c.library_id, c.created_by, c.name, c.description, c.cover_colour,
			       c.is_public, c.is_collaborative, c.deleted_at, c.created_at, c.updated_at,
		       COUNT(cb.book_copy_id) FILTER (
		           WHERE EXISTS (
		               SELECT 1
		               FROM book_copies bc
		               JOIN book_editions be ON be.id=bc.edition_id
		               JOIN books b ON b.id=be.book_id
		               WHERE bc.id=cb.book_copy_id
		                 AND bc.deleted_at IS NULL
		                 AND be.deleted_at IS NULL AND be.status='approved'
		                 AND b.deleted_at IS NULL AND b.status='approved'
		           )
		       )
			FROM collections c
			LEFT JOIN collection_books cb ON cb.collection_id=c.id
			WHERE c.library_id=$1 AND c.deleted_at IS NULL
			  AND (NOT $2 OR c.is_public)
			GROUP BY c.id
			ORDER BY c.name ASC`, libraryID, publicOnly)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()

	var cols []Collection
	for rows.Next() {
		var c Collection
		if err := rows.Scan(
			&c.ID, &c.LibraryID, &c.CreatedBy, &c.Name, &c.Description, &c.CoverColour,
			&c.IsPublic, &c.IsCollaborative, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt,
			&c.BookCount,
		); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
		return Collection{}, ErrCollectionNotFound
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
		return ErrCollectionNotFound
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

func (r *Repository) CopyBelongsToLibrary(ctx context.Context, libraryID, copyID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM library_book_copies lbc
			JOIN book_copies bc ON bc.id=lbc.book_copy_id
			WHERE lbc.library_id=$1
			  AND lbc.book_copy_id=$2
			  AND bc.deleted_at IS NULL
		)`, libraryID, copyID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check copy library: %w", err)
	}
	return exists, nil
}

func (r *Repository) RemoveBook(ctx context.Context, collectionID, copyID string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM collection_books WHERE collection_id=$1 AND book_copy_id=$2`,
		collectionID, copyID)
	if err != nil {
		return fmt.Errorf("remove book from collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBookNotInCollection
	}
	return nil
}

func (r *Repository) ListBooks(ctx context.Context, collectionID string, page, limit int) ([]CollectionBook, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM collection_books cb
		JOIN book_copies bc ON bc.id=cb.book_copy_id
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN books b ON b.id=be.book_id
		WHERE cb.collection_id=$1
		  AND bc.deleted_at IS NULL
		  AND be.deleted_at IS NULL AND be.status='approved'
		  AND b.deleted_at IS NULL AND b.status='approved'`, collectionID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count collection books: %w", err)
	}

	rows, err := r.db.Query(ctx, `
			SELECT cb.collection_id, cb.book_copy_id, b.id, be.id, b.title,
			       ARRAY(
			           SELECT c.name
			           FROM book_contributors bco
			           JOIN contributors c ON c.id=bco.contributor_id
			           WHERE bco.book_id=b.id
			             AND bco.role IN ('author', 'co_author')
			             AND c.deleted_at IS NULL
			             AND c.status='approved'
			           ORDER BY c.name
			       ),
			       be.format, be.language, be.cover_url, cb.added_at
			FROM collection_books cb
			JOIN book_copies bc ON bc.id=cb.book_copy_id
			JOIN book_editions be ON be.id=bc.edition_id
			JOIN books b ON b.id=be.book_id
			WHERE cb.collection_id=$1
			  AND bc.deleted_at IS NULL
			  AND be.deleted_at IS NULL
			  AND be.status='approved'
			  AND b.deleted_at IS NULL
			  AND b.status='approved'
			ORDER BY cb.added_at DESC
			LIMIT $2 OFFSET $3`, collectionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list collection books: %w", err)
	}
	defer rows.Close()

	var collectionBooks []CollectionBook
	for rows.Next() {
		var book CollectionBook
		if err := rows.Scan(
			&book.CollectionID, &book.BookCopyID, &book.BookID, &book.EditionID, &book.Title,
			&book.Authors, &book.Format, &book.Language, &book.CoverURL, &book.AddedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan collection book: %w", err)
		}
		collectionBooks = append(collectionBooks, book)
	}
	return collectionBooks, total, rows.Err()
}
