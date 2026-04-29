package notifications

import (
	"context"
	"encoding/json"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create is called internally by other services to emit a notification.
func (s *Service) Create(ctx context.Context, userID, notifType, title, body string, data any) error {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err == nil {
			raw = b
		}
	}
	_, err := s.repo.Create(ctx, userID, notifType, title, body, raw)
	return err
}

// NotifyLibraryInvitation creates a library invitation notification.
func (s *Service) NotifyLibraryInvitation(ctx context.Context, userID, libraryName, inviterUsername string) error {
	return s.Create(ctx, userID, TypeLibraryInvitation,
		"Library Invitation",
		inviterUsername+" invited you to join "+libraryName,
		map[string]string{"library_name": libraryName, "inviter": inviterUsername},
	)
}

// NotifyReviewLike creates a review like notification.
func (s *Service) NotifyReviewLike(ctx context.Context, userID, likerUsername, bookTitle string) error {
	return s.Create(ctx, userID, TypeReviewLike,
		"Review Liked",
		likerUsername+" liked your review of "+bookTitle,
		map[string]string{"liker": likerUsername, "book_title": bookTitle},
	)
}

// NotifyLibraryActivity creates a cooperative library activity notification.
func (s *Service) NotifyLibraryActivity(ctx context.Context, userID, message string, data any) error {
	return s.Create(ctx, userID, TypeLibraryActivity, "Library Activity", message, data)
}

// List returns paginated notifications for a user with optional filters.
func (s *Service) List(ctx context.Context, userID, filterType, filterRead string, page, limit int) ([]Notification, int, error) {
	return s.repo.List(ctx, userID, filterType, filterRead, page, limit)
}

// MarkRead marks a single notification as read.
func (s *Service) MarkRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkRead(ctx, id, userID)
}

// MarkAllRead marks all unread notifications for a user as read.
func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}
