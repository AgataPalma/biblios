package books

import (
	"context"
	"fmt"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *Repository
	db   *pgxpool.Pool
}

func NewService(repo *Repository, db *pgxpool.Pool) *Service {
	return &Service{repo: repo, db: db}
}

type SubmitBookInput struct {
	Title       string
	Description *string
	CoverURL    *string
	Authors     []string
	Genres      []string
	Edition     EditionInput
	Condition   *string
	UserID      string
	UserRole    apictx.Role
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

func (s *Service) AddCopyOfExistingEdition(ctx context.Context, input AddCopyInput) (AddCopyResult, error) {
	var result AddCopyResult
	var err error

	// Check edition exists and is approved
	var edition *Edition
	edition, err = s.repo.FindEditionByISBN(ctx, input.EditionID)
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
	defer tx.Rollback(ctx)

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
	defer tx.Rollback(ctx)

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
		return SubmitBookResult{}, err
	}

	result.Book = book
	result.Edition = edition

	// If auto approve, create copy immediately
	if autoApprove {
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

type ListBooksResult struct {
	Books []Book `json:"books"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

func (s *Service) ListBooks(ctx context.Context, page int, limit int) (ListBooksResult, error) {
	var bookList []Book
	var total int
	var err error

	bookList, total, err = s.repo.ListApprovedBooks(ctx, page, limit)
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

type UpdateBookInput struct {
	ID          string
	Title       string
	Description *string
	CoverURL    *string
}

func (s *Service) UpdateBook(ctx context.Context, input UpdateBookInput) error {
	if input.Title == "" {
		return fmt.Errorf("title is required")
	}
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
