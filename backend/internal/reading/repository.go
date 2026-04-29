package reading

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ─── Challenges ───────────────────────────────────────────────────────────────

func (r *Repository) CreateChallenge(ctx context.Context, userID string, year, goal int) (Challenge, error) {
	var c Challenge
	err := r.db.QueryRow(ctx, `
		INSERT INTO reading_challenges (user_id, year, goal)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, year, goal, created_at`,
		userID, year, goal).Scan(&c.ID, &c.UserID, &c.Year, &c.Goal, &c.CreatedAt)
	if err != nil {
		return Challenge{}, fmt.Errorf("create challenge: %w", err)
	}
	return c, nil
}

func (r *Repository) ListChallenges(ctx context.Context, userID string) ([]Challenge, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, year, goal, created_at
		FROM reading_challenges WHERE user_id=$1 ORDER BY year DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list challenges: %w", err)
	}
	defer rows.Close()

	var challenges []Challenge
	for rows.Next() {
		var c Challenge
		if err := rows.Scan(&c.ID, &c.UserID, &c.Year, &c.Goal, &c.CreatedAt); err != nil {
			return nil, err
		}
		challenges = append(challenges, c)
	}
	return challenges, rows.Err()
}

func (r *Repository) FindChallengeByID(ctx context.Context, id string) (*Challenge, error) {
	var c Challenge
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, year, goal, created_at
		FROM reading_challenges WHERE id=$1`, id).Scan(&c.ID, &c.UserID, &c.Year, &c.Goal, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find challenge: %w", err)
	}
	return &c, nil
}

func (r *Repository) DeleteChallenge(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM reading_challenges WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete challenge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("challenge not found")
	}
	return nil
}

// CountBooksReadInYear counts books marked as 'read' in a given year for a user.
func (r *Repository) CountBooksReadInYear(ctx context.Context, userID string, year int) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM book_copies
		WHERE owner_id=$1
		  AND reading_status='read'
		  AND deleted_at IS NULL
		  AND EXTRACT(YEAR FROM finished_reading_at)=$2`,
		userID, year).Scan(&count)
	return count, err
}

// ─── Sessions ─────────────────────────────────────────────────────────────────

func (r *Repository) CreateSession(ctx context.Context, userID, copyID, loggedDate string, pagesRead *int, progressPct *float64, note *string) (Session, error) {
	var s Session
	err := r.db.QueryRow(ctx, `
		INSERT INTO reading_sessions (user_id, copy_id, logged_date, pages_read, progress_pct, note)
		VALUES ($1, $2, $3::date, $4, $5, $6)
		RETURNING id, user_id, copy_id, logged_date::text, pages_read, progress_pct, note, created_at`,
		userID, copyID, loggedDate, pagesRead, progressPct, note).Scan(
		&s.ID, &s.UserID, &s.CopyID, &s.LoggedDate, &s.PagesRead, &s.ProgressPct, &s.Note, &s.CreatedAt)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return s, nil
}

func (r *Repository) ListSessions(ctx context.Context, userID, copyID string, page, limit int) ([]Session, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	args := []any{userID}
	where := "user_id=$1"
	if copyID != "" {
		where += " AND copy_id=$2"
		args = append(args, copyID)
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reading_sessions WHERE `+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sessions: %w", err)
	}

	argN := len(args) + 1
	queryArgs := make([]any, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, user_id, copy_id, logged_date::text, pages_read, progress_pct, note, created_at
		FROM reading_sessions WHERE %s
		ORDER BY logged_date DESC, created_at DESC
		LIMIT $%d OFFSET $%d`, where, argN, argN+1), queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.CopyID, &s.LoggedDate, &s.PagesRead, &s.ProgressPct, &s.Note, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, s)
	}
	return sessions, total, rows.Err()
}

// ─── Statistics ───────────────────────────────────────────────────────────────

