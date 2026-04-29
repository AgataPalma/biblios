package collections

import "time"

type Collection struct {
	ID              string     `json:"id"`
	LibraryID       string     `json:"library_id"`
	CreatedBy       string     `json:"created_by"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	CoverColour     *string    `json:"cover_colour,omitempty"`
	IsPublic        bool       `json:"is_public"`
	IsCollaborative bool       `json:"is_collaborative"`
	BookCount       int        `json:"book_count,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CollectionBook struct {
	CollectionID string    `json:"collection_id"`
	BookCopyID   string    `json:"book_copy_id"`
	AddedBy      string    `json:"added_by"`
	AddedAt      time.Time `json:"added_at"`
}
