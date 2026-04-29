package shelves

import (
	"context"
	"fmt"
	"strings"

	"github.com/AgataPalma/biblios/internal/books"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID, name string) (Shelf, error) {
	if strings.TrimSpace(name) == "" {
		return Shelf{}, fmt.Errorf("name is required")
	}
	return s.repo.Create(ctx, userID, name)
}

func (s *Service) ListMyShelf(ctx context.Context, userID string) ([]Shelf, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Rename(ctx context.Context, id, userID, name string) (Shelf, error) {
	if strings.TrimSpace(name) == "" {
		return Shelf{}, fmt.Errorf("name cannot be empty")
	}
	return s.repo.Rename(ctx, id, userID, name)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) AddBook(ctx context.Context, shelfID, userID, copyID string) error {
	shelf, err := s.repo.FindByID(ctx, shelfID)
	if err != nil {
		return err
	}
	if shelf == nil {
		return fmt.Errorf("shelf not found")
	}
	if shelf.UserID != userID {
		return fmt.Errorf("shelf not found")
	}
	return s.repo.AddBook(ctx, shelfID, copyID)
}

func (s *Service) RemoveBook(ctx context.Context, shelfID, userID, copyID string) error {
	shelf, err := s.repo.FindByID(ctx, shelfID)
	if err != nil {
		return err
	}
	if shelf == nil || shelf.UserID != userID {
		return fmt.Errorf("shelf not found")
	}
	return s.repo.RemoveBook(ctx, shelfID, copyID)
}

func (s *Service) ListBooks(ctx context.Context, shelfID, userID string, page, limit int) ([]books.UserBook, int, error) {
	shelf, err := s.repo.FindByID(ctx, shelfID)
	if err != nil {
		return nil, 0, err
	}
	if shelf == nil || shelf.UserID != userID {
		return nil, 0, fmt.Errorf("shelf not found")
	}
	return s.repo.ListBooks(ctx, shelfID, page, limit)
}
