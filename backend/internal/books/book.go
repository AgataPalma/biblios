package books

import (
	"fmt"
	"time"

	"github.com/AgataPalma/biblios/internal/apictx"
)

type Author struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Narrator struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Translator struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Genre struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	RejectionReason *string   `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type Book struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	SeriesID        *string    `json:"series_id,omitempty"`
	SeriesPosition  *float64   `json:"series_position,omitempty"`
	Series          *Series    `json:"series,omitempty"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	Authors         []Author   `json:"authors"`
	Genres          []Genre    `json:"genres"`
	Editions        []Edition  `json:"editions,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Edition struct {
	ID              string       `json:"id"`
	BookID          string       `json:"book_id"`
	Title           string       `json:"title"`
	OriginalTitle   string       `json:"original_title"`
	Format          string       `json:"format"`
	Description     *string      `json:"description,omitempty"`
	CoverURL        *string      `json:"cover_url,omitempty"`
	ISBN10          *string      `json:"isbn10,omitempty"`
	ISBN13          *string      `json:"isbn13,omitempty"`
	ASIN            *string      `json:"asin,omitempty"`
	Language        string       `json:"language"`
	Publisher       *string      `json:"publisher,omitempty"`
	Edition         *string      `json:"edition,omitempty"`
	PublishedAt     *time.Time   `json:"published_at,omitempty"`
	PageCount       *int         `json:"page_count,omitempty"`
	FileFormat      *string      `json:"file_format,omitempty"`
	DurationMinutes *int         `json:"duration_minutes,omitempty"`
	AudioFormat     *string      `json:"audio_format,omitempty"`
	Status          string       `json:"status"`
	RejectionReason *string      `json:"rejection_reason,omitempty"`
	Translators     []Translator `json:"translators,omitempty"`
	Narrators       []Narrator   `json:"narrators,omitempty"`
	DeletedAt       *time.Time   `json:"deleted_at,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

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
	BorrowedFrom      *string    `json:"borrowed_from,omitempty"` // user ID of real owner if borrowed
	Location          *string    `json:"location,omitempty"`
	RereadCount       int        `json:"reread_count"`
	PersonalNotes     *string    `json:"personal_notes,omitempty"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Submission struct {
	ID              string     `json:"id"`
	SubmittedBy     string     `json:"submitted_by"`
	Status          string     `json:"status"`
	CatalogueOnly   bool       `json:"catalogue_only"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	ReviewedBy      *string    `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	BookID          *string    `json:"book_id,omitempty"`
	EditionID       *string    `json:"edition_id,omitempty"`
	CopyID          *string    `json:"copy_id,omitempty"`
	Book            *Book      `json:"book,omitempty"`
	Edition         *Edition   `json:"edition,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type BookWithDetails struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Editions []Edition `json:"editions"`
}

type LookupResult struct {
	Title       string `json:"title"`
	CoverURL    string `json:"cover_url"`
	ISBN        string `json:"isbn"`
	Description string `json:"description"`
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
	RereadCount       int        `json:"reread_count"`
	PersonalNotes     *string    `json:"personal_notes,omitempty"`
	AddedAt           time.Time  `json:"added_at"`
	EditionID         string     `json:"edition_id"`
	Format            string     `json:"format"`
	Language          *string    `json:"language,omitempty"`
	CoverURL          *string    `json:"cover_url,omitempty"` // from book_editions
	Book              Book       `json:"book"`
}

type UserBooksResult struct {
	Books []UserBook `json:"books"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

func (e Edition) PreferredISBN() string {
	if e.ISBN13 != nil && *e.ISBN13 != "" {
		return *e.ISBN13
	}
	if e.ISBN10 != nil && *e.ISBN10 != "" {
		return *e.ISBN10
	}
	return ""
}

type Series struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Role check helpers
func CanAutoApprove(role apictx.Role) bool {
	return role == apictx.RoleModerator || role == apictx.RoleAdmin
}

// parsePublishedAt parses a user-supplied year ("2001") or full date ("2001-09-01")
// into a time.Time. Returns an error if the string is not a recognised format.
func parsePublishedAt(s string) (time.Time, error) {
	// Try full date first
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// Try year only
	if t, err := time.Parse("2006", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognised date format: %q", s)
}
