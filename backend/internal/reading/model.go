package reading

import "time"

type Challenge struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Year      int       `json:"year"`
	Goal      int       `json:"goal"`
	CreatedAt time.Time `json:"created_at"`
}

type ChallengeProgress struct {
	Challenge     Challenge `json:"challenge"`
	BooksRead     int       `json:"books_read"`
	BooksRemaining int      `json:"books_remaining"`
	ProgressPct   float64   `json:"progress_pct"`
	MonthlyPace   float64   `json:"monthly_pace"`   // avg books/month so far
	ProjectedEnd  *string   `json:"projected_end,omitempty"` // estimated completion month
}

type Session struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	CopyID      string    `json:"copy_id"`
	LoggedDate  string    `json:"logged_date"` // date only: YYYY-MM-DD
	PagesRead   *int      `json:"pages_read,omitempty"`
	ProgressPct *float64  `json:"progress_pct,omitempty"`
	Note        *string   `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type MonthStat struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Count int `json:"count"`
	Pages int `json:"pages"`
}

type GenreStat struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

type OverallStats struct {
	TotalBooksRead   int         `json:"total_books_read"`
	TotalPagesRead   int         `json:"total_pages_read"`
	AverageRating    *float64    `json:"average_rating,omitempty"`
	ReadingStreak    int         `json:"reading_streak_days"`
	LongestStreak    int         `json:"longest_streak_days"`
	GenreDistribution []GenreStat `json:"genre_distribution"`
	MonthlyPattern   []MonthStat `json:"monthly_pattern"`
}

type YearStats struct {
	Year             int         `json:"year"`
	TotalBooksRead   int         `json:"total_books_read"`
	TotalPagesRead   int         `json:"total_pages_read"`
	AverageRating    *float64    `json:"average_rating,omitempty"`
	MostReadGenre    *string     `json:"most_read_genre,omitempty"`
	MostReadAuthor   *string     `json:"most_read_author,omitempty"`
	LongestBook      *string     `json:"longest_book,omitempty"`
	ShortestBook     *string     `json:"shortest_book,omitempty"`
	AvgDaysToFinish  *float64    `json:"avg_days_to_finish,omitempty"`
	MonthlyPattern   []MonthStat `json:"monthly_pattern"`
	FormatDistribution []GenreStat `json:"format_distribution"`
}

type MonthStats struct {
	Year           int         `json:"year"`
	Month          int         `json:"month"`
	TotalBooksRead int         `json:"total_books_read"`
	TotalPagesRead int         `json:"total_pages_read"`
	FavoriteBook   *string     `json:"favorite_book,omitempty"`
	GenreBreakdown []GenreStat `json:"genre_breakdown"`
}