func (r *Repository) GetOverallStats(ctx context.Context, userID string) (OverallStats, error) {
	var stats OverallStats

	// Total books read and pages
	r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(be.page_count), 0)
		FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL`,
		userID).Scan(&stats.TotalBooksRead, &stats.TotalPagesRead)

	// Average rating
	var avg *float64
	r.db.QueryRow(ctx, `SELECT AVG(rating) FROM reviews WHERE user_id=$1 AND deleted_at IS NULL`, userID).Scan(&avg)
	stats.AverageRating = avg

	// Genre distribution
	gRows, err := r.db.Query(ctx, `
		SELECT g.name, COUNT(*) as cnt
		FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN book_genres bg ON bg.book_id=be.book_id
		JOIN genres g ON g.id=bg.genre_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		GROUP BY g.name ORDER BY cnt DESC LIMIT 10`, userID)
	if err == nil {
		for gRows.Next() {
			var gs GenreStat
			if err := gRows.Scan(&gs.Genre, &gs.Count); err == nil {
				stats.GenreDistribution = append(stats.GenreDistribution, gs)
			}
		}
		gRows.Close()
	}

	// Monthly pattern (last 24 months)
	mRows, err := r.db.Query(ctx, `
		SELECT EXTRACT(YEAR FROM finished_reading_at)::int,
		       EXTRACT(MONTH FROM finished_reading_at)::int,
		       COUNT(*),
		       COALESCE(SUM(be.page_count), 0)
		FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		  AND finished_reading_at >= NOW() - INTERVAL '24 months'
		GROUP BY 1, 2 ORDER BY 1 DESC, 2 DESC`, userID)
	if err == nil {
		for mRows.Next() {
			var ms MonthStat
			if err := mRows.Scan(&ms.Year, &ms.Month, &ms.Count, &ms.Pages); err == nil {
				stats.MonthlyPattern = append(stats.MonthlyPattern, ms)
			}
		}
		mRows.Close()
	}

	// Reading streak (consecutive days with a reading session)
	stats.ReadingStreak, stats.LongestStreak = r.calculateStreaks(ctx, userID)

	return stats, nil
}

func (r *Repository) calculateStreaks(ctx context.Context, userID string) (current, longest int) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT logged_date FROM reading_sessions
		WHERE user_id=$1 ORDER BY logged_date DESC`, userID)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err == nil {
			dates = append(dates, d)
		}
	}

	if len(dates) == 0 {
		return 0, 0
	}

	today := time.Now().Truncate(24 * time.Hour)
	cur := 0
	if dates[0].Truncate(24*time.Hour).Equal(today) || dates[0].Truncate(24*time.Hour).Equal(today.AddDate(0, 0, -1)) {
		cur = 1
		for i := 1; i < len(dates); i++ {
			diff := dates[i-1].Truncate(24*time.Hour).Sub(dates[i].Truncate(24 * time.Hour))
			if diff == 24*time.Hour {
				cur++
			} else {
				break
			}
		}
	}

	// Longest streak
	lon := 1
	run := 1
	for i := 1; i < len(dates); i++ {
		diff := dates[i-1].Truncate(24*time.Hour).Sub(dates[i].Truncate(24 * time.Hour))
		if diff == 24*time.Hour {
			run++
			if run > lon {
				lon = run
			}
		} else {
			run = 1
		}
	}

	return cur, lon
}

