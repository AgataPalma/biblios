package collections

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type fakeCollectionRepository struct {
	access           *LibraryAccess
	collection       *Collection
	collections      []Collection
	books            []CollectionBook
	copyBelongs      bool
	listPublicOnly   bool
	listCalled       bool
	addCalled        bool
	removeCalled     bool
	createCalled     bool
	lastLibraryID    string
	lastCollectionID string
	lastCopyID       string
}

func (f *fakeCollectionRepository) Create(_ context.Context, libraryID, _ string, _ string, _, _ *string, _, _ bool) (Collection, error) {
	f.createCalled = true
	f.lastLibraryID = libraryID
	return Collection{LibraryID: libraryID}, nil
}

func (f *fakeCollectionRepository) GetLibraryAccess(_ context.Context, libraryID, _ string) (*LibraryAccess, error) {
	f.lastLibraryID = libraryID
	return f.access, nil
}

func (f *fakeCollectionRepository) FindByID(_ context.Context, libraryID, collectionID string) (*Collection, error) {
	f.lastLibraryID = libraryID
	f.lastCollectionID = collectionID
	return f.collection, nil
}

func (f *fakeCollectionRepository) ListByLibrary(_ context.Context, libraryID string, publicOnly bool) ([]Collection, error) {
	f.listCalled = true
	f.lastLibraryID = libraryID
	f.listPublicOnly = publicOnly
	return f.collections, nil
}

func (f *fakeCollectionRepository) Update(_ context.Context, id string, _, _ *string, _ *bool) (Collection, error) {
	f.lastCollectionID = id
	return Collection{ID: id}, nil
}

func (f *fakeCollectionRepository) Delete(_ context.Context, id string) error {
	f.lastCollectionID = id
	return nil
}

func (f *fakeCollectionRepository) CopyBelongsToLibrary(_ context.Context, libraryID, copyID string) (bool, error) {
	f.lastLibraryID = libraryID
	f.lastCopyID = copyID
	return f.copyBelongs, nil
}

func (f *fakeCollectionRepository) AddBook(_ context.Context, collectionID, copyID, _ string) error {
	f.addCalled = true
	f.lastCollectionID = collectionID
	f.lastCopyID = copyID
	return nil
}

func (f *fakeCollectionRepository) RemoveBook(_ context.Context, collectionID, copyID string) error {
	f.removeCalled = true
	f.lastCollectionID = collectionID
	f.lastCopyID = copyID
	return nil
}

func (f *fakeCollectionRepository) ListBooks(_ context.Context, collectionID string, _, _ int) ([]CollectionBook, int, error) {
	f.lastCollectionID = collectionID
	return f.books, len(f.books), nil
}

func TestListByLibraryEnforcesVisibility(t *testing.T) {
	tests := []struct {
		name           string
		access         *LibraryAccess
		wantErr        error
		wantList       bool
		wantPublicOnly bool
	}{
		{
			name:     "member with view permission sees all collections",
			access:   &LibraryAccess{Visibility: "private", IsMember: true, CanView: true},
			wantList: true,
		},
		{
			name:           "public visitor sees only public collections",
			access:         &LibraryAccess{Visibility: "public"},
			wantList:       true,
			wantPublicOnly: true,
		},
		{
			name:    "private visitor is denied",
			access:  &LibraryAccess{Visibility: "private"},
			wantErr: ErrAccessDenied,
		},
		{
			name:    "private member without view permission is denied",
			access:  &LibraryAccess{Visibility: "private", IsMember: true},
			wantErr: ErrAccessDenied,
		},
		{
			name:           "public member without view permission gets public subset",
			access:         &LibraryAccess{Visibility: "public", IsMember: true},
			wantList:       true,
			wantPublicOnly: true,
		},
		{
			name:    "missing library is not found",
			access:  nil,
			wantErr: ErrLibraryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCollectionRepository{access: tt.access}
			service := NewService(repo)

			_, err := service.ListByLibrary(context.Background(), "library-1", "user-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListByLibrary() error = %v, want %v", err, tt.wantErr)
			}
			if repo.listCalled != tt.wantList {
				t.Fatalf("ListByLibrary() repository call = %v, want %v", repo.listCalled, tt.wantList)
			}
			if repo.listCalled && repo.listPublicOnly != tt.wantPublicOnly {
				t.Fatalf("ListByLibrary() publicOnly = %v, want %v", repo.listPublicOnly, tt.wantPublicOnly)
			}
		})
	}
}

