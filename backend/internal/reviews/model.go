package reviews

import "time"

type Review struct {
	ID        string     `json:"id"`
	BookID    string     `json:"book_id"`
	UserID    *string    `json:"user_id,omitempty"` // nullable after account deletion
	Username  *string    `json:"username,omitempty"`
	AvatarURL *string    `json:"avatar_url,omitempty"`
	Rating    float64    `json:"rating"`
	Body      *string    `json:"body,omitempty"`
	IsPublic  bool       `json:"is_public"`
	LikeCount int        `json:"like_count"`
	LikedByMe bool       `json:"liked_by_me,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type ReviewsResponse struct {
	Reviews       []Review `json:"reviews"`
	Total         int      `json:"total"`
	Page          int      `json:"page"`
	Limit         int      `json:"limit"`
	AverageRating *float64 `json:"average_rating,omitempty"`
}
