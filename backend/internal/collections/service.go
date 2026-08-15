package collections

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrLibraryNotFound     = errors.New("library not found")
	ErrCollectionNotFound  = errors.New("collection not found")
	ErrAccessDenied        = errors.New("access denied")
	ErrCopyNotInLibrary    = errors.New("copy does not belong to this library")
	ErrBookNotInCollection = errors.New("book not found in collection")
	ErrNameRequired        = errors.New("name is required")
	ErrNameEmpty           = errors.New("name cannot be empty")
)

type collectionRepository interface {
	Create(ctx context.Context, libraryID, createdBy, name string, description, coverColour *string, isPublic, isCollaborative bool) (Collection, error)
	GetLibraryAccess(ctx context.Context, libraryID, userID string) (*LibraryAccess, error)
	FindByID(ctx context.Context, libraryID, id string) (*Collection, error)
	ListByLibrary(ctx context.Context, libraryID string, publicOnly bool) ([]Collection, error)
	Update(ctx context.Context, id string, name, description *string, isPublic *bool) (Collection, error)
	Delete(ctx context.Context, id string) error
	CopyBelongsToLibrary(ctx context.Context, libraryID, copyID string) (bool, error)
	AddBook(ctx context.Context, collectionID, copyID, addedBy string) error
	RemoveBook(ctx context.Context, collectionID, copyID string) error
	ListBooks(ctx context.Context, collectionID string, page, limit int) ([]CollectionBook, int, error)
}

type Service struct {
	repo collectionRepository
}

func NewService(repo collectionRepository) *Service {
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

func (s *Service) libraryAccess(ctx context.Context, libraryID, userID string) (*LibraryAccess, error) {
	access, err := s.repo.GetLibraryAccess(ctx, libraryID, userID)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, ErrLibraryNotFound
	}
	return access, nil
}

func canViewAllCollections(access *LibraryAccess) bool {
	return access.IsMember && access.CanView
}

func canViewPublicCollections(access *LibraryAccess) bool {
	return access.Visibility == "public"
}

func canViewCollection(access *LibraryAccess, collection *Collection) bool {
	return canViewAllCollections(access) || (canViewPublicCollections(access) && collection.IsPublic)
}

func publicCollectionView(collection Collection) Collection {
	collection.CreatedBy = ""
	return collection
}

func (s *Service) findCollection(ctx context.Context, libraryID, collectionID string) (*Collection, error) {
	collection, err := s.repo.FindByID(ctx, libraryID, collectionID)
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, ErrCollectionNotFound
	}
	return collection, nil
}

func (s *Service) Create(ctx context.Context, libraryID, userID string, input CreateCollectionInput) (Collection, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Collection{}, ErrNameRequired
	}

	access, err := s.libraryAccess(ctx, libraryID, userID)
	if err != nil {
		return Collection{}, err
	}
	if !access.IsMember || !access.CanEdit {
		return Collection{}, ErrAccessDenied
	}

	return s.repo.Create(ctx, libraryID, userID, input.Name, input.Description, input.CoverColour, input.IsPublic, input.IsCollaborative)
}

func (s *Service) Get(ctx context.Context, libraryID, collectionID, userID string) (*Collection, error) {
	access, err := s.libraryAccess(ctx, libraryID, userID)
	if err != nil {
		return nil, err
	}
	collection, err := s.findCollection(ctx, libraryID, collectionID)
	if err != nil {
		return nil, err
	}
	if !canViewCollection(access, collection) {
		return nil, ErrAccessDenied
	}
	if !canViewAllCollections(access) {
		publicView := publicCollectionView(*collection)
		return &publicView, nil
	}
	return collection, nil
}

