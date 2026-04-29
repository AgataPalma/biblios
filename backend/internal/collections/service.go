package collections

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

type CreateCollectionInput struct {
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	CoverColour     *string `json:"cover_colour"`
	IsPublic        bool    `json:"is_public"`
	IsCollaborative bool    `json:"is_collaborative"`
}

type UpdateCollectionInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
}

func (s *Service) Create(ctx context.Context, libraryID, userID string, input CreateCollectionInput) (Collection, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Collection{}, fmt.Errorf("name is required")
	}
	return s.repo.Create(ctx, libraryID, userID, input.Name, input.Description, input.CoverColour, input.IsPublic, input.IsCollaborative)
}

func (s *Service) Get(ctx context.Context, id string) (*Collection, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByLibrary(ctx context.Context, libraryID string) ([]Collection, error) {
	return s.repo.ListByLibrary(ctx, libraryID)
}

func (s *Service) Update(ctx context.Context, id, userID string, input UpdateCollectionInput) (Collection, error) {
	col, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Collection{}, err
	}
	if col == nil {
		return Collection{}, fmt.Errorf("collection not found")
	}
	if col.CreatedBy != userID {
		return Collection{}, fmt.Errorf("only the collection creator can update it")
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Collection{}, fmt.Errorf("name cannot be empty")
	}
	return s.repo.Update(ctx, id, input.Name, input.Description, input.IsPublic)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	col, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if col == nil {
		return fmt.Errorf("collection not found")
	}
	if col.CreatedBy != userID {
		return fmt.Errorf("only the collection creator can delete it")
	}
	return s.repo.Delete(ctx, id)
}

// AddBook adds a copy to a collection.
// - If collaborative: any library member can add.
// - If not collaborative: only the creator can add.
func (s *Service) AddBook(ctx context.Context, collectionID, userID, copyID string) error {
	col, err := s.repo.FindByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if col == nil {
		return fmt.Errorf("collection not found")
	}
	if !col.IsCollaborative && col.CreatedBy != userID {
		return fmt.Errorf("only the collection creator can add books to this collection")
	}
	return s.repo.AddBook(ctx, collectionID, copyID, userID)
}

func (s *Service) RemoveBook(ctx context.Context, collectionID, userID, copyID string) error {
	col, err := s.repo.FindByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if col == nil {
		return fmt.Errorf("collection not found")
	}
	if col.CreatedBy != userID {
		return fmt.Errorf("only the collection creator can remove books")
	}
	return s.repo.RemoveBook(ctx, collectionID, copyID)
}

func (s *Service) ListBooks(ctx context.Context, collectionID string, page, limit int) ([]books.UserBook, int, error) {
	return s.repo.ListBooks(ctx, collectionID, page, limit)
}
