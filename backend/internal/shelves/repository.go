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
		INSERT INTO shelves (user_id, name)
		VALUES ($1, $2)
		RETURNING id, user_id, name, created_at`,
		userID, name).Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err != nil {
		return Shelf{}, fmt.Errorf("create shelf: %w", err)
	}
	return s, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Shelf, error) {
	var s Shelf
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, name, created_at FROM shelves WHERE id=$1`, id).Scan(
		&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find shelf: %w", err)
	}
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM shelf_books WHERE shelf_id=$1`, id).Scan(&s.BookCount)
	return &s, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Shelf, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, created_at FROM shelves
		WHERE user_id=$1 ORDER BY name ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list shelves: %w", err)
	}
	defer rows.Close()

	var shelves []Shelf
	for rows.Next() {
		var s Shelf
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt); err != nil {
			return nil, err
		}
		shelves = append(shelves, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, sh := range shelves {
		r.db.QueryRow(ctx, `SELECT COUNT(*) FROM shelf_books WHERE shelf_id=$1`, sh.ID).Scan(&shelves[i].BookCount)
	}
	return shelves, nil
}

func (r *Repository) Rename(ctx context.Context, id, userID, name string) (Shelf, error) {
	var s Shelf
	err := r.db.QueryRow(ctx, `
		UPDATE shelves SET name=$3
		WHERE id=$1 AND user_id=$2
		RETURNING id, user_id, name, created_at`,
		id, userID, name).Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return Shelf{}, fmt.Errorf("shelf not found")
	}
	if err != nil {
		return Shelf{}, fmt.Errorf("rename shelf: %w", err)
	}
	return s, nil
}

// Delete removes the shelf and cascades shelf_books via FK ON DELETE CASCADE.
func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM shelves WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete shelf: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("shelf not found")
	}
	return nil
}

func (r *Repository) AddBook(ctx context.Context, shelfID, copyID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO shelf_books (shelf_id, copy_id) VALUES ($1,$2)
		ON CONFLICT DO NOTHING`, shelfID, copyID)
	if err != nil {
		return fmt.Errorf("add book to shelf: %w", err)
	}
	return nil
}

func (r *Repository) RemoveBook(ctx context.Context, shelfID, copyID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM shelf_books WHERE shelf_id=$1 AND copy_id=$2`, shelfID, copyID)
	if err != nil {
		return fmt.Errorf("remove book from shelf: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("book not found on shelf")
	}
	return nil
}

func (r *Repository) ListBooks(ctx context.Context, shelfID string, page, limit int) ([]books.UserBook, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM shelf_books WHERE shelf_id=$1`, shelfID).Scan(&total); err != nil {
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
		JOIN book_copies bc ON bc.id=sb.copy_id
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN books b ON b.id=be.book_id
		WHERE sb.shelf_id=$1 AND bc.deleted_at IS NULL
		ORDER BY sb.added_at DESC
		LIMIT $2 OFFSET $3`, shelfID, limit, offset)
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
