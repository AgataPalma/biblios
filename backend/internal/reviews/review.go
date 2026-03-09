package reviews

import "time"

type Review struct {
	ID        string     `json:"id"`
	BookID    string     `json:"book_id"`
	UserID    *string    `json:"user_id,omitempty"`
	Username  *string    `json:"username,omitempty"`
	Rating    int        `json:"rating"`
	Body      *string    `json:"body,omitempty"`
	IsPublic  bool       `json:"is_public"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type UpsertReviewInput struct {
	BookID   string
	UserID   string
	Rating   int
	Body     *string
	IsPublic *bool
}
