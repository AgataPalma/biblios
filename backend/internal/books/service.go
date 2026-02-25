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
