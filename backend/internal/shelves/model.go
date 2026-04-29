package shelves

import "time"

type Shelf struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	BookCount int       `json:"book_count,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ShelfBook struct {
	ShelfID string    `json:"shelf_id"`
	CopyID  string    `json:"copy_id"`
	AddedAt time.Time `json:"added_at"`
}
