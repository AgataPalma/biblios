package library

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/AgataPalma/biblios/internal/books"
)

var (
	ErrLibraryNotFound           = errors.New("library not found")
	ErrAccessDenied              = errors.New("access denied")
	ErrNotMember                 = errors.New("not a member of this library")
	ErrCopyNotOwned              = errors.New("copy not found")
	ErrBookNotInLibrary          = errors.New("book copy not found in library")
	ErrInvalidInvitationEmail    = errors.New("invalid email address")
	ErrInvitationNotFound        = errors.New("invitation not found")
	ErrInvitationNotForUser      = errors.New("invitation is not intended for this user")
	ErrInvitationNotPending      = errors.New("invitation is no longer pending")
	ErrInvitationExpired         = errors.New("invitation has expired")
	ErrInvitationAlreadyPending  = errors.New("an invitation is already pending for this email")
	ErrInvitationRecipientMember = errors.New("user is already a member of this library")
	ErrInvitationCreateForbidden = errors.New("not allowed to invite members to this library")
)

type libraryRepository interface {
	CreateLibrary(ctx context.Context, ownerID, name string, description *string, isCooperative bool, visibility string) (Library, error)
	FindLibraryByID(ctx context.Context, id string) (*Library, error)
	ListUserLibraries(ctx context.Context, userID string) ([]Library, error)
	ListPublicLibraries(ctx context.Context, page, limit int) ([]Library, int, error)
	UpdateLibrary(ctx context.Context, id string, name, description, visibility *string) (Library, error)
	DeleteLibrary(ctx context.Context, id string) error
	GetMember(ctx context.Context, libraryID, userID string) (*LibraryMember, error)
	ListMembers(ctx context.Context, libraryID string) ([]LibraryMember, error)
	UpdateMemberPermissions(ctx context.Context, libraryID, userID string, perms LibraryMember) error
	RemoveMember(ctx context.Context, libraryID, userID string) error
	CreateInvitation(ctx context.Context, libraryID, invitedBy, invitedEmail string, expiresAt time.Time) (LibraryInvitation, error)
	AcceptInvitation(ctx context.Context, token, userID string) error
	DeclineInvitation(ctx context.Context, token, userID string) error
	ListUserInvitations(ctx context.Context, userID string) ([]LibraryInvitation, error)
	AddBookCopyToLibrary(ctx context.Context, libraryID, userID, copyID string) error
	RemoveBookCopyFromLibrary(ctx context.Context, libraryID, userID, copyID string) error
	ListLibraryBooks(ctx context.Context, libraryID string, page, limit int) ([]LibraryBook, int, error)
	ListUserLibraryBooks(ctx context.Context, libraryID, userID string, page, limit int) ([]books.UserBook, int, error)
}

type Service struct {
	repo libraryRepository
}

func NewService(repo libraryRepository) *Service {
	return &Service{repo: repo}
}

// ─── Input types ──────────────────────────────────────────────────────────────

type CreateLibraryInput struct {
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	IsCooperative bool    `json:"is_cooperative"`
	Visibility    string  `json:"visibility"` // private | semi_public | public
}

type UpdateLibraryInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Visibility  *string `json:"visibility"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func validVisibility(v string) bool {
	return v == "private" || v == "semi_public" || v == "public"
}

// getMemberOrFail returns the caller's membership record, or an error if not a member.
func (s *Service) getMemberOrFail(ctx context.Context, libraryID, userID string) (*LibraryMember, error) {
	m, err := s.repo.GetMember(ctx, libraryID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if m == nil {
		return nil, ErrNotMember
	}
	return m, nil
}

// ─── Library CRUD ─────────────────────────────────────────────────────────────

func (s *Service) CreateLibrary(ctx context.Context, userID string, input CreateLibraryInput) (Library, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Library{}, fmt.Errorf("name is required")
	}
	if input.Visibility == "" {
		input.Visibility = "private"
	}
	if !validVisibility(input.Visibility) {
		return Library{}, fmt.Errorf("invalid visibility: must be private, semi_public, or public")
	}
	return s.repo.CreateLibrary(ctx, userID, input.Name, input.Description, input.IsCooperative, input.Visibility)
}

func publicLibraryView(library Library) Library {
	library.OwnerID = ""
	return library
}

// GetLibrary returns the full library to members with can_view permission.
// Public libraries are visible to other callers with internal owner data
// redacted. Semi-public and private libraries require can_view membership.
func (s *Service) GetLibrary(ctx context.Context, libraryID, userID string) (*Library, error) {
	lib, err := s.repo.FindLibraryByID(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	if lib == nil {
		return nil, ErrLibraryNotFound
	}

	m, err := s.repo.GetMember(ctx, libraryID, userID)
	if err != nil {
		return nil, err
	}
	if m != nil && m.CanView {
		return lib, nil
	}
	if lib.Visibility == "public" {
		publicView := publicLibraryView(*lib)
		return &publicView, nil
	}
	return nil, ErrAccessDenied
}

func (s *Service) ListMyLibraries(ctx context.Context, userID string) ([]Library, error) {
	return s.repo.ListUserLibraries(ctx, userID)
}

func (s *Service) ListPublicLibraries(ctx context.Context, page, limit int) ([]Library, int, error) {
	libraries, total, err := s.repo.ListPublicLibraries(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range libraries {
		libraries[i] = publicLibraryView(libraries[i])
	}
	return libraries, total, nil
}

// UpdateLibrary is restricted to the library owner.
func (s *Service) UpdateLibrary(ctx context.Context, libraryID, userID string, input UpdateLibraryInput) (Library, error) {
	m, err := s.getMemberOrFail(ctx, libraryID, userID)
	if err != nil {
		return Library{}, err
	}
	if !m.IsOwner {
		return Library{}, fmt.Errorf("only the library owner can update library settings")
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Library{}, fmt.Errorf("name cannot be empty")
	}
	if input.Visibility != nil && !validVisibility(*input.Visibility) {
		return Library{}, fmt.Errorf("invalid visibility")
	}
	return s.repo.UpdateLibrary(ctx, libraryID, input.Name, input.Description, input.Visibility)
}

func (s *Service) DeleteLibrary(ctx context.Context, libraryID, userID string) error {
	m, err := s.getMemberOrFail(ctx, libraryID, userID)
	if err != nil {
		return err
	}
	if !m.IsOwner {
		return ErrAccessDenied
	}
	return s.repo.DeleteLibrary(ctx, libraryID)
}

// ─── Invitations ──────────────────────────────────────────────────────────────

// InviteMember sends an invitation to an email address. Requires can_invite permission.
func (s *Service) InviteMember(ctx context.Context, libraryID, inviterID, email string) (LibraryInvitation, error) {
	m, err := s.getMemberOrFail(ctx, libraryID, inviterID)
	if err != nil {
		return LibraryInvitation{}, err
	}
	if !m.CanInvite {
		return LibraryInvitation{}, fmt.Errorf("you do not have permission to invite members")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(address.Address, email) {
		return LibraryInvitation{}, ErrInvalidInvitationEmail
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	return s.repo.CreateInvitation(ctx, libraryID, inviterID, email, expiresAt)
}

func (s *Service) AcceptInvitation(ctx context.Context, token, userID string) error {
	return s.repo.AcceptInvitation(ctx, token, userID)
}

func (s *Service) DeclineInvitation(ctx context.Context, token, userID string) error {
	return s.repo.DeclineInvitation(ctx, token, userID)
}

func (s *Service) ListMyInvitations(ctx context.Context, userID string) ([]LibraryInvitation, error) {
	return s.repo.ListUserInvitations(ctx, userID)
}

// ─── Member management ────────────────────────────────────────────────────────

// UpdateMemberPermissions requires can_manage_members. Cannot change owner permissions.
func (s *Service) UpdateMemberPermissions(ctx context.Context, libraryID, requesterID, targetUserID string, perms LibraryMember) error {
	requester, err := s.getMemberOrFail(ctx, libraryID, requesterID)
	if err != nil {
		return err
	}
	if !requester.CanManageMembers {
		return fmt.Errorf("you do not have permission to manage members")
	}

	target, err := s.repo.GetMember(ctx, libraryID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("target user is not a member of this library")
	}
	if target.IsOwner {
		return fmt.Errorf("cannot change owner permissions")
	}

	return s.repo.UpdateMemberPermissions(ctx, libraryID, targetUserID, perms)
}

// RemoveMember is restricted to the library owner. Cannot remove the owner.
func (s *Service) RemoveMember(ctx context.Context, libraryID, requesterID, targetUserID string) error {
	requester, err := s.getMemberOrFail(ctx, libraryID, requesterID)
	if err != nil {
		return err
	}
	if !requester.IsOwner {
		return fmt.Errorf("only the library owner can remove members")
	}
	if requesterID == targetUserID {
		return fmt.Errorf("owner cannot remove themselves")
	}

	target, err := s.repo.GetMember(ctx, libraryID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("user is not a member of this library")
	}
	if target.IsOwner {
		return fmt.Errorf("cannot remove the library owner")
	}

	return s.repo.RemoveMember(ctx, libraryID, targetUserID)
}

func (s *Service) ListMembers(ctx context.Context, libraryID, requesterID string) ([]LibraryMember, error) {
	member, err := s.getMemberOrFail(ctx, libraryID, requesterID)
	if err != nil {
		return nil, err
	}
	if !member.CanView {
		return nil, ErrAccessDenied
	}
	return s.repo.ListMembers(ctx, libraryID)
}

// ─── Library books ────────────────────────────────────────────────────────────

// AddBookToLibrary requires can_add permission.
func (s *Service) AddBookToLibrary(ctx context.Context, libraryID, userID, copyID string) error {
	m, err := s.getMemberOrFail(ctx, libraryID, userID)
	if err != nil {
		return err
	}
	if !m.CanAdd {
		return fmt.Errorf("you do not have permission to add books to this library")
	}
	return s.repo.AddBookCopyToLibrary(ctx, libraryID, userID, copyID)
}

// RemoveBookFromLibrary requires can_remove permission.
func (s *Service) RemoveBookFromLibrary(ctx context.Context, libraryID, userID, copyID string) error {
	m, err := s.getMemberOrFail(ctx, libraryID, userID)
	if err != nil {
		return err
	}
	if !m.CanRemove {
		return fmt.Errorf("you do not have permission to remove books from this library")
	}
	return s.repo.RemoveBookCopyFromLibrary(ctx, libraryID, userID, copyID)
}

// ListLibraryBooks returns catalogue-only data for shared/public library
// browsing. Physical copy IDs are included only for members with can_view.
func (s *Service) ListLibraryBooks(ctx context.Context, libraryID, userID string, page, limit int) ([]LibraryBook, int, error) {
	lib, err := s.GetLibrary(ctx, libraryID, userID)
	if err != nil {
		return nil, 0, err
	}
	books, total, err := s.repo.ListLibraryBooks(ctx, libraryID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// GetLibrary clears OwnerID only for the redacted public view. Members with
	// can_view receive the full library, so no second membership query is needed.
	if lib.OwnerID == "" {
		for i := range books {
			books[i].CopyID = ""
		}
	}
	return books, total, nil
}

// ListMyLibraryBooks returns private copy state only for copies owned by the
// caller. It is used by /users/me/library, never by shared/public endpoints.
func (s *Service) ListMyLibraryBooks(ctx context.Context, libraryID, userID string, page, limit int) ([]books.UserBook, int, error) {
	member, err := s.getMemberOrFail(ctx, libraryID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !member.CanView {
		return nil, 0, ErrAccessDenied
	}
	return s.repo.ListUserLibraryBooks(ctx, libraryID, userID, page, limit)
}
