package shelves

import (
	"context"
	"errors"
	"strings"

	"github.com/AgataPalma/biblios/internal/books"
)

var (
	ErrShelfNotFound  = errors.New("shelf not found")
	ErrBookNotOnShelf = errors.New("book not found on shelf")
	ErrCopyNotOwned   = errors.New("copy not found")
	ErrNameRequired   = errors.New("name is required")
	ErrNameEmpty      = errors.New("name cannot be empty")
)

type shelfRepository interface {
	Create(ctx context.Context, userID, name string) (Shelf, error)
	FindByID(ctx context.Context, id, userID string) (*Shelf, error)
	ListByUser(ctx context.Context, userID string) ([]Shelf, error)
	Rename(ctx context.Context, id, userID, name string) (Shelf, error)
	Delete(ctx context.Context, id, userID string) error
	AddBook(ctx context.Context, shelfID, userID, copyID string) error
	RemoveBook(ctx context.Context, shelfID, userID, copyID string) error
	ListBooks(ctx context.Context, shelfID, userID string, page, limit int) ([]books.UserBook, int, error)
}

type Service struct {
	repo shelfRepository
}

func NewService(repo shelfRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID, name string) (Shelf, error) {
	if strings.TrimSpace(name) == "" {
		return Shelf{}, ErrNameRequired
	}
	return s.repo.Create(ctx, userID, name)
}

func (s *Service) ListMyShelf(ctx context.Context, userID string) ([]Shelf, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Rename(ctx context.Context, id, userID, name string) (Shelf, error) {
	if strings.TrimSpace(name) == "" {
		return Shelf{}, ErrNameEmpty
	}
	return s.repo.Rename(ctx, id, userID, name)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) personalShelf(ctx context.Context, shelfID, userID string) (*Shelf, error) {
	shelf, err := s.repo.FindByID(ctx, shelfID, userID)
	if err != nil {
		return nil, err
	}
	if shelf == nil {
		return nil, ErrShelfNotFound
	}
	return shelf, nil
}

func (s *Service) AddBook(ctx context.Context, shelfID, userID, copyID string) error {
	if _, err := s.personalShelf(ctx, shelfID, userID); err != nil {
		return err
	}
	return s.repo.AddBook(ctx, shelfID, userID, copyID)
}

func (s *Service) RemoveBook(ctx context.Context, shelfID, userID, copyID string) error {
	if _, err := s.personalShelf(ctx, shelfID, userID); err != nil {
		return err
	}
	return s.repo.RemoveBook(ctx, shelfID, userID, copyID)
}

func (s *Service) ListBooks(ctx context.Context, shelfID, userID string, page, limit int) ([]books.UserBook, int, error) {
	if _, err := s.personalShelf(ctx, shelfID, userID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListBooks(ctx, shelfID, userID, page, limit)
}
