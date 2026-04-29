package books

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/isbn"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

type CopyOptions struct {
	ReadingStatus     string
	CurrentPage       *int
	StartedReadingAt  *string
	FinishedReadingAt *string
	OwnedByUser       *bool
	BorrowedFrom      *string
	Location          *string
}

type EditionInput struct {
	Title           string   `json:"title"`
	OriginalTitle   string   `json:"original_title"`
	Format          string   `json:"format"`
	Description     *string  `json:"description"`
	CoverURL        *string  `json:"cover_url"`
	ISBN            *string  `json:"isbn"`
	ISBN10          *string  `json:"isbn10"`
	ISBN13          *string  `json:"isbn13"`
	ASIN            *string  `json:"asin"`
	Language        string   `json:"language"`
	Publisher       *string  `json:"publisher"`
	Edition         *string  `json:"edition"`
	PublishedAt     *string  `json:"published_at"`
	PageCount       *int     `json:"page_count"`
	FileFormat      *string  `json:"file_format"`
	DurationMinutes *int     `json:"duration_minutes"`
	AudioFormat     *string  `json:"audio_format"`
	Narrator        *string  `json:"narrator"`
	Translators     []string `json:"translators"`
}

type SubmitBookInput struct {
	Title          string
	Authors        []string
	Genres         []string
	SeriesName     *string
	SeriesPosition *float64
	Edition        EditionInput
	Condition      *string
	UserID         string
	UserRole       apictx.Role
	CatalogueOnly  bool
	CopyOptions
}

type AddCopyInput struct {
	EditionID string
	Condition *string
	UserID    string
	CopyOptions
}

type UpdateBookInput struct {
	ID             string
	Title          *string
	Description    *string
	Authors        []string
	Genres         []string
	SeriesName     *string
	SeriesPosition *float64
	EditionID      string
	Edition        *EditionInput
}

type SubmitBookResult struct {
	Submission Submission `json:"submission"`
	Book       Book       `json:"book"`
	Edition    Edition    `json:"edition"`
	Copy       *Copy      `json:"copy,omitempty"`
}

type AddCopyResult struct {
	Copy       Copy       `json:"copy"`
	Edition    Edition    `json:"edition"`
	Book       Book       `json:"book"`
	Submission Submission `json:"submission"`
}

// ─── Service ──────────────────────────────────────────────────────────────────

type Service struct {
	repo *Repository
	db   *pgxpool.Pool
}

func NewService(repo *Repository, db *pgxpool.Pool) *Service {
	return &Service{repo: repo, db: db}
}

