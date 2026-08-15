package collections

import "time"

type Collection struct {
	ID              string     `json:"id"`
	LibraryID       string     `json:"library_id"`
	CreatedBy       string     `json:"created_by,omitempty"`
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
	BookCopyID   string    `json:"book_copy_id,omitempty"`
	BookID       string    `json:"book_id"`
	EditionID    string    `json:"edition_id"`
	Title        string    `json:"title"`
	Authors      []string  `json:"authors"`
	Format       string    `json:"format"`
	Language     *string   `json:"language,omitempty"`
	CoverURL     *string   `json:"cover_url,omitempty"`
	AddedAt      time.Time `json:"added_at"`
}

// LibraryAccess is the collection package's view of a caller's access to a
// library. Keeping this small prevents collection handlers from reaching into
// the library repository and makes the authorization rules testable.
type LibraryAccess struct {
	Visibility string
	IsMember   bool
	IsOwner    bool
	CanView    bool
	CanAdd     bool
	CanRemove  bool
	CanEdit    bool
}
