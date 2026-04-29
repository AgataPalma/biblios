package books

import "time"

type Contributor struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Bio             *string    `json:"bio,omitempty"`
	BornDate        *time.Time `json:"born_date,omitempty"`
	DiedDate        *time.Time `json:"died_date,omitempty"`
	PhotoURL        *string    `json:"photo_url,omitempty"`
	Website         *string    `json:"website,omitempty"`
	Nationality     *string    `json:"nationality,omitempty"`
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

type Mood struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
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

type Award struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type BookAward struct {
	Award    Award   `json:"award"`
	Year     int     `json:"year"`
	Category *string `json:"category,omitempty"`
	Result   *string `json:"result,omitempty"` // winner | nominee
}

type Book struct {
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Description     *string       `json:"description,omitempty"`
	SeriesID        *string       `json:"series_id,omitempty"`
	SeriesPosition  *float64      `json:"series_position,omitempty"`
	Series          *Series       `json:"series,omitempty"`
	Status          string        `json:"status"`
	RejectionReason *string       `json:"rejection_reason,omitempty"`
	Authors         []Contributor `json:"authors"`
	Genres          []Genre       `json:"genres"`
	Moods           []Mood        `json:"moods,omitempty"`
	Awards          []BookAward   `json:"awards,omitempty"`
	Editions        []Edition     `json:"editions,omitempty"`
	AverageRating   *float64      `json:"average_rating,omitempty"`
	DeletedAt       *time.Time    `json:"deleted_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type Edition struct {
	ID              string        `json:"id"`
	BookID          string        `json:"book_id"`
	Title           string        `json:"title"`
	OriginalTitle   string        `json:"original_title"`
	Format          string        `json:"format"`
	Description     *string       `json:"description,omitempty"`
	CoverURL        *string       `json:"cover_url,omitempty"`
	ISBN10          *string       `json:"isbn10,omitempty"`
	ISBN13          *string       `json:"isbn13,omitempty"`
	ASIN            *string       `json:"asin,omitempty"`
	Language        string        `json:"language"`
	Publisher       *string       `json:"publisher,omitempty"`
	Edition         *string       `json:"edition,omitempty"`
	PublishedAt     *time.Time    `json:"published_at,omitempty"`
	PageCount       *int          `json:"page_count,omitempty"`
	FileFormat      *string       `json:"file_format,omitempty"`
	DurationMinutes *int          `json:"duration_minutes,omitempty"`
	AudioFormat     *string       `json:"audio_format,omitempty"`
	Status          string        `json:"status"`
	RejectionReason *string       `json:"rejection_reason,omitempty"`
	Narrators       []Contributor `json:"narrators,omitempty"`
	Translators     []Contributor `json:"translators,omitempty"`
	DeletedAt       *time.Time    `json:"deleted_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
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
	ContributorID   *string    `json:"contributor_id,omitempty"`
	Book            *Book      `json:"book,omitempty"`
	Edition         *Edition   `json:"edition,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
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
	CoverURL          *string    `json:"cover_url,omitempty"`
	Book              Book       `json:"book"`
}

type SearchFilters struct {
	Query     string
	Format    string
	Language  string
	Genre     string
	Series    string
	YearFrom  int
	YearTo    int
	Publisher string
	Award     string
	Mood      string
	Sort      string // relevance | newest | oldest | title
	Page      int
	Limit     int
}

type SearchResult struct {
	Books []Book `json:"books"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}
