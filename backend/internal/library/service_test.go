package library

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AgataPalma/biblios/internal/books"
)

type fakeLibraryRepository struct {
	library             *Library
	member              *LibraryMember
	members             []LibraryMember
	libraries           []Library
	libraryBooks        []LibraryBook
	userBooks           []books.UserBook
	addErr              error
	removeErr           error
	deleteCalled        bool
	listMembersCalled   bool
	listBooksCalled     bool
	listUserBooksCalled bool
	addCalled           bool
	removeCalled        bool
	lastLibraryID       string
	lastUserID          string
	lastCopyID          string
}

func (f *fakeLibraryRepository) CreateLibrary(_ context.Context, ownerID, name string, _ *string, cooperative bool, visibility string) (Library, error) {
	return Library{OwnerID: ownerID, Name: name, IsCooperative: cooperative, Visibility: visibility}, nil
}

func (f *fakeLibraryRepository) FindLibraryByID(_ context.Context, id string) (*Library, error) {
	f.lastLibraryID = id
	return f.library, nil
}

func (f *fakeLibraryRepository) ListUserLibraries(_ context.Context, userID string) ([]Library, error) {
	f.lastUserID = userID
	return f.libraries, nil
}

func (f *fakeLibraryRepository) ListPublicLibraries(_ context.Context, _, _ int) ([]Library, int, error) {
	return f.libraries, len(f.libraries), nil
}

func (f *fakeLibraryRepository) UpdateLibrary(_ context.Context, id string, _, _, _ *string) (Library, error) {
	f.lastLibraryID = id
	return Library{ID: id}, nil
}

func (f *fakeLibraryRepository) DeleteLibrary(_ context.Context, id string) error {
	f.deleteCalled = true
	f.lastLibraryID = id
	return nil
}

func (f *fakeLibraryRepository) GetMember(_ context.Context, libraryID, userID string) (*LibraryMember, error) {
	f.lastLibraryID = libraryID
	f.lastUserID = userID
	return f.member, nil
}

func (f *fakeLibraryRepository) ListMembers(_ context.Context, libraryID string) ([]LibraryMember, error) {
	f.listMembersCalled = true
	f.lastLibraryID = libraryID
	return f.members, nil
}

func (f *fakeLibraryRepository) UpdateMemberPermissions(_ context.Context, libraryID, _ string, _ LibraryMember) error {
	f.lastLibraryID = libraryID
	return nil
}

func (f *fakeLibraryRepository) RemoveMember(_ context.Context, libraryID, _ string) error {
	f.lastLibraryID = libraryID
	return nil
}

func (f *fakeLibraryRepository) CreateInvitation(_ context.Context, libraryID, invitedBy, invitedEmail string, expiresAt time.Time) (LibraryInvitation, error) {
	return LibraryInvitation{LibraryID: libraryID, InvitedBy: invitedBy, InvitedEmail: invitedEmail, ExpiresAt: expiresAt}, nil
}

func (f *fakeLibraryRepository) AcceptInvitation(_ context.Context, _, _ string) error {
	return nil
}

func (f *fakeLibraryRepository) DeclineInvitation(_ context.Context, _, _ string) error {
	return nil
}

func (f *fakeLibraryRepository) ListUserInvitations(_ context.Context, _ string) ([]LibraryInvitation, error) {
	return nil, nil
}

func (f *fakeLibraryRepository) AddBookCopyToLibrary(_ context.Context, libraryID, userID, copyID string) error {
	f.addCalled = true
	f.lastLibraryID = libraryID
	f.lastUserID = userID
	f.lastCopyID = copyID
	return f.addErr
}

func (f *fakeLibraryRepository) RemoveBookCopyFromLibrary(_ context.Context, libraryID, userID, copyID string) error {
	f.removeCalled = true
	f.lastLibraryID = libraryID
	f.lastUserID = userID
	f.lastCopyID = copyID
	return f.removeErr
}

func (f *fakeLibraryRepository) ListLibraryBooks(_ context.Context, libraryID string, _, _ int) ([]LibraryBook, int, error) {
	f.listBooksCalled = true
	f.lastLibraryID = libraryID
	return f.libraryBooks, len(f.libraryBooks), nil
}