func TestGetScopesCollectionToLibraryAndVisibility(t *testing.T) {
	tests := []struct {
		name       string
		access     *LibraryAccess
		collection *Collection
		wantErr    error
	}{
		{
			name:       "public visitor can read public collection",
			access:     &LibraryAccess{Visibility: "public"},
			collection: &Collection{ID: "collection-1", LibraryID: "library-1", IsPublic: true},
		},
		{
			name:       "public visitor cannot read private collection",
			access:     &LibraryAccess{Visibility: "public"},
			collection: &Collection{ID: "collection-1", LibraryID: "library-1"},
			wantErr:    ErrAccessDenied,
		},
		{
			name:       "viewing member can read private collection",
			access:     &LibraryAccess{Visibility: "private", IsMember: true, CanView: true},
			collection: &Collection{ID: "collection-1", LibraryID: "library-1"},
		},
		{
			name:    "collection outside nested library is not found",
			access:  &LibraryAccess{Visibility: "private", IsMember: true, CanView: true},
			wantErr: ErrCollectionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCollectionRepository{access: tt.access, collection: tt.collection}
			service := NewService(repo)

			_, err := service.Get(context.Background(), "library-1", "collection-1", "user-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Get() error = %v, want %v", err, tt.wantErr)
			}
			if repo.lastLibraryID != "library-1" || repo.lastCollectionID != "collection-1" {
				t.Fatalf("Get() scope = %q/%q, want library-1/collection-1", repo.lastLibraryID, repo.lastCollectionID)
			}
		})
	}
}

func TestCreateRequiresEditPermission(t *testing.T) {
	tests := []struct {
		name       string
		access     *LibraryAccess
		wantErr    error
		wantCreate bool
	}{
		{
			name:       "editor can create",
			access:     &LibraryAccess{IsMember: true, CanEdit: true},
			wantCreate: true,
		},
		{
			name:    "member without edit permission is denied",
			access:  &LibraryAccess{IsMember: true, CanView: true},
			wantErr: ErrAccessDenied,
		},
		{
			name:    "public visitor is denied",
			access:  &LibraryAccess{Visibility: "public"},
			wantErr: ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCollectionRepository{access: tt.access}
			service := NewService(repo)

			_, err := service.Create(context.Background(), "library-1", "user-1", CreateCollectionInput{Name: "Favourites"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
			}
			if repo.createCalled != tt.wantCreate {
				t.Fatalf("Create() repository call = %v, want %v", repo.createCalled, tt.wantCreate)
			}
		})
	}
}

func TestAddBookEnforcesMembershipPermissionAndLibraryScope(t *testing.T) {
	tests := []struct {
		name        string
		access      *LibraryAccess
		collection  *Collection
		copyBelongs bool
		wantErr     error
		wantAdd     bool
	}{
		{
			name:        "collaborating member can add library copy",
			access:      &LibraryAccess{IsMember: true, CanAdd: true},
			collection:  &Collection{ID: "collection-1", CreatedBy: "owner", IsCollaborative: true},
			copyBelongs: true,
			wantAdd:     true,
		},
		{
			name:        "copy from another library is rejected",
			access:      &LibraryAccess{IsMember: true, CanAdd: true},
			collection:  &Collection{ID: "collection-1", CreatedBy: "owner", IsCollaborative: true},
			copyBelongs: false,
			wantErr:     ErrCopyNotInLibrary,
		},
		{
			name:       "member without add permission is denied",
			access:     &LibraryAccess{IsMember: true, CanView: true},
			collection: &Collection{ID: "collection-1", CreatedBy: "owner", IsCollaborative: true},
			wantErr:    ErrAccessDenied,
		},
		{
			name:       "non-creator cannot add to private collection",
			access:     &LibraryAccess{IsMember: true, CanAdd: true},
			collection: &Collection{ID: "collection-1", CreatedBy: "owner"},
			wantErr:    ErrAccessDenied,
		},
		{
			name:        "creator can add to private collection",
			access:      &LibraryAccess{IsMember: true, CanAdd: true},
			collection:  &Collection{ID: "collection-1", CreatedBy: "user-1"},
			copyBelongs: true,
			wantAdd:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCollectionRepository{
				access:      tt.access,
				collection:  tt.collection,
				copyBelongs: tt.copyBelongs,
			}
			service := NewService(repo)

			err := service.AddBook(context.Background(), "library-1", "collection-1", "user-1", "copy-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AddBook() error = %v, want %v", err, tt.wantErr)
			}
			if repo.addCalled != tt.wantAdd {
				t.Fatalf("AddBook() repository call = %v, want %v", repo.addCalled, tt.wantAdd)
			}
		})
	}
}

