package series

import "time"

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

type SeriesBook struct {
	BookID         string   `json:"book_id"`
	Title          string   `json:"title"`
	SeriesPosition *float64 `json:"series_position,omitempty"`
	CoverURL       *string  `json:"cover_url,omitempty"`
	Authors        []string `json:"authors"`
}

type SeriesDetail struct {
	Series
	Books []SeriesBook `json:"books"`
}
