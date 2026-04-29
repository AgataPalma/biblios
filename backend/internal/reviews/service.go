package reviews

import (
	"context"
	"fmt"
	"math"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type UpsertReviewInput struct {
	BookID   string
	UserID   string
	Rating   float64
	Body     *string
	IsPublic bool
}

func validateRating(r float64) error {
	if r < 0.0 || r > 5.0 {
		return fmt.Errorf("rating must be between 0.0 and 5.0")
	}
	// Must be a multiple of 0.1 — multiply by 10 and check for integer
	if math.Round(r*10) != r*10 {
		return fmt.Errorf("rating must be in 0.1 increments")
	}
	return nil
}

func (s *Service) UpsertReview(ctx context.Context, input UpsertReviewInput) (Review, error) {
	if err := validateRating(input.Rating); err != nil {
		return Review{}, err
	}
	if input.Body != nil && len(*input.Body) > 5000 {
		return Review{}, fmt.Errorf("review body must be 5000 characters or fewer")
	}
	return s.repo.Upsert(ctx, input.BookID, input.UserID, input.Rating, input.Body, input.IsPublic)
}

func (s *Service) GetMyReview(ctx context.Context, bookID, userID string) (*Review, error) {
	return s.repo.FindByBookAndUser(ctx, bookID, userID)
}

func (s *Service) ListPublicReviews(ctx context.Context, bookID, callerID string, page, limit int) (ReviewsResponse, error) {
	revs, total, avg, err := s.repo.ListPublic(ctx, bookID, callerID, page, limit)
	if err != nil {
		return ReviewsResponse{}, err
	}
	if revs == nil {
		revs = []Review{}
	}
	return ReviewsResponse{
		Reviews:       revs,
		Total:         total,
		Page:          page,
		Limit:         limit,
		AverageRating: avg,
	}, nil
}

func (s *Service) DeleteMyReview(ctx context.Context, bookID, userID string) error {
	return s.repo.SoftDelete(ctx, bookID, userID)
}

func (s *Service) LikeReview(ctx context.Context, reviewID, userID string) error {
	return s.repo.AddLike(ctx, reviewID, userID)
}

func (s *Service) UnlikeReview(ctx context.Context, reviewID, userID string) error {
	return s.repo.RemoveLike(ctx, reviewID, userID)
}