func TestUpdateRequiresCurrentEditorAndCreatorOrOwner(t *testing.T) {
	tests := []struct {
		name      string
		access    *LibraryAccess
		createdBy string
		wantErr   error
	}{
		{
			name:      "creator with edit permission can update",
			access:    &LibraryAccess{IsMember: true, CanEdit: true},
			createdBy: "user-1",
		},
		{
			name:      "library owner can manage collection",
			access:    &LibraryAccess{IsMember: true, IsOwner: true, CanEdit: true},
			createdBy: "other-user",
		},
		{
			name:      "former creator is denied",
			access:    &LibraryAccess{CanEdit: true},
			createdBy: "user-1",
			wantErr:   ErrAccessDenied,
		},
		{
			name:      "unrelated editor is denied",
			access:    &LibraryAccess{IsMember: true, CanEdit: true},
			createdBy: "other-user",
			wantErr:   ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCollectionRepository{
				access:     tt.access,
				collection: &Collection{ID: "collection-1", CreatedBy: tt.createdBy},
			}
			service := NewService(repo)

			_, err := service.Update(context.Background(), "library-1", "collection-1", "user-1", UpdateCollectionInput{})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Update() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveBookRequiresRemovePermission(t *testing.T) {
	tests := []struct {
		name       string
		access     *LibraryAccess
		collection *Collection
		wantErr    error
		wantRemove bool
	}{
		{
			name:       "collaborator with remove permission can remove",
			access:     &LibraryAccess{IsMember: true, CanRemove: true},
			collection: &Collection{ID: "collection-1", IsCollaborative: true},
			wantRemove: true,
		},
		{
			name:       "collaborator without remove permission is denied",
			access:     &LibraryAccess{IsMember: true},
			collection: &Collection{ID: "collection-1", IsCollaborative: true},
			wantErr:    ErrAccessDenied,
		},
		{
			name:       "unrelated member cannot remove from private collection",
			access:     &LibraryAccess{IsMember: true, CanRemove: true},
			collection: &Collection{ID: "collection-1", CreatedBy: "other-user"},
			wantErr:    ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCollectionRepository{access: tt.access, collection: tt.collection}
			service := NewService(repo)

			err := service.RemoveBook(context.Background(), "library-1", "collection-1", "user-1", "copy-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RemoveBook() error = %v, want %v", err, tt.wantErr)
			}
			if repo.removeCalled != tt.wantRemove {
				t.Fatalf("RemoveBook() repository call = %v, want %v", repo.removeCalled, tt.wantRemove)
			}
		})
	}
}

func TestPublicCollectionViewsRedactInternalUserAndCopyIDs(t *testing.T) {
	repo := &fakeCollectionRepository{
		access:     &LibraryAccess{Visibility: "public"},
		collection: &Collection{ID: "collection-1", IsPublic: true, CreatedBy: "private-user-id"},
		books:      []CollectionBook{{CollectionID: "collection-1", BookCopyID: "private-copy-id", BookID: "book-1"}},
	}
	service := NewService(repo)

	collection, err := service.Get(context.Background(), "library-1", "collection-1", "visitor")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if collection.CreatedBy != "" {
		t.Fatalf("Get() exposed creator ID %q", collection.CreatedBy)
	}

	books, _, err := service.ListBooks(context.Background(), "library-1", "collection-1", "visitor", 1, 20)
	if err != nil {
		t.Fatalf("ListBooks() error = %v", err)
	}
	if books[0].BookCopyID != "" {
		t.Fatalf("ListBooks() exposed copy ID %q", books[0].BookCopyID)
	}
}

func TestCollectionBookJSONDoesNotExposePersonalCopyState(t *testing.T) {
	payload, err := json.Marshal(CollectionBook{
		CollectionID: "collection-1",
		BookCopyID:   "copy-1",
		BookID:       "book-1",
		Title:        "A Book",
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	for _, privateField := range []string{
		"reading_status",
		"current_page",
		"started_reading_at",
		"finished_reading_at",
		"borrowed_from",
		"location",
		"personal_notes",
		"reread_count",
	} {
		if strings.Contains(string(payload), privateField) {
			t.Fatalf("collection payload leaked private field %q: %s", privateField, payload)
		}
	}
}
