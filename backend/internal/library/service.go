package library

import (
	"context"
	"github.com/AgataPalma/biblios/internal/books"
	"github.com/jackc/pgx/v5/pgxpool"
)

type editionLookup interface {
	FindEditionByID(ctx context.Context, id string) (*books.Edition, error)
	FindBookByID(ctx context.Context, id string) (*books.Book, error)
}

type RepositoryInterface interface {
	InsertCopy(ctx context.Context, editionID, ownerID string, condition *string, opts CopyOptions) (Copy, error)
	UpdateReadingStatus(ctx context.Context, copyID, userID string, input UpdateCopyInput) error
	RemoveCopy(ctx context.Context, copyID, userID string) error
	FindUserLibrary(ctx context.Context, userID string, page, limit int) ([]UserBook, int, error)
}

type Service struct {
	repo     RepositoryInterface
	editions editionLookup
	db       *pgxpool.Pool
}

func NewService(repo RepositoryInterface, editions editionLookup, db *pgxpool.Pool) *Service {
	return &Service{repo: repo, editions: editions, db: db}
}

func (s *Service) GetMyLibrary(ctx context.Context, userID string, page, limit int) (ListLibraryResult, error) {
	books, total, err := s.repo.FindUserLibrary(ctx, userID, page, limit)
	if err != nil {
		return ListLibraryResult{}, err
	}
	if books == nil {
		books = []UserBook{}
	}
	return ListLibraryResult{
		Books: books,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *Service) UpdateReadingStatus(ctx context.Context, copyID, userID string, input UpdateCopyInput) error {
	return s.repo.UpdateReadingStatus(ctx, copyID, userID, input)
}

func (s *Service) RemoveCopy(ctx context.Context, copyID, userID string) error {
	return s.repo.RemoveCopy(ctx, copyID, userID)
}

// AddCopyOfExistingEdition can stay in books.Service for now; later you can move it here
// and inject edition/book lookup via the editionLookup interface.
