package library

import "time"

type Copy struct {
	ID                string     `json:"id"`
	EditionID         string     `json:"edition_id"`
	OwnerID           string     `json:"owner_id"`
	Condition         *string    `json:"condition,omitempty"`
	ReadingStatus     string     `json:"reading_status"`
	CurrentPage       *int       `json:"current_page,omitempty"`
	StartedReadingAt  *time.Time `json:"started_reading_at,omitempty"`
	FinishedReadingAt *time.Time `json:"finished_reading_at,omitempty"`
	OwnedByUser       bool       `json:"owned_by_user"`
	BorrowedFrom      *string    `json:"borrowed_from,omitempty"`
	Location          *string    `json:"location,omitempty"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Author struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type BookSummary struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	// add series or description later if needed
	Authors []Author `json:"authors"`
}

type EditionSummary struct {
	ID       string  `json:"id"`
	Format   string  `json:"format"`
	Language *string `json:"language,omitempty"`
	CoverURL *string `json:"cover_url,omitempty"`
}

type UserBook struct {
	CopyID            string     `json:"copy_id"`
	ReadingStatus     string     `json:"reading_status"`
	CurrentPage       *int       `json:"current_page,omitempty"`
	StartedReadingAt  *time.Time `json:"started_reading_at,omitempty"`
	FinishedReadingAt *time.Time `json:"finished_reading_at,omitempty"`
	OwnedByUser       bool       `json:"owned_by_user"`
	BorrowedFrom      *string    `json:"borrowed_from,omitempty"`
	Location          *string    `json:"location,omitempty"`
	Condition         *string    `json:"condition,omitempty"`
	AddedAt           time.Time  `json:"added_at"`

	Edition EditionSummary `json:"edition"`
	Book    BookSummary    `json:"book"`
}

type ListLibraryResult struct {
	Books []UserBook `json:"books"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

type CopyOptions struct {
	ReadingStatus     string
	CurrentPage       *int
	StartedReadingAt  *string
	FinishedReadingAt *string
	OwnedByUser       *bool
	BorrowedFrom      *string
	Location          *string
}

type UpdateCopyInput struct {
	Status            string
	CurrentPage       *int
	StartedReadingAt  *string
	FinishedReadingAt *string
	OwnedByUser       *bool
	BorrowedFrom      *string
	Location          *string
}
