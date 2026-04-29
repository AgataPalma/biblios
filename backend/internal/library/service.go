package library

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AgataPalma/biblios/internal/books"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
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
		return nil, fmt.Errorf("not a member of this library")
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

// GetLibrary returns a library if the caller is allowed to view it.
// - Owners and members can always view.
// - Public libraries are visible to everyone.
// - Semi-public and private libraries require membership.
func (s *Service) GetLibrary(ctx context.Context, libraryID, userID string) (*Library, error) {
	lib, err := s.repo.FindLibraryByID(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	if lib == nil {
		return nil, fmt.Errorf("library not found")
	}

	if lib.Visibility == "public" {
		return lib, nil
	}

	// For private / semi_public: caller must be a member
	m, err := s.repo.GetMember(ctx, libraryID, userID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("access denied")
	}
	return lib, nil
}

func (s *Service) ListMyLibraries(ctx context.Context, userID string) ([]Library, error) {
	return s.repo.ListUserLibraries(ctx, userID)
}

func (s *Service) ListPublicLibraries(ctx context.Context, page, limit int) ([]Library, int, error) {
	return s.repo.ListPublicLibraries(ctx, page, limit)
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
	if strings.TrimSpace(email) == "" {
		return LibraryInvitation{}, fmt.Errorf("email is required")
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	return s.repo.CreateInvitation(ctx, libraryID, inviterID, email, expiresAt)
}

func (s *Service) AcceptInvitation(ctx context.Context, token, userID string) error {
	return s.repo.AcceptInvitation(ctx, token, userID)
}

func (s *Service) DeclineInvitation(ctx context.Context, token string) error {
	return s.repo.DeclineInvitation(ctx, token)
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
	// Must be a member to list members
	if _, err := s.getMemberOrFail(ctx, libraryID, requesterID); err != nil {
		return nil, err
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
	return s.repo.AddBookCopyToLibrary(ctx, libraryID, copyID)
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
	return s.repo.RemoveBookCopyFromLibrary(ctx, libraryID, copyID)
}

// ListLibraryBooks checks visibility before returning books.
func (s *Service) ListLibraryBooks(ctx context.Context, libraryID, userID string, page, limit int) ([]books.UserBook, int, error) {
	lib, err := s.GetLibrary(ctx, libraryID, userID)
	if err != nil {
		return nil, 0, err
	}
	if lib == nil {
		return nil, 0, fmt.Errorf("library not found")
	}
	return s.repo.ListLibraryBooks(ctx, libraryID, page, limit)
}
