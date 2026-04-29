package contributors

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

type ContributorAward struct {
	AwardID   string  `json:"award_id"`
	AwardName string  `json:"award_name"`
	Year      int     `json:"year"`
	Category  *string `json:"category,omitempty"`
	Result    *string `json:"result,omitempty"` // winner | nominee
}

type ContributorDetail struct {
	Contributor
	Awards []ContributorAward `json:"awards,omitempty"`
}