func (r *Repository) GetYearStats(ctx context.Context, userID string, year int) (YearStats, error) {
	stats := YearStats{Year: year}

	r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(be.page_count), 0)
		FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		  AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2`,
		userID, year).Scan(&stats.TotalBooksRead, &stats.TotalPagesRead)

	var avg *float64
	r.db.QueryRow(ctx, `
		SELECT AVG(r.rating) FROM reviews r
		JOIN book_copies bc ON bc.edition_id IN (SELECT id FROM book_editions WHERE book_id=r.book_id)
		WHERE r.user_id=$1 AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2 AND r.deleted_at IS NULL`,
		userID, year).Scan(&avg)
	stats.AverageRating = avg

	// Most read genre
	var genre *string
	r.db.QueryRow(ctx, `
		SELECT g.name FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN book_genres bg ON bg.book_id=be.book_id
		JOIN genres g ON g.id=bg.genre_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		  AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2
		GROUP BY g.name ORDER BY COUNT(*) DESC LIMIT 1`, userID, year).Scan(&genre)
	stats.MostReadGenre = genre

	// Most read author
	var author *string
	r.db.QueryRow(ctx, `
		SELECT c.name FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN book_contributors bco ON bco.book_id=be.book_id
		JOIN contributors c ON c.id=bco.contributor_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		  AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2
		  AND bco.role IN ('author','co_author')
		GROUP BY c.name ORDER BY COUNT(*) DESC LIMIT 1`, userID, year).Scan(&author)
	stats.MostReadAuthor = author

	// Monthly pattern
	mRows, err := r.db.Query(ctx, `
		SELECT EXTRACT(MONTH FROM bc.finished_reading_at)::int, COUNT(*), COALESCE(SUM(be.page_count),0)
		FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		  AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2
		GROUP BY 1 ORDER BY 1`, userID, year)
	if err == nil {
		for mRows.Next() {
			var ms MonthStat
			ms.Year = year
			if err := mRows.Scan(&ms.Month, &ms.Count, &ms.Pages); err == nil {
				stats.MonthlyPattern = append(stats.MonthlyPattern, ms)
			}
		}
		mRows.Close()
	}

	// Format distribution
	fRows, err := r.db.Query(ctx, `
		SELECT be.format, COUNT(*) FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		  AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2
		GROUP BY be.format ORDER BY COUNT(*) DESC`, userID, year)
	if err == nil {
		for fRows.Next() {
			var fs GenreStat
			if err := fRows.Scan(&fs.Genre, &fs.Count); err == nil {
				stats.FormatDistribution = append(stats.FormatDistribution, fs)
			}
		}
		fRows.Close()
	}

	return stats, nil
}

func (r *Repository) GetMonthStats(ctx context.Context, userID string, year, month int) (MonthStats, error) {
	stats := MonthStats{Year: year, Month: month}

	r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(be.page_count), 0)
		FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		  AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2
		  AND EXTRACT(MONTH FROM bc.finished_reading_at)=$3`,
		userID, year, month).Scan(&stats.TotalBooksRead, &stats.TotalPagesRead)

	// Favorite book (highest rated this month)
	var fav *string
	r.db.QueryRow(ctx, `
		SELECT b.title FROM reviews rv
		JOIN books b ON b.id=rv.book_id
		JOIN book_copies bc ON bc.edition_id IN (SELECT id FROM book_editions WHERE book_id=b.id)
		WHERE rv.user_id=$1 AND bc.owner_id=$1
		  AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2
		  AND EXTRACT(MONTH FROM bc.finished_reading_at)=$3
		  AND rv.deleted_at IS NULL AND bc.deleted_at IS NULL
		ORDER BY rv.rating DESC LIMIT 1`, userID, year, month).Scan(&fav)
	stats.FavoriteBook = fav

	// Genre breakdown
	gRows, err := r.db.Query(ctx, `
		SELECT g.name, COUNT(*) FROM book_copies bc
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN book_genres bg ON bg.book_id=be.book_id
		JOIN genres g ON g.id=bg.genre_id
		WHERE bc.owner_id=$1 AND bc.reading_status='read' AND bc.deleted_at IS NULL
		  AND EXTRACT(YEAR FROM bc.finished_reading_at)=$2
		  AND EXTRACT(MONTH FROM bc.finished_reading_at)=$3
		GROUP BY g.name ORDER BY COUNT(*) DESC`, userID, year, month)
	if err == nil {
		for gRows.Next() {
			var gs GenreStat
			if err := gRows.Scan(&gs.Genre, &gs.Count); err == nil {
				stats.GenreBreakdown = append(stats.GenreBreakdown, gs)
			}
		}
		gRows.Close()
	}

	return stats, nil
}