func (s *Service) ListByLibrary(ctx context.Context, libraryID, userID string) ([]Collection, error) {
	access, err := s.libraryAccess(ctx, libraryID, userID)
	if err != nil {
		return nil, err
	}

	if canViewAllCollections(access) {
		return s.repo.ListByLibrary(ctx, libraryID, false)
	}
	if canViewPublicCollections(access) {
		collections, err := s.repo.ListByLibrary(ctx, libraryID, true)
		if err != nil {
			return nil, err
		}
		for i := range collections {
			collections[i] = publicCollectionView(collections[i])
		}
		return collections, nil
	}
	return nil, ErrAccessDenied
}

func (s *Service) Update(ctx context.Context, libraryID, collectionID, userID string, input UpdateCollectionInput) (Collection, error) {
	access, err := s.libraryAccess(ctx, libraryID, userID)
	if err != nil {
		return Collection{}, err
	}
	collection, err := s.findCollection(ctx, libraryID, collectionID)
	if err != nil {
		return Collection{}, err
	}
	if !access.IsMember || !access.CanEdit || (collection.CreatedBy != userID && !access.IsOwner) {
		return Collection{}, ErrAccessDenied
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Collection{}, ErrNameEmpty
	}
	return s.repo.Update(ctx, collectionID, input.Name, input.Description, input.IsPublic)
}

func (s *Service) Delete(ctx context.Context, libraryID, collectionID, userID string) error {
	access, err := s.libraryAccess(ctx, libraryID, userID)
	if err != nil {
		return err
	}
	collection, err := s.findCollection(ctx, libraryID, collectionID)
	if err != nil {
		return err
	}
	if !access.IsMember || !access.CanEdit || (collection.CreatedBy != userID && !access.IsOwner) {
		return ErrAccessDenied
	}
	return s.repo.Delete(ctx, collectionID)
}

// AddBook adds a copy that is already assigned to the collection's library.
// Collaborative collections allow members with can_add permission to add;
// otherwise the creator or library owner must perform the operation.
func (s *Service) AddBook(ctx context.Context, libraryID, collectionID, userID, copyID string) error {
	access, err := s.libraryAccess(ctx, libraryID, userID)
	if err != nil {
		return err
	}
	collection, err := s.findCollection(ctx, libraryID, collectionID)
	if err != nil {
		return err
	}
	if !access.IsMember || !access.CanAdd {
		return ErrAccessDenied
	}
	if !collection.IsCollaborative && collection.CreatedBy != userID && !access.IsOwner {
		return ErrAccessDenied
	}

	belongs, err := s.repo.CopyBelongsToLibrary(ctx, libraryID, copyID)
	if err != nil {
		return err
	}
	if !belongs {
		return ErrCopyNotInLibrary
	}
	return s.repo.AddBook(ctx, collectionID, copyID, userID)
}

func (s *Service) RemoveBook(ctx context.Context, libraryID, collectionID, userID, copyID string) error {
	access, err := s.libraryAccess(ctx, libraryID, userID)
	if err != nil {
		return err
	}
	collection, err := s.findCollection(ctx, libraryID, collectionID)
	if err != nil {
		return err
	}
	if !access.IsMember || !access.CanRemove {
		return ErrAccessDenied
	}
	if !collection.IsCollaborative && collection.CreatedBy != userID && !access.IsOwner {
		return ErrAccessDenied
	}
	return s.repo.RemoveBook(ctx, collectionID, copyID)
}

func (s *Service) ListBooks(ctx context.Context, libraryID, collectionID, userID string, page, limit int) ([]CollectionBook, int, error) {
	access, err := s.libraryAccess(ctx, libraryID, userID)
	if err != nil {
		return nil, 0, err
	}
	collection, err := s.findCollection(ctx, libraryID, collectionID)
	if err != nil {
		return nil, 0, err
	}
	if !canViewCollection(access, collection) {
		return nil, 0, ErrAccessDenied
	}
	books, total, err := s.repo.ListBooks(ctx, collectionID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	if !canViewAllCollections(access) {
		for i := range books {
			books[i].BookCopyID = ""
		}
	}
	return books, total, nil
}