func (f *fakeLibraryRepository) ListUserLibraryBooks(_ context.Context, libraryID, userID string, _, _ int) ([]books.UserBook, int, error) {
	f.listUserBooksCalled = true
	f.lastLibraryID = libraryID
	f.lastUserID = userID
	return f.userBooks, len(f.userBooks), nil
}

func TestGetLibraryEnforcesCanViewAndRedactsPublicView(t *testing.T) {
	tests := []struct {
		name        string
		visibility  string
		member      *LibraryMember
		wantErr     error
		wantOwnerID string
	}{
		{
			name:       "private visitor is denied",
			visibility: "private",
			wantErr:    ErrAccessDenied,
		},
		{
			name:       "private member without can view is denied",
			visibility: "private",
			member:     &LibraryMember{IsOwner: false, CanView: false},
			wantErr:    ErrAccessDenied,
		},
		{
			name:        "viewing member receives full library",
			visibility:  "private",
			member:      &LibraryMember{CanView: true},
			wantOwnerID: "owner-1",
		},
		{
			name:       "public visitor receives redacted library",
			visibility: "public",
		},
		{
			name:       "public member without can view receives redacted library",
			visibility: "public",
			member:     &LibraryMember{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLibraryRepository{
				library: &Library{ID: "library-1", OwnerID: "owner-1", Visibility: tt.visibility},
				member:  tt.member,
			}
			service := NewService(repo)

			library, err := service.GetLibrary(context.Background(), "library-1", "user-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetLibrary() error = %v, want %v", err, tt.wantErr)
			}
			if err == nil && library.OwnerID != tt.wantOwnerID {
				t.Fatalf("GetLibrary() owner ID = %q, want %q", library.OwnerID, tt.wantOwnerID)
			}
		})
	}
}

func TestListPublicLibrariesRedactsOwnerIDs(t *testing.T) {
	repo := &fakeLibraryRepository{libraries: []Library{{ID: "library-1", OwnerID: "owner-1", Visibility: "public"}}}
	service := NewService(repo)

	libraries, _, err := service.ListPublicLibraries(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("ListPublicLibraries() error = %v", err)
	}
	if libraries[0].OwnerID != "" {
		t.Fatalf("ListPublicLibraries() exposed owner ID %q", libraries[0].OwnerID)
	}
}

func TestListMembersRequiresCanView(t *testing.T) {
	tests := []struct {
		name     string
		member   *LibraryMember
		wantErr  error
		wantList bool
	}{
		{name: "viewer can list members", member: &LibraryMember{CanView: true}, wantList: true},
		{name: "member without view permission is denied", member: &LibraryMember{}, wantErr: ErrAccessDenied},
		{name: "non-member is denied", wantErr: ErrNotMember},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLibraryRepository{member: tt.member}
			service := NewService(repo)

			_, err := service.ListMembers(context.Background(), "library-1", "user-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListMembers() error = %v, want %v", err, tt.wantErr)
			}
			if repo.listMembersCalled != tt.wantList {
				t.Fatalf("ListMembers() repository call = %v, want %v", repo.listMembersCalled, tt.wantList)
			}
		})
	}
}

func TestAddBookPassesCallerToAtomicOwnershipCheck(t *testing.T) {
	tests := []struct {
		name    string
		member  *LibraryMember
		addErr  error
		wantErr error
		wantAdd bool
	}{
		{name: "member with add permission can add owned copy", member: &LibraryMember{CanAdd: true}, wantAdd: true},
		{name: "foreign copy is rejected by repository", member: &LibraryMember{CanAdd: true}, addErr: ErrCopyNotOwned, wantErr: ErrCopyNotOwned, wantAdd: true},
		{name: "member without add permission is denied", member: &LibraryMember{}, wantErr: errors.New("you do not have permission to add books to this library")},
		{name: "non-member is denied", wantErr: ErrNotMember},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLibraryRepository{member: tt.member, addErr: tt.addErr}
			service := NewService(repo)

			err := service.AddBookToLibrary(context.Background(), "library-1", "user-1", "copy-1")
			if tt.wantErr == nil && err != nil || tt.wantErr != nil && (err == nil || err.Error() != tt.wantErr.Error()) {
				t.Fatalf("AddBookToLibrary() error = %v, want %v", err, tt.wantErr)
			}
			if repo.addCalled != tt.wantAdd {
				t.Fatalf("AddBookToLibrary() repository call = %v, want %v", repo.addCalled, tt.wantAdd)
			}
			if repo.addCalled && (repo.lastUserID != "user-1" || repo.lastCopyID != "copy-1") {
				t.Fatalf("AddBookToLibrary() scope = user %q copy %q", repo.lastUserID, repo.lastCopyID)
			}
		})
	}
}

