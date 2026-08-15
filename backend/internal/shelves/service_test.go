package shelves

import (
	"context"
	"errors"
	"testing"

	"github.com/AgataPalma/biblios/internal/books"
)

type fakeShelfRepository struct {
	shelf           *Shelf
	shelves         []Shelf
	books           []books.UserBook
	findErr         error
	addErr          error
	removeErr       error
	addCalled       bool
	removeCalled    bool
	listBooksCalled bool
	lastShelfID     string
	lastUserID      string
	lastCopyID      string
}

func (f *fakeShelfRepository) Create(_ context.Context, userID, name string) (Shelf, error) {
	return Shelf{UserID: userID, Name: name}, nil
}

func (f *fakeShelfRepository) FindByID(_ context.Context, id, userID string) (*Shelf, error) {
	f.lastShelfID = id
	f.lastUserID = userID
	return f.shelf, f.findErr
}

func (f *fakeShelfRepository) ListByUser(_ context.Context, userID string) ([]Shelf, error) {
	f.lastUserID = userID
	return f.shelves, nil
}

func (f *fakeShelfRepository) Rename(_ context.Context, id, userID, name string) (Shelf, error) {
	f.lastShelfID = id
	f.lastUserID = userID
	return Shelf{ID: id, UserID: userID, Name: name}, nil
}

func (f *fakeShelfRepository) Delete(_ context.Context, id, userID string) error {
	f.lastShelfID = id
	f.lastUserID = userID
	return nil
}

func (f *fakeShelfRepository) AddBook(_ context.Context, shelfID, userID, copyID string) error {
	f.addCalled = true
	f.lastShelfID = shelfID
	f.lastUserID = userID
	f.lastCopyID = copyID
	return f.addErr
}

func (f *fakeShelfRepository) RemoveBook(_ context.Context, shelfID, userID, copyID string) error {
	f.removeCalled = true
	f.lastShelfID = shelfID
	f.lastUserID = userID
	f.lastCopyID = copyID
	return f.removeErr
}

func (f *fakeShelfRepository) ListBooks(_ context.Context, shelfID, userID string, _, _ int) ([]books.UserBook, int, error) {
	f.listBooksCalled = true
	f.lastShelfID = shelfID
	f.lastUserID = userID
	return f.books, len(f.books), nil
}

func TestAddBookEnforcesShelfAndCopyOwnership(t *testing.T) {
	tests := []struct {
		name    string
		shelf   *Shelf
		addErr  error
		wantErr error
		wantAdd bool
	}{
		{
			name:    "foreign or missing shelf is hidden",
			wantErr: ErrShelfNotFound,
		},
		{
			name:    "foreign or missing copy is rejected",
			shelf:   &Shelf{ID: "shelf-1", UserID: "user-1"},
			addErr:  ErrCopyNotOwned,
			wantErr: ErrCopyNotOwned,
			wantAdd: true,
		},
		{
			name:    "owned copy can be added",
			shelf:   &Shelf{ID: "shelf-1", UserID: "user-1"},
			wantAdd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeShelfRepository{shelf: tt.shelf, addErr: tt.addErr}
			service := NewService(repo)

			err := service.AddBook(context.Background(), "shelf-1", "user-1", "copy-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AddBook() error = %v, want %v", err, tt.wantErr)
			}
			if repo.addCalled != tt.wantAdd {
				t.Fatalf("AddBook() repository call = %v, want %v", repo.addCalled, tt.wantAdd)
			}
			if repo.lastUserID != "user-1" || repo.lastShelfID != "shelf-1" {
				t.Fatalf("AddBook() scope = user %q shelf %q", repo.lastUserID, repo.lastShelfID)
			}
			if repo.addCalled && repo.lastCopyID != "copy-1" {
				t.Fatalf("AddBook() copy = %q, want copy-1", repo.lastCopyID)
			}
		})
	}
}

func TestListBooksAlwaysScopesQueryToShelfOwner(t *testing.T) {
	tests := []struct {
		name     string
		shelf    *Shelf
		wantErr  error
		wantList bool
	}{
		{
			name:     "owner list is scoped by user",
			shelf:    &Shelf{ID: "shelf-1", UserID: "user-1"},
			wantList: true,
		},
		{
			name:    "foreign shelf cannot be listed",
			wantErr: ErrShelfNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeShelfRepository{shelf: tt.shelf}
			service := NewService(repo)

			_, _, err := service.ListBooks(context.Background(), "shelf-1", "user-1", 1, 20)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListBooks() error = %v, want %v", err, tt.wantErr)
			}
			if repo.listBooksCalled != tt.wantList {
				t.Fatalf("ListBooks() repository call = %v, want %v", repo.listBooksCalled, tt.wantList)
			}
			if repo.lastUserID != "user-1" || repo.lastShelfID != "shelf-1" {
				t.Fatalf("ListBooks() scope = user %q shelf %q", repo.lastUserID, repo.lastShelfID)
			}
		})
	}
}

func TestRemoveBookAlwaysScopesMutationToShelfOwner(t *testing.T) {
	tests := []struct {
		name       string
		shelf      *Shelf
		removeErr  error
		wantErr    error
		wantRemove bool
	}{
		{
			name:       "owner can remove shelf entry",
			shelf:      &Shelf{ID: "shelf-1", UserID: "user-1"},
			wantRemove: true,
		},
		{
			name:       "missing book remains not found",
			shelf:      &Shelf{ID: "shelf-1", UserID: "user-1"},
			removeErr:  ErrBookNotOnShelf,
			wantErr:    ErrBookNotOnShelf,
			wantRemove: true,
		},
		{
			name:    "foreign shelf cannot be mutated",
			wantErr: ErrShelfNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeShelfRepository{shelf: tt.shelf, removeErr: tt.removeErr}
			service := NewService(repo)

			err := service.RemoveBook(context.Background(), "shelf-1", "user-1", "copy-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RemoveBook() error = %v, want %v", err, tt.wantErr)
			}
			if repo.removeCalled != tt.wantRemove {
				t.Fatalf("RemoveBook() repository call = %v, want %v", repo.removeCalled, tt.wantRemove)
			}
			if repo.lastUserID != "user-1" || repo.lastShelfID != "shelf-1" {
				t.Fatalf("RemoveBook() scope = user %q shelf %q", repo.lastUserID, repo.lastShelfID)
			}
		})
	}
}

func TestShelfNameValidation(t *testing.T) {
	service := NewService(&fakeShelfRepository{})

	if _, err := service.Create(context.Background(), "user-1", "  "); !errors.Is(err, ErrNameRequired) {
		t.Fatalf("Create() error = %v, want %v", err, ErrNameRequired)
	}
	if _, err := service.Rename(context.Background(), "shelf-1", "user-1", "\t"); !errors.Is(err, ErrNameEmpty) {
		t.Fatalf("Rename() error = %v, want %v", err, ErrNameEmpty)
	}
}
