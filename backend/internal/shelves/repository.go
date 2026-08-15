package shelves

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

func (r *Repository) Create(ctx context.Context, userID, name string) (Shelf, error) {
	var s Shelf
	err := r.db.QueryRow(ctx, `
			INSERT INTO shelves (user_id, name, scope)
			VALUES ($1, $2, 'personal')
			RETURNING id, user_id, name, created_at`,
		userID, name).Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err != nil {
		return Shelf{}, fmt.Errorf("create shelf: %w", err)
	}
	return s, nil
}

func (r *Repository) FindByID(ctx context.Context, id, userID string) (*Shelf, error) {
	var s Shelf
	err := r.db.QueryRow(ctx, `
			SELECT id, user_id, name, created_at
			FROM shelves
			WHERE id=$1 AND user_id=$2 AND scope='personal'`, id, userID).Scan(
		&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find shelf: %w", err)
	}
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM shelf_books sb
		JOIN book_copies bc ON bc.id=sb.copy_id
		WHERE sb.shelf_id=$1 AND bc.owner_id=$2 AND bc.deleted_at IS NULL`, id, userID).Scan(&s.BookCount); err != nil {
		return nil, fmt.Errorf("count shelf books: %w", err)
	}
	return &s, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Shelf, error) {
	rows, err := r.db.Query(ctx, `
			SELECT s.id, s.user_id, s.name, s.created_at,
			       COUNT(sb.copy_id) FILTER (
			           WHERE bc.owner_id=$1 AND bc.deleted_at IS NULL
			       )
			FROM shelves s
			LEFT JOIN shelf_books sb ON sb.shelf_id=s.id
			LEFT JOIN book_copies bc ON bc.id=sb.copy_id
			WHERE s.user_id=$1 AND s.scope='personal'
			GROUP BY s.id
			ORDER BY s.name ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list shelves: %w", err)
	}
	defer rows.Close()

	var shelves []Shelf
	for rows.Next() {
		var s Shelf
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt, &s.BookCount); err != nil {
			return nil, err
		}
		shelves = append(shelves, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shelves, nil
}

func (r *Repository) Rename(ctx context.Context, id, userID, name string) (Shelf, error) {
	var s Shelf
	err := r.db.QueryRow(ctx, `
			UPDATE shelves SET name=$3
			WHERE id=$1 AND user_id=$2 AND scope='personal'
		RETURNING id, user_id, name, created_at`,
		id, userID, name).Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return Shelf{}, ErrShelfNotFound
	}
	if err != nil {
		return Shelf{}, fmt.Errorf("rename shelf: %w", err)
	}
	return s, nil
}

// Delete removes the shelf and cascades shelf_books via FK ON DELETE CASCADE.
func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM shelves WHERE id=$1 AND user_id=$2 AND scope='personal'`, id, userID)
	if err != nil {
		return fmt.Errorf("delete shelf: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrShelfNotFound
	}
	return nil
}

func (r *Repository) AddBook(ctx context.Context, shelfID, userID, copyID string) error {
	tag, err := r.db.Exec(ctx, `
			INSERT INTO shelf_books (shelf_id, copy_id)
			SELECT s.id, bc.id
			FROM shelves s
			JOIN book_copies bc ON bc.id=$3
			WHERE s.id=$1
			  AND s.user_id=$2
			  AND s.scope='personal'
			  AND bc.owner_id=$2
			  AND bc.deleted_at IS NULL
			ON CONFLICT (shelf_id, copy_id) DO UPDATE
			SET copy_id=EXCLUDED.copy_id`, shelfID, userID, copyID)
	if err != nil {
		return fmt.Errorf("add book to shelf: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCopyNotOwned
	}
	return nil
}

func (r *Repository) RemoveBook(ctx context.Context, shelfID, userID, copyID string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM shelf_books sb
		USING shelves s
		WHERE sb.shelf_id=s.id
		  AND s.id=$1
		  AND s.user_id=$2
		  AND s.scope='personal'
		  AND sb.copy_id=$3`, shelfID, userID, copyID)
	if err != nil {
		return fmt.Errorf("remove book from shelf: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBookNotOnShelf
	}
	return nil
}

func (r *Repository) ListBooks(ctx context.Context, shelfID, userID string, page, limit int) ([]books.UserBook, int, error) {
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
		FROM shelf_books sb
		JOIN shelves s ON s.id=sb.shelf_id
		JOIN book_copies bc ON bc.id=sb.copy_id
		WHERE sb.shelf_id=$1
		  AND s.user_id=$2
		  AND s.scope='personal'
		  AND bc.owner_id=$2
		  AND bc.deleted_at IS NULL`, shelfID, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count shelf books: %w", err)
	}

	rows, err := r.db.Query(ctx, `
			SELECT bc.id, bc.reading_status, bc.current_page, bc.started_reading_at, bc.finished_reading_at,
		       bc.owned_by_user, bc.borrowed_from, bc.location, bc.condition, bc.reread_count,
		       bc.personal_notes, sb.added_at,
		       be.id, be.format, be.language, be.cover_url,
		       b.id, b.title, b.description, b.series_id, b.series_position, b.status,
			       b.deleted_at, b.created_at, b.updated_at
			FROM shelf_books sb
			JOIN shelves s ON s.id=sb.shelf_id
			JOIN book_copies bc ON bc.id=sb.copy_id
			JOIN book_editions be ON be.id=bc.edition_id
			JOIN books b ON b.id=be.book_id
			WHERE sb.shelf_id=$1
			  AND s.user_id=$2
			  AND s.scope='personal'
			  AND bc.owner_id=$2
			  AND bc.deleted_at IS NULL
			ORDER BY sb.added_at DESC
			LIMIT $3 OFFSET $4`, shelfID, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list shelf books: %w", err)
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
			return nil, 0, fmt.Errorf("scan shelf book: %w", err)
		}
		ub.Book.Authors = []books.Contributor{}
		ub.Book.Genres = []books.Genre{}
		userBooks = append(userBooks, ub)
	}
	return userBooks, total, rows.Err()
}