func TestSharedLibraryBooksNeverExposePersonalCopyState(t *testing.T) {
	repo := &fakeLibraryRepository{
		library:      &Library{ID: "library-1", OwnerID: "owner-1", Visibility: "public"},
		libraryBooks: []LibraryBook{{LibraryID: "library-1", CopyID: "private-copy-id", BookID: "book-1"}},
	}
	service := NewService(repo)

	publicBooks, _, err := service.ListLibraryBooks(context.Background(), "library-1", "visitor", 1, 20)
	if err != nil {
		t.Fatalf("ListLibraryBooks() error = %v", err)
	}
	if publicBooks[0].CopyID != "" {
		t.Fatalf("ListLibraryBooks() exposed public copy ID %q", publicBooks[0].CopyID)
	}

	payload, err := json.Marshal(publicBooks[0])
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	for _, privateField := range []string{
		"reading_status", "current_page", "started_reading_at", "finished_reading_at",
		"borrowed_from", "location", "condition", "personal_notes", "reread_count",
	} {
		if strings.Contains(string(payload), privateField) {
			t.Fatalf("library payload leaked private field %q: %s", privateField, payload)
		}
	}
}

func TestMemberLibraryBooksKeepCopyIDAndPrivateEndpointScopesOwner(t *testing.T) {
	repo := &fakeLibraryRepository{
		library:      &Library{ID: "library-1", OwnerID: "owner-1", Visibility: "private"},
		member:       &LibraryMember{CanView: true},
		libraryBooks: []LibraryBook{{LibraryID: "library-1", CopyID: "copy-1", BookID: "book-1"}},
	}
	service := NewService(repo)

	sharedBooks, _, err := service.ListLibraryBooks(context.Background(), "library-1", "user-1", 1, 20)
	if err != nil {
		t.Fatalf("ListLibraryBooks() error = %v", err)
	}
	if sharedBooks[0].CopyID != "copy-1" {
		t.Fatalf("ListLibraryBooks() copy ID = %q, want copy-1", sharedBooks[0].CopyID)
	}

	if _, _, err := service.ListMyLibraryBooks(context.Background(), "library-1", "user-1", 1, 20); err != nil {
		t.Fatalf("ListMyLibraryBooks() error = %v", err)
	}
	if !repo.listUserBooksCalled || repo.lastUserID != "user-1" {
		t.Fatalf("ListMyLibraryBooks() was not scoped to user-1")
	}
}

func TestDeleteLibraryRequiresOwner(t *testing.T) {
	tests := []struct {
		name       string
		member     *LibraryMember
		wantErr    error
		wantDelete bool
	}{
		{name: "owner can delete", member: &LibraryMember{IsOwner: true}, wantDelete: true},
		{name: "non-owner is denied", member: &LibraryMember{}, wantErr: ErrAccessDenied},
		{name: "non-member is denied", wantErr: ErrNotMember},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLibraryRepository{member: tt.member}
			service := NewService(repo)

			err := service.DeleteLibrary(context.Background(), "library-1", "user-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DeleteLibrary() error = %v, want %v", err, tt.wantErr)
			}
			if repo.deleteCalled != tt.wantDelete {
				t.Fatalf("DeleteLibrary() repository call = %v, want %v", repo.deleteCalled, tt.wantDelete)
			}
		})
	}
}