func canAutoApprove(role apictx.Role) bool {
	return role == apictx.RoleModerator || role == apictx.RoleAdmin
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// normalizeEditionISBNs validates and normalises whichever ISBN field is provided.
// Returns (isbn10, isbn13, error).
func normalizeEditionISBNs(in *EditionInput) (*string, *string, error) {
	if in == nil {
		return nil, nil, nil
	}
	raw := ""
	if in.ISBN13 != nil && *in.ISBN13 != "" {
		raw = *in.ISBN13
	} else if in.ISBN10 != nil && *in.ISBN10 != "" {
		raw = *in.ISBN10
	} else if in.ISBN != nil && *in.ISBN != "" {
		raw = *in.ISBN
	}
	if raw == "" {
		return nil, nil, nil
	}
	pair, err := isbn.Normalize(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid ISBN: %w", err)
	}
	isbn13 := pair.ISBN13
	return pair.ISBN10, &isbn13, nil
}

// parsePublishedAt accepts "2001-09-01" or "2001".
func parsePublishedAt(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognised date format: %q", s)
}

// ─── SubmitBook ───────────────────────────────────────────────────────────────

func (s *Service) SubmitBook(ctx context.Context, input SubmitBookInput) (SubmitBookResult, error) {
	autoApprove := canAutoApprove(input.UserRole)
	var result SubmitBookResult

	isbn10, isbn13, err := normalizeEditionISBNs(&input.Edition)
	if err != nil {
		return SubmitBookResult{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return SubmitBookResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	repo := s.repo.withDB(tx)

	// ── Book: reuse existing if same title+authors, otherwise create ──────────
	var book Book
	if existing, _ := s.repo.FindBookByTitleAndAuthors(ctx, input.Title, input.Authors); existing != nil {
		book = *existing
		// Reload full details so response is complete
		if full, err := s.repo.FindBookWithDetails(ctx, book.ID); err == nil && full != nil {
			book = *full
		}
		// Merge any new genres the caller supplied
		existingGenres := map[string]bool{}
		for _, g := range book.Genres {
			existingGenres[strings.ToLower(g.Name)] = true
		}
		for _, gName := range input.Genres {
			if gName == "" || existingGenres[strings.ToLower(gName)] {
				continue
			}
			g, err := repo.FindOrCreateGenre(ctx, gName, autoApprove)
			if err != nil {
				return SubmitBookResult{}, err
			}
			if err := repo.LinkBookGenre(ctx, book.ID, g.ID); err != nil {
				return SubmitBookResult{}, err
			}
			book.Genres = append(book.Genres, g)
		}
	} else {
		book, err = repo.InsertBook(ctx, input.Title, nil, autoApprove)
		if err != nil {
			return SubmitBookResult{}, err
		}
		book.Authors = []Contributor{}
		book.Genres = []Genre{}

		for _, aName := range input.Authors {
			if aName == "" {
				continue
			}
			c, err := repo.FindOrCreateContributor(ctx, aName, autoApprove)
			if err != nil {
				return SubmitBookResult{}, err
			}
			if err := repo.LinkBookContributor(ctx, book.ID, c.ID, "author"); err != nil {
				return SubmitBookResult{}, err
			}
			book.Authors = append(book.Authors, c)
		}

		for _, gName := range input.Genres {
			if gName == "" {
				continue
			}
			g, err := repo.FindOrCreateGenre(ctx, gName, autoApprove)
			if err != nil {
				return SubmitBookResult{}, err
			}
			if err := repo.LinkBookGenre(ctx, book.ID, g.ID); err != nil {
				return SubmitBookResult{}, err
			}
			book.Genres = append(book.Genres, g)
		}
	}

	// ── Series ────────────────────────────────────────────────────────────────
	if input.SeriesName != nil && *input.SeriesName != "" {
		series, err := repo.FindOrCreateSeries(ctx, *input.SeriesName, autoApprove)
		if err != nil {
			return SubmitBookResult{}, err
		}
		if err := repo.UpdateBookSeries(ctx, book.ID, series.ID, input.SeriesPosition); err != nil {
			return SubmitBookResult{}, err
		}
		book.SeriesID = &series.ID
		book.SeriesPosition = input.SeriesPosition
		book.Series = &series
	}

	// ── Edition ───────────────────────────────────────────────────────────────
	edTitle := strings.TrimSpace(input.Edition.Title)
	if edTitle == "" {
		edTitle = strings.TrimSpace(input.Title)
	}
	origTitle := strings.TrimSpace(input.Edition.OriginalTitle)
	if origTitle == "" {
		origTitle = edTitle
	}

	ed := Edition{
		BookID:          book.ID,
		Title:           edTitle,
		OriginalTitle:   origTitle,
		Format:          input.Edition.Format,
		Description:     input.Edition.Description,
		CoverURL:        input.Edition.CoverURL,
		ISBN10:          isbn10,
		ISBN13:          isbn13,
		ASIN:            input.Edition.ASIN,
		Language:        input.Edition.Language,
		Publisher:       input.Edition.Publisher,
		Edition:         input.Edition.Edition,
		PageCount:       input.Edition.PageCount,
		FileFormat:      input.Edition.FileFormat,
		DurationMinutes: input.Edition.DurationMinutes,
		AudioFormat:     input.Edition.AudioFormat,
	}
	if input.Edition.PublishedAt != nil && *input.Edition.PublishedAt != "" {
		if t, err := parsePublishedAt(*input.Edition.PublishedAt); err == nil {
			ed.PublishedAt = &t
		}
	}

	edition, err := repo.InsertEdition(ctx, ed, autoApprove)
	if err != nil {
		if isUniqueViolation(err) {
			return SubmitBookResult{}, fmt.Errorf("edition with this ISBN already exists")
		}
		return SubmitBookResult{}, err
	}

	// ── Edition contributors ──────────────────────────────────────────────────
	if input.Edition.Narrator != nil && *input.Edition.Narrator != "" {
		c, err := repo.FindOrCreateContributor(ctx, *input.Edition.Narrator, autoApprove)
		if err != nil {
			return SubmitBookResult{}, err
		}
		if err := repo.LinkEditionContributor(ctx, edition.ID, c.ID, "narrator"); err != nil {
			return SubmitBookResult{}, err
		}
		edition.Narrators = append(edition.Narrators, c)
	}
	for _, tName := range input.Edition.Translators {
		if tName == "" {
			continue
		}
		c, err := repo.FindOrCreateContributor(ctx, tName, autoApprove)
		if err != nil {
			return SubmitBookResult{}, err
		}
		if err := repo.LinkEditionContributor(ctx, edition.ID, c.ID, "translator"); err != nil {
			return SubmitBookResult{}, err
		}
		edition.Translators = append(edition.Translators, c)
	}

	result.Book = book
	result.Edition = edition

	// ── Copy + Submission ─────────────────────────────────────────────────────
	if autoApprove && !input.CatalogueOnly {
		copy, err := repo.InsertCopy(ctx, edition.ID, input.UserID, input.Condition, input.CopyOptions)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Copy = &copy
		sub, err := repo.InsertSubmissionApproved(ctx, input.UserID, book.ID, edition.ID, copy.ID)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Submission = sub
	} else if autoApprove && input.CatalogueOnly {
		sub, err := repo.InsertSubmissionApprovedNoCopy(ctx, input.UserID, book.ID, edition.ID)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Submission = sub
	} else {
		sub, err := repo.InsertSubmission(ctx, input.UserID, book.ID, edition.ID)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Submission = sub
	}

	if err := tx.Commit(ctx); err != nil {
		return SubmitBookResult{}, fmt.Errorf("commit tx: %w", err)
	}

	// Reload book after commit so authors/genres are fully populated
	if full, err := s.repo.FindBookWithDetails(ctx, book.ID); err == nil && full != nil {
		result.Book = *full
	}

	return result, nil
}

// ─── AddCopyOfExistingEdition ─────────────────────────────────────────────────

func (s *Service) AddCopyOfExistingEdition(ctx context.Context, input AddCopyInput) (AddCopyResult, error) {
	edition, err := s.repo.FindEditionByID(ctx, input.EditionID)
	if err != nil || edition == nil {
		return AddCopyResult{}, fmt.Errorf("edition not found")
	}
	if edition.Status != "approved" {
		return AddCopyResult{}, fmt.Errorf("edition is not yet approved")
	}

	book, err := s.repo.FindBookByID(ctx, edition.BookID)
	if err != nil || book == nil {
		return AddCopyResult{}, fmt.Errorf("book not found")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AddCopyResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	repo := s.repo.withDB(tx)

	copy, err := repo.InsertCopy(ctx, edition.ID, input.UserID, input.Condition, input.CopyOptions)
	if err != nil {
		return AddCopyResult{}, err
	}

	sub, err := repo.InsertSubmissionApproved(ctx, input.UserID, book.ID, edition.ID, copy.ID)
	if err != nil {
		return AddCopyResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AddCopyResult{}, fmt.Errorf("commit tx: %w", err)
	}

	return AddCopyResult{Copy: copy, Edition: *edition, Book: *book, Submission: sub}, nil
}

// ─── Read operations ──────────────────────────────────────────────────────────

func (s *Service) ListBooks(ctx context.Context, page, limit int, sortBy string) (SearchResult, error) {
	books, total, err := s.repo.ListApprovedBooks(ctx, page, limit, sortBy)
	if err != nil {
		return SearchResult{}, err
	}
	if books == nil {
		books = []Book{}
	}
	return SearchResult{Books: books, Total: total, Page: page, Limit: limit}, nil
}

func (s *Service) SearchBooks(ctx context.Context, filters SearchFilters) (SearchResult, error) {
	books, total, err := s.repo.SearchBooks(ctx, filters)
	if err != nil {
		return SearchResult{}, err
	}
	if books == nil {
		books = []Book{}
	}
	return SearchResult{Books: books, Total: total, Page: filters.Page, Limit: filters.Limit}, nil
}

func (s *Service) GetBook(ctx context.Context, id string) (*Book, error) {
	return s.repo.FindBookWithDetails(ctx, id)
}

func (s *Service) FindExistingEditionByISBN(ctx context.Context, isbnRaw string) (*Edition, error) {
	if isbnRaw == "" {
		return nil, nil
	}
	pair, err := isbn.Normalize(isbnRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid ISBN: %w", err)
	}
	return s.repo.FindEditionByISBN(ctx, pair.ISBN13)
}

func (s *Service) FindEditionByID(ctx context.Context, id string) (*Edition, error) {
	return s.repo.FindEditionByID(ctx, id)
}

func (s *Service) GetUserBooks(ctx context.Context, userID string, page, limit int, status, genre, search, sortBy string) ([]UserBook, int, error) {
	return s.repo.GetUserBooks(ctx, userID, page, limit, status, genre, search, sortBy)
}

func (s *Service) GetBooksWithoutCovers(ctx context.Context) ([]Book, error) {
	return s.repo.FindBooksWithoutCovers(ctx)
}

func (s *Service) UpdateEditionCoverURL(ctx context.Context, editionID, coverURL string) error {
	return s.repo.UpdateEditionCoverURL(ctx, editionID, coverURL)
}

// ─── UpdateBook ───────────────────────────────────────────────────────────────

func (s *Service) UpdateBook(ctx context.Context, input UpdateBookInput) error {
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if err := s.repo.UpdateBook(ctx, input.ID, input.Title, input.Description); err != nil {
		return err
	}

	if input.SeriesName != nil && *input.SeriesName != "" {
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		repo := s.repo.withDB(tx)
		series, err := repo.FindOrCreateSeries(ctx, *input.SeriesName, true)
		if err != nil {
			return err
		}
		if err := repo.UpdateBookSeries(ctx, input.ID, series.ID, input.SeriesPosition); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}
	}

	if len(input.Authors) > 0 {
		// Replace authors: delete existing links then re-create
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()
		repo := s.repo.withDB(tx)
		if _, err := tx.Exec(ctx, `DELETE FROM book_contributors WHERE book_id=$1 AND role IN ('author','co_author')`, input.ID); err != nil {
			return err
		}
		for _, aName := range input.Authors {
			if aName == "" {
				continue
			}
			c, err := repo.FindOrCreateContributor(ctx, aName, true)
			if err != nil {
				return err
			}
			if err := repo.LinkBookContributor(ctx, input.ID, c.ID, "author"); err != nil {
				return err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	if input.Genres != nil {
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()
		repo := s.repo.withDB(tx)
		if _, err := tx.Exec(ctx, `DELETE FROM book_genres WHERE book_id=$1`, input.ID); err != nil {
			return err
		}
		for _, gName := range input.Genres {
			if gName == "" {
				continue
			}
			g, err := repo.FindOrCreateGenre(ctx, gName, true)
			if err != nil {
				return err
			}
			if err := repo.LinkBookGenre(ctx, input.ID, g.ID); err != nil {
				return err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	if input.EditionID != "" && input.Edition != nil {
		isbn10, isbn13, err := normalizeEditionISBNs(input.Edition)
		if err != nil {
			return err
		}
		ed := Edition{
			Title:           strings.TrimSpace(input.Edition.Title),
			OriginalTitle:   strings.TrimSpace(input.Edition.OriginalTitle),
			Format:          input.Edition.Format,
			Description:     input.Edition.Description,
			CoverURL:        input.Edition.CoverURL,
			ISBN10:          isbn10,
			ISBN13:          isbn13,
			ASIN:            input.Edition.ASIN,
			Language:        input.Edition.Language,
			Publisher:       input.Edition.Publisher,
			Edition:         input.Edition.Edition,
			PageCount:       input.Edition.PageCount,
			FileFormat:      input.Edition.FileFormat,
			DurationMinutes: input.Edition.DurationMinutes,
			AudioFormat:     input.Edition.AudioFormat,
		}
		if input.Edition.PublishedAt != nil && *input.Edition.PublishedAt != "" {
			if t, err := parsePublishedAt(*input.Edition.PublishedAt); err == nil {
				ed.PublishedAt = &t
			}
		}
		if err := s.repo.UpdateEditionDetails(ctx, input.EditionID, ed); err != nil {
			return err
		}

		// Replace translators if provided
		if input.Edition.Translators != nil {
			tx, err := s.db.Begin(ctx)
			if err != nil {
				return err
			}
			defer func() { _ = tx.Rollback(ctx) }()
			repo := s.repo.withDB(tx)
			if _, err := tx.Exec(ctx, `DELETE FROM edition_contributors WHERE edition_id=$1 AND role='translator'`, input.EditionID); err != nil {
				return err
			}
			for _, tName := range input.Edition.Translators {
				if tName == "" {
					continue
				}
				c, err := repo.FindOrCreateContributor(ctx, tName, true)
				if err != nil {
					return err
				}
				if err := repo.LinkEditionContributor(ctx, input.EditionID, c.ID, "translator"); err != nil {
					return err
				}
			}
			if err := tx.Commit(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *Service) DeleteBook(ctx context.Context, id string) error {
	return s.repo.DeleteBook(ctx, id)
}

// ─── Copy operations ──────────────────────────────────────────────────────────

func (s *Service) UpdateCopyStatus(ctx context.Context, copyID, userID, status string, currentPage *int) error {
	validStatuses := map[string]bool{
		"want_to_read": true, "reading": true, "read": true, "did_not_finish": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid reading status: %s", status)
	}
	return s.repo.UpdateCopyStatus(ctx, copyID, userID, status, currentPage)
}

func (s *Service) RemoveCopy(ctx context.Context, copyID, userID string) error {
	return s.repo.RemoveCopy(ctx, copyID, userID)
}
