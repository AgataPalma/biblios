package books

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/isbn"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bookRepository interface {
	UpdateBook(ctx context.Context, id string, title *string, description *string) error
	UpdateEditionDetails(ctx context.Context, editionID string, e Edition) error
	ListApprovedBooks(ctx context.Context, page, limit int, sort string) ([]Book, int, error)
	FindBookWithDetails(ctx context.Context, id string) (*Book, error)
	FindBooksWithoutCovers(ctx context.Context) ([]BookWithDetails, error)
	UpdateEditionCoverURL(ctx context.Context, editionID, coverURL string) error
	FindEditionByISBN(ctx context.Context, isbn string) (*Edition, error)
	FindBookByID(ctx context.Context, id string) (*Book, error)
	FindBookByTitleAndAuthors(ctx context.Context, title string, authorNames []string) (*Book, error)
	FindSubmissionByID(ctx context.Context, id string) (*Submission, error)
	DeleteBook(ctx context.Context, id string) error
	ForceDeleteBook(ctx context.Context, id string) error
	DeleteEdition(ctx context.Context, editionID string) error
	ReplaceBookAuthors(ctx context.Context, bookID string, authorNames []string, autoApprove bool) error
	ReplaceBookGenres(ctx context.Context, bookID string, genreNames []string, autoApprove bool) error
	ReplaceEditionTranslators(ctx context.Context, editionID string, names []string) error
	FindEditionByID(ctx context.Context, id string) (*Edition, error)
	withDB(db DB) *txRepository // needed so SubmitBook can create a txRepository
}

type Service struct {
	repo bookRepository
	db   *pgxpool.Pool
}

func NewService(repo *Repository, db *pgxpool.Pool) *Service {
	return &Service{repo: repo, db: db}
}

type SubmitBookInput struct {
	Title          string
	Authors        []string
	Genres         []string
	SeriesName     *string  `json:"series_name"`
	SeriesPosition *float64 `json:"series_position"`
	Edition        EditionInput
	Condition      *string
	UserID         string
	UserRole       apictx.Role
	CatalogueOnly  bool // when true, book is added to catalogue only — no personal copy created
	CopyOptions         // initial reading state for the personal copy
}

type EditionInput struct {
	Title           string   `json:"title"`
	OriginalTitle   string   `json:"original_title"`
	Format          string   `json:"format"`
	Description     *string  `json:"description"`
	CoverURL        *string  `json:"cover_url"`
	ISBN            *string  `json:"isbn"` // accepted as an alias; normalized into ISBN10/ISBN13
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
	Narrator        *string  `json:"narrator"`    // single name — backend does FindOrCreate + link
	Translators     []string `json:"translators"` // multiple names — each gets FindOrCreate + link
}

type SubmitBookResult struct {
	Submission Submission
	Book       Book
	Edition    Edition
	Copy       *Copy
}

// CopyOptions holds the initial state of a newly created copy.
// Used by both AddCopyOfExistingEdition and SubmitBook.
type CopyOptions struct {
	ReadingStatus     string
	CurrentPage       *int
	StartedReadingAt  *string // RFC3339 or empty string to skip
	FinishedReadingAt *string // RFC3339 or empty string to skip
	OwnedByUser       *bool   // nil = default true
	BorrowedFrom      *string // user ID of real owner if borrowed
	Location          *string
}

type AddCopyInput struct {
	EditionID string
	Condition *string
	UserID    string
	CopyOptions
}

type UpdateCopyInput struct {
	Status            string
	CurrentPage       *int
	StartedReadingAt  *string // RFC3339 or empty
	FinishedReadingAt *string // RFC3339 or empty
	OwnedByUser       *bool
	BorrowedFrom      *string
	Location          *string
}

type AddCopyResult struct {
	Copy       Copy
	Edition    Edition
	Book       Book
	Submission Submission
}

