package books

import (
	"time"

	"github.com/AgataPalma/biblios/internal/apictx"
)

type Author struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Narrator struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Translator struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Genre struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Book struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	CoverURL    *string    `json:"cover_url,omitempty"`
	Status      string     `json:"status"`
	Authors     []Author   `json:"authors"`
	Genres      []Genre    `json:"genres"`
	Editions    []Edition  `json:"editions,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Edition struct {
	ID              string       `json:"id"`
	BookID          string       `json:"book_id"`
	Format          string       `json:"format"`
	ISBN            *string      `json:"isbn,omitempty"`
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
	Translators     []Translator `json:"translators,omitempty"`
	Narrators       []Narrator   `json:"narrators,omitempty"`
	DeletedAt       *time.Time   `json:"deleted_at,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type Copy struct {
	ID        string     `json:"id"`
	EditionID string     `json:"edition_id"`
	OwnerID   string     `json:"owner_id"`
	Condition *string    `json:"condition,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Submission struct {
	ID              string     `json:"id"`
	SubmittedBy     string     `json:"submitted_by"`
	Status          string     `json:"status"`
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

// Role check helpers
func CanAutoApprove(role apictx.Role) bool {
	return role == apictx.RoleModerator || role == apictx.RoleAdmin
}
