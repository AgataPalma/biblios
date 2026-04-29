package reading

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateChallengeInput struct {
	Year int `json:"year"`
	Goal int `json:"goal"`
}

type LogSessionInput struct {
	CopyID      string   `json:"copy_id"`
	LoggedDate  string   `json:"logged_date"` // YYYY-MM-DD
	PagesRead   *int     `json:"pages_read"`
	ProgressPct *float64 `json:"progress_pct"`
	Note        *string  `json:"note"`
}

func (s *Service) CreateChallenge(ctx context.Context, userID string, input CreateChallengeInput) (Challenge, error) {
	if input.Year < 2000 || input.Year > time.Now().Year()+5 {
		return Challenge{}, fmt.Errorf("invalid year")
	}
	if input.Goal < 1 {
		return Challenge{}, fmt.Errorf("goal must be at least 1")
	}
	return s.repo.CreateChallenge(ctx, userID, input.Year, input.Goal)
}

func (s *Service) ListChallenges(ctx context.Context, userID string) ([]Challenge, error) {
	return s.repo.ListChallenges(ctx, userID)
}

func (s *Service) DeleteChallenge(ctx context.Context, id, userID string) error {
	return s.repo.DeleteChallenge(ctx, id, userID)
}

func (s *Service) GetProgress(ctx context.Context, id, userID string) (ChallengeProgress, error) {
	c, err := s.repo.FindChallengeByID(ctx, id)
	if err != nil {
		return ChallengeProgress{}, err
	}
	if c == nil {
		return ChallengeProgress{}, fmt.Errorf("challenge not found")
	}
	if c.UserID != userID {
		return ChallengeProgress{}, fmt.Errorf("challenge not found")
	}

	booksRead, err := s.repo.CountBooksReadInYear(ctx, userID, c.Year)
	if err != nil {
		return ChallengeProgress{}, err
	}

	remaining := c.Goal - booksRead
	if remaining < 0 {
		remaining = 0
	}
	pct := 0.0
	if c.Goal > 0 {
		pct = float64(booksRead) / float64(c.Goal) * 100
		if pct > 100 {
			pct = 100
		}
	}

	// Monthly pace: months elapsed in the challenge year so far
	now := time.Now()
	monthsElapsed := 1.0
	if c.Year == now.Year() {
		monthsElapsed = float64(now.Month())
	} else if c.Year < now.Year() {
		monthsElapsed = 12
	}
	pace := float64(booksRead) / monthsElapsed

	// Projected completion month
	var projected *string
	if pace > 0 && booksRead < c.Goal {
		monthsNeeded := float64(remaining) / pace
		projMonth := now.AddDate(0, int(monthsNeeded)+1, 0)
		s := projMonth.Format("January 2006")
		projected = &s
	}

	return ChallengeProgress{
		Challenge:      *c,
		BooksRead:      booksRead,
		BooksRemaining: remaining,
		ProgressPct:    pct,
		MonthlyPace:    pace,
		ProjectedEnd:   projected,
	}, nil
}

func (s *Service) LogSession(ctx context.Context, userID string, input LogSessionInput) (Session, error) {
	if input.CopyID == "" {
		return Session{}, fmt.Errorf("copy_id is required")
	}
	if input.LoggedDate == "" {
		input.LoggedDate = time.Now().Format("2006-01-02")
	}
	if input.ProgressPct != nil && (*input.ProgressPct < 0 || *input.ProgressPct > 100) {
		return Session{}, fmt.Errorf("progress_pct must be between 0 and 100")
	}
	return s.repo.CreateSession(ctx, userID, input.CopyID, input.LoggedDate, input.PagesRead, input.ProgressPct, input.Note)
}

func (s *Service) ListSessions(ctx context.Context, userID, copyID string, page, limit int) ([]Session, int, error) {
	return s.repo.ListSessions(ctx, userID, copyID, page, limit)
}

func (s *Service) GetOverallStats(ctx context.Context, userID string) (OverallStats, error) {
	return s.repo.GetOverallStats(ctx, userID)
}

func (s *Service) GetYearStats(ctx context.Context, userID string, year int) (YearStats, error) {
	return s.repo.GetYearStats(ctx, userID, year)
}

func (s *Service) GetMonthStats(ctx context.Context, userID string, year, month int) (MonthStats, error) {
	if month < 1 || month > 12 {
		return MonthStats{}, fmt.Errorf("invalid month")
	}
	return s.repo.GetMonthStats(ctx, userID, year, month)
}