type ListBooksResult struct {
	Books []Book `json:"books"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
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

func normalizeEditionISBNs(input *EditionInput) (*string, *string, error) {
	if input == nil {
		return nil, nil, nil
	}
	if input.ISBN13 != nil && *input.ISBN13 != "" {
		pair, err := isbn.Normalize(*input.ISBN13)
		if err != nil {
			return nil, nil, err
		}
		var isbn10 *string
		if pair.ISBN10 != nil {
			v := *pair.ISBN10
			isbn10 = &v
		}
		isbn13 := pair.ISBN13
		return isbn10, &isbn13, nil
	}
	if input.ISBN10 != nil && *input.ISBN10 != "" {
		pair, err := isbn.Normalize(*input.ISBN10)
		if err != nil {
			return nil, nil, err
		}
		var isbn10 *string
		if pair.ISBN10 != nil {
			v := *pair.ISBN10
			isbn10 = &v
		}
		isbn13 := pair.ISBN13
		return isbn10, &isbn13, nil
	}
	if input.ISBN != nil && *input.ISBN != "" {
		pair, err := isbn.Normalize(*input.ISBN)
		if err != nil {
			return nil, nil, err
		}
		var isbn10 *string
		if pair.ISBN10 != nil {
			v := *pair.ISBN10
			isbn10 = &v
		}
		isbn13 := pair.ISBN13
		return isbn10, &isbn13, nil
	}
	return nil, nil, nil
}

func (s *Service) AddCopyOfExistingEdition(ctx context.Context, input AddCopyInput) (AddCopyResult, error) {
	var result AddCopyResult
	var err error

	// Check edition exists and is approved
	var edition *Edition
	edition, err = s.repo.FindEditionByID(ctx, input.EditionID)
	if err != nil || edition == nil {
		return AddCopyResult{}, fmt.Errorf("edition not found")
	}

	if edition.Status != "approved" {
		return AddCopyResult{}, fmt.Errorf("edition is not yet approved")
	}

	// Get the book
	var book *Book
	book, err = s.repo.FindBookByID(ctx, edition.BookID)
	if err != nil || book == nil {
		return AddCopyResult{}, fmt.Errorf("book not found")
	}

	// Start transaction
	var tx pgxTx
	tx, err = s.db.Begin(ctx)
	if err != nil {
		return AddCopyResult{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var repo *txRepository = s.repo.withDB(tx)

	// Create copy
	var copyStruct Copy
	copyStruct, err = repo.InsertCopy(ctx, edition.ID, input.UserID, input.Condition, input.CopyOptions)
	if err != nil {
		return AddCopyResult{}, err
	}

	// Create approved submission immediately since edition already exists and is approved
	var submission Submission
	submission, err = repo.InsertSubmissionApproved(ctx, input.UserID, book.ID, edition.ID, copyStruct.ID)
	if err != nil {
		return AddCopyResult{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return AddCopyResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result.Copy = copyStruct
	result.Edition = *edition
	result.Book = *book
	result.Submission = submission

	return result, nil
}

func (s *Service) FindExistingEditionByISBN(ctx context.Context, isbnInput string) (*Edition, error) {
	if isbnInput == "" {
		return nil, nil
	}
	if _, err := isbn.Normalize(isbnInput); err != nil {
		return nil, fmt.Errorf("invalid ISBN: %w", err)
	}
	return s.repo.FindEditionByISBN(ctx, isbn.Clean(isbnInput))
}

func (s *Service) SubmitBook(ctx context.Context, input SubmitBookInput) (SubmitBookResult, error) {
	var autoApprove bool = CanAutoApprove(input.UserRole)
	var result SubmitBookResult

	isbn10, isbn13, err := normalizeEditionISBNs(&input.Edition)
	if err != nil {
		return SubmitBookResult{}, fmt.Errorf("invalid ISBN: %w", err)
	}

	// Start transaction
	var tx pgxTx
	tx, err = s.db.Begin(ctx)
	if err != nil {
		return SubmitBookResult{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// All operations go through the transaction
	var repo *txRepository = s.repo.withDB(tx)

	// Check for existing book with same title + same authors before creating a new one
	var book Book
	if existingBook, lookupErr := s.repo.FindBookByTitleAndAuthors(ctx, input.Title, input.Authors); lookupErr == nil && existingBook != nil {
		book = *existingBook

		// Load existing authors and genres back onto the book struct for the response
		existingWithDetails, detailErr := s.repo.FindBookWithDetails(ctx, book.ID)
		if detailErr == nil && existingWithDetails != nil {
			book.Authors = existingWithDetails.Authors
			book.Genres = existingWithDetails.Genres
		}

		// Add any genres from the new submission that the book doesn't already have
		existingGenreNames := make(map[string]bool)
		for _, g := range book.Genres {
			existingGenreNames[strings.ToLower(g.Name)] = true
		}
		for _, genreName := range input.Genres {
			if genreName == "" || existingGenreNames[strings.ToLower(genreName)] {
				continue
			}
			var genre Genre
			genre, err = repo.FindOrCreateGenre(ctx, genreName, autoApprove)
			if err != nil {
				return SubmitBookResult{}, err
			}
			if linkErr := repo.LinkBookGenre(ctx, book.ID, genre.ID); linkErr != nil {
				return SubmitBookResult{}, linkErr
			}
			book.Genres = append(book.Genres, genre)
		}
	} else {
		// Create book
		book, err = repo.InsertBook(ctx, input.Title, autoApprove)
		if err != nil {
			return SubmitBookResult{}, err
		}

		// Create and link authors
		for _, authorName := range input.Authors {
			var author Author
			author, err = repo.FindOrCreateAuthor(ctx, authorName, autoApprove)
			if err != nil {
				return SubmitBookResult{}, err
			}
			err = repo.LinkBookAuthor(ctx, book.ID, author.ID)
			if err != nil {
				return SubmitBookResult{}, err
			}
			book.Authors = append(book.Authors, author)
		}

		// Create and link genres
		for _, genreName := range input.Genres {
			var genre Genre
			genre, err = repo.FindOrCreateGenre(ctx, genreName, autoApprove)
			if err != nil {
				return SubmitBookResult{}, err
			}
			err = repo.LinkBookGenre(ctx, book.ID, genre.ID)
			if err != nil {
				return SubmitBookResult{}, err
			}
			book.Genres = append(book.Genres, genre)
		}
	}

	// After book is created/found, link series if provided
	if input.SeriesName != nil && *input.SeriesName != "" {
		series, err := repo.FindOrCreateSeries(ctx, *input.SeriesName, autoApprove)
		if err != nil {
			return SubmitBookResult{}, err
		}
		if err = repo.UpdateBookSeries(ctx, book.ID, series.ID, input.SeriesPosition); err != nil {
			return SubmitBookResult{}, err
		}
		book.SeriesID = &series.ID
		book.SeriesPosition = input.SeriesPosition
		book.Series = &series
	}

	editionTitle := strings.TrimSpace(input.Edition.Title)
	if editionTitle == "" {
		editionTitle = strings.TrimSpace(input.Title)
	}
	originalTitle := strings.TrimSpace(input.Edition.OriginalTitle)
	if originalTitle == "" {
		originalTitle = editionTitle
	}

	// Create edition
	var edition Edition = Edition{
		BookID:          book.ID,
		Title:           editionTitle,
		OriginalTitle:   originalTitle,
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
		if t, parseErr := parsePublishedAt(*input.Edition.PublishedAt); parseErr == nil {
			edition.PublishedAt = &t
		}
	}
	edition, err = repo.InsertEdition(ctx, edition, autoApprove)
	if err != nil {
		if isUniqueViolation(err) {
			return SubmitBookResult{}, fmt.Errorf("edition with this ISBN already exists")
		}
		return SubmitBookResult{}, err
	}

	// Link narrator (audiobooks)
	if input.Edition.Narrator != nil && *input.Edition.Narrator != "" {
		var narrator Narrator
		narrator, err = repo.FindOrCreateNarrator(ctx, *input.Edition.Narrator, autoApprove)
		if err != nil {
			return SubmitBookResult{}, err
		}
		err = repo.LinkEditionNarrator(ctx, edition.ID, narrator.ID)
		if err != nil {
			return SubmitBookResult{}, err
		}
		edition.Narrators = append(edition.Narrators, narrator)
	}

	// Link translators (non-audiobooks)
	for _, translatorName := range input.Edition.Translators {
		if translatorName == "" {
			continue
		}
		var translator Translator
		translator, err = repo.FindOrCreateTranslator(ctx, translatorName, autoApprove)
		if err != nil {
			return SubmitBookResult{}, err
		}
		err = repo.LinkEditionTranslator(ctx, edition.ID, translator.ID)
		if err != nil {
			return SubmitBookResult{}, err
		}
		edition.Translators = append(edition.Translators, translator)
	}

	result.Book = book
	result.Edition = edition

	// If auto approve, create copy — unless this is a catalogue-only submission
	if autoApprove && !input.CatalogueOnly {
		var copyStruct Copy
		copyStruct, err = repo.InsertCopy(ctx, edition.ID, input.UserID, input.Condition, input.CopyOptions)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Copy = &copyStruct

		var submission Submission
		submission, err = repo.InsertSubmissionApproved(ctx, input.UserID, book.ID, edition.ID, copyStruct.ID)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Submission = submission
	} else if autoApprove && input.CatalogueOnly {
		// Book and edition are approved and visible in the catalogue — no copy, no owner
		var submission Submission
		submission, err = repo.InsertSubmissionApprovedNoCopy(ctx, input.UserID, book.ID, edition.ID)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Submission = submission
	} else {
		var submission Submission
		submission, err = repo.InsertSubmission(ctx, input.UserID, book.ID, edition.ID)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Submission = submission
	}

	if err = tx.Commit(ctx); err != nil {
		return SubmitBookResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload the book after commit to ensure authors/genres are populated in the response
	if reloaded, reloadErr := s.repo.FindBookWithDetails(ctx, book.ID); reloadErr == nil && reloaded != nil {
		result.Book = *reloaded
	}

	return result, nil
}

func (s *Service) ListBooks(ctx context.Context, page int, limit int, sort string) (ListBooksResult, error) {
	var bookList []Book
	var total int
	var err error

	bookList, total, err = s.repo.ListApprovedBooks(ctx, page, limit, sort)
	if err != nil {
		return ListBooksResult{}, err
	}

	if bookList == nil {
		bookList = []Book{}
	}

	return ListBooksResult{
		Books: bookList,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *Service) GetBook(ctx context.Context, id string) (*Book, error) {
	return s.repo.FindBookWithDetails(ctx, id)
}

func (s *Service) UpdateBook(ctx context.Context, input UpdateBookInput) error {
	if input.Title != nil && *input.Title == "" {
		return fmt.Errorf("title is required")
	}
	if err := s.repo.UpdateBook(ctx, input.ID, input.Title, input.Description); err != nil {
		return err
	}
	if input.SeriesName != nil && *input.SeriesName != "" {
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
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
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}
	if len(input.Authors) > 0 {
		if err := s.repo.ReplaceBookAuthors(ctx, input.ID, input.Authors, true); err != nil {
			return err
		}
	}
	if input.Genres != nil {
		if err := s.repo.ReplaceBookGenres(ctx, input.ID, input.Genres, true); err != nil {
			return err
		}
	}
	if input.EditionID != "" && input.Edition != nil {
		isbn10, isbn13, err := normalizeEditionISBNs(input.Edition)
		if err != nil {
			return fmt.Errorf("invalid ISBN: %w", err)
		}
		edition := Edition{
			ID:              input.EditionID,
			Title:           strings.TrimSpace(input.Edition.Title),
			OriginalTitle:   strings.TrimSpace(input.Edition.OriginalTitle),
			Format:          input.Edition.Format,
			Description:     input.Edition.Description,
			CoverURL:        input.Edition.CoverURL,
			Language:        input.Edition.Language,
			ISBN10:          isbn10,
			ISBN13:          isbn13,
			ASIN:            input.Edition.ASIN,
			Publisher:       input.Edition.Publisher,
			Edition:         input.Edition.Edition,
			PageCount:       input.Edition.PageCount,
			FileFormat:      input.Edition.FileFormat,
			DurationMinutes: input.Edition.DurationMinutes,
			AudioFormat:     input.Edition.AudioFormat,
		}
		if input.Edition.PublishedAt != nil && *input.Edition.PublishedAt != "" {
			t, err := parsePublishedAt(*input.Edition.PublishedAt)
			if err == nil {
				edition.PublishedAt = &t
			}
		}
		if err := s.repo.UpdateEditionDetails(ctx, input.EditionID, edition); err != nil {
			return err
		}
		if input.Edition.Translators != nil {
			if err := s.repo.ReplaceEditionTranslators(ctx, input.EditionID, input.Edition.Translators); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) DeleteBook(ctx context.Context, id string) error {
	return s.repo.DeleteBook(ctx, id)
}

func (s *Service) ForceDeleteBook(ctx context.Context, id string) error {
	return s.repo.ForceDeleteBook(ctx, id)
}

func (s *Service) DeleteEdition(ctx context.Context, editionID string) error {
	return s.repo.DeleteEdition(ctx, editionID)
}

func (s *Service) GetBooksWithoutCovers(ctx context.Context) ([]BookWithDetails, error) {
	return s.repo.FindBooksWithoutCovers(ctx)
}

func (s *Service) UpdateEditionCoverURL(ctx context.Context, editionID, coverURL string) error {
	return s.repo.UpdateEditionCoverURL(ctx, editionID, coverURL)
}

func (s *Service) FindEditionByID(ctx context.Context, id string) (*Edition, error) {
	return s.repo.FindEditionByID(ctx, id)
}

func (s *Service) GetUserBooks(ctx context.Context, userID string, page int, limit int) (UserBooksResult, error) {
	offset := (page - 1) * limit

	var total int
	if err := s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM book_copies bc
		JOIN book_editions be ON be.id = bc.edition_id
		JOIN books b ON b.id = be.book_id
		WHERE bc.owner_id = $1 AND bc.deleted_at IS NULL AND be.deleted_at IS NULL AND b.deleted_at IS NULL
	`, userID).Scan(&total); err != nil {
		return UserBooksResult{}, fmt.Errorf("failed to count user books: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT
			bc.id, bc.reading_status, bc.current_page, bc.started_reading_at, bc.finished_reading_at,
			bc.owned_by_user, bc.borrowed_from, bc.location, bc.condition, bc.created_at,
			be.id, be.format, be.language, be.cover_url,
			b.id, b.title, b.description, b.series_id, b.series_position, b.status, b.deleted_at, b.created_at, b.updated_at
		FROM book_copies bc
		JOIN book_editions be ON be.id = bc.edition_id
		JOIN books b ON b.id = be.book_id
		WHERE bc.owner_id = $1 AND bc.deleted_at IS NULL AND be.deleted_at IS NULL AND b.deleted_at IS NULL
		ORDER BY bc.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return UserBooksResult{}, fmt.Errorf("failed to list user books: %w", err)
	}
	defer rows.Close()

	books := make([]UserBook, 0)
	bookIDs := make([]string, 0)
	seen := map[string]struct{}{}
	index := map[string]int{}
	for rows.Next() {
		var ub UserBook
		var language string
		if err := rows.Scan(
			&ub.CopyID, &ub.ReadingStatus, &ub.CurrentPage, &ub.StartedReadingAt, &ub.FinishedReadingAt,
			&ub.OwnedByUser, &ub.BorrowedFrom, &ub.Location, &ub.Condition, &ub.AddedAt,
			&ub.EditionID, &ub.Format, &language, &ub.CoverURL,
			&ub.Book.ID, &ub.Book.Title, &ub.Book.Description, &ub.Book.SeriesID, &ub.Book.SeriesPosition,
			&ub.Book.Status, &ub.Book.DeletedAt, &ub.Book.CreatedAt, &ub.Book.UpdatedAt,
		); err != nil {
			return UserBooksResult{}, fmt.Errorf("failed to scan user book: %w", err)
		}
		ub.Language = &language
		ub.Book.Authors = []Author{}
		ub.Book.Genres = []Genre{}
		index[ub.Book.ID] = len(books)
		books = append(books, ub)
		if _, ok := seen[ub.Book.ID]; !ok {
			seen[ub.Book.ID] = struct{}{}
			bookIDs = append(bookIDs, ub.Book.ID)
		}
	}
	if rows.Err() != nil {
		return UserBooksResult{}, rows.Err()
	}

	if len(bookIDs) > 0 {
		aRows, err := s.db.Query(ctx, `
			SELECT bc.book_id, c.id, c.name, c.status, c.deleted_at, c.created_at, c.updated_at
			FROM contributors c
			JOIN book_contributors bc ON bc.contributor_id = c.id
			WHERE bc.book_id = ANY($1) AND bc.role IN ('author','co_author') AND c.deleted_at IS NULL
		`, bookIDs)
		if err == nil {
			for aRows.Next() {
				var bookID string
				var a Author
				if err := aRows.Scan(&bookID, &a.ID, &a.Name, &a.Status, &a.DeletedAt, &a.CreatedAt, &a.UpdatedAt); err == nil {
					for i := range books {
						if books[i].Book.ID == bookID {
							books[i].Book.Authors = append(books[i].Book.Authors, a)
						}
					}
				}
			}
			aRows.Close()
		}
	}

	return UserBooksResult{Books: books, Total: total, Page: page, Limit: limit}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
