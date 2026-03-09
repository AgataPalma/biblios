package books

import (
	"context"
	"errors"
	"fmt"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bookRepository interface {
	UpdateBook(ctx context.Context, id string, title, description, coverURL *string) error
	FindUserBooksWithCopies(ctx context.Context, userID string, page, limit int) ([]UserBook, int, error)
	UpdateReadingStatus(ctx context.Context, copyID, userID, status string) error
	RemoveCopy(ctx context.Context, copyID, userID string) error
	ListApprovedBooks(ctx context.Context, page, limit int, sort string) ([]Book, int, error)
	FindBookWithDetails(ctx context.Context, id string) (*Book, error)
	FindUserBooks(ctx context.Context, userID string, page, limit int) ([]Book, int, error)
	FindBooksWithoutCovers(ctx context.Context) ([]BookWithDetails, error)
	UpdateCoverURL(ctx context.Context, bookID, coverURL string) error
	FindEditionByISBN(ctx context.Context, isbn string) (*Edition, error)
	FindBookByID(ctx context.Context, id string) (*Book, error)
	FindSubmissionByID(ctx context.Context, id string) (*Submission, error)
	DeleteBook(ctx context.Context, id string) error
	withDB(db DB) *txRepository // needed so SubmitBook can create a txRepository
	FindEditionByID(ctx context.Context, id string) (*Edition, error)
}

type Service struct {
	repo bookRepository
	db   *pgxpool.Pool
}

func NewService(repo *Repository, db *pgxpool.Pool) *Service {
	return &Service{repo: repo, db: db}
}

type SubmitBookInput struct {
	Title         string
	Description   *string
	CoverURL      *string
	Authors       []string
	Genres        []string
	Edition       EditionInput
	Condition     *string
	UserID        string
	UserRole      apictx.Role
	CatalogueOnly bool // when true, book is added to catalogue only — no personal copy created
}

type EditionInput struct {
	Format          string
	ISBN            *string
	ASIN            *string
	Language        string
	Publisher       *string
	Edition         *string
	PublishedAt     *string
	PageCount       *int
	FileFormat      *string
	DurationMinutes *int
	AudioFormat     *string
	Narrator        *string  // single name — backend does FindOrCreate + link
	Translators     []string // multiple names — each gets FindOrCreate + link
}

type SubmitBookResult struct {
	Submission Submission
	Book       Book
	Edition    Edition
	Copy       *Copy
}

type AddCopyInput struct {
	EditionID string
	Condition *string
	UserID    string
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
	ID          string
	Title       *string
	Description *string
	CoverURL    *string
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
	var copy Copy
	copy, err = repo.InsertCopy(ctx, edition.ID, input.UserID, input.Condition)
	if err != nil {
		return AddCopyResult{}, err
	}

	// Create approved submission immediately since edition already exists and is approved
	var submission Submission
	submission, err = repo.InsertSubmissionApproved(ctx, input.UserID, book.ID, edition.ID, copy.ID)
	if err != nil {
		return AddCopyResult{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return AddCopyResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result.Copy = copy
	result.Edition = *edition
	result.Book = *book
	result.Submission = submission

	return result, nil
}

func (s *Service) FindExistingEditionByISBN(ctx context.Context, isbn string) (*Edition, error) {
	return s.repo.FindEditionByISBN(ctx, isbn)
}

func (s *Service) SubmitBook(ctx context.Context, input SubmitBookInput) (SubmitBookResult, error) {
	var autoApprove bool = CanAutoApprove(input.UserRole)
	var result SubmitBookResult

	// Start transaction
	var tx pgxTx
	var err error
	tx, err = s.db.Begin(ctx)
	if err != nil {
		return SubmitBookResult{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// All operations go through the transaction
	var repo *txRepository = s.repo.withDB(tx)

	// Create book
	var book Book
	book, err = repo.InsertBook(ctx, input.Title, input.Description, input.CoverURL, autoApprove)
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

	// Create edition
	var edition Edition = Edition{
		BookID:          book.ID,
		Format:          input.Edition.Format,
		ISBN:            input.Edition.ISBN,
		ASIN:            input.Edition.ASIN,
		Language:        input.Edition.Language,
		Publisher:       input.Edition.Publisher,
		Edition:         input.Edition.Edition,
		PageCount:       input.Edition.PageCount,
		FileFormat:      input.Edition.FileFormat,
		DurationMinutes: input.Edition.DurationMinutes,
		AudioFormat:     input.Edition.AudioFormat,
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
		var copy Copy
		copy, err = repo.InsertCopy(ctx, edition.ID, input.UserID, input.Condition)
		if err != nil {
			return SubmitBookResult{}, err
		}
		result.Copy = &copy

		var submission Submission
		submission, err = repo.InsertSubmissionApproved(ctx, input.UserID, book.ID, edition.ID, copy.ID)
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
	return s.repo.UpdateBook(ctx, input.ID, input.Title, input.Description, input.CoverURL)
}

func (s *Service) DeleteBook(ctx context.Context, id string) error {
	return s.repo.DeleteBook(ctx, id)
}

func (s *Service) GetUserBooks(ctx context.Context, userID string, page int, limit int) (ListBooksResult, error) {
	var bookList []Book
	var total int
	var err error

	bookList, total, err = s.repo.FindUserBooks(ctx, userID, page, limit)
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

func (s *Service) GetBooksWithoutCovers(ctx context.Context) ([]BookWithDetails, error) {
	return s.repo.FindBooksWithoutCovers(ctx)
}

func (s *Service) UpdateCoverURL(ctx context.Context, bookID, coverURL string) error {
	return s.repo.UpdateCoverURL(ctx, bookID, coverURL)
}

func (s *Service) UpdateReadingStatus(ctx context.Context, copyID, userID, status string) error {
	return s.repo.UpdateReadingStatus(ctx, copyID, userID, status)
}

func (s *Service) RemoveCopy(ctx context.Context, copyID, userID string) error {
	return s.repo.RemoveCopy(ctx, copyID, userID)
}

func (s *Service) GetMyLibrary(ctx context.Context, userID string, page, limit int) ([]UserBook, int, error) {
	books, total, err := s.repo.FindUserBooksWithCopies(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	if books == nil {
		books = []UserBook{}
	}
	return books, total, nil
}

func (s *Service) FindEditionByID(ctx context.Context, id string) (*Edition, error) {
	return s.repo.FindEditionByID(ctx, id)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
