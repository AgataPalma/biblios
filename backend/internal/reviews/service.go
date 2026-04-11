package reviews

import "context"

type ReviewsResult struct {
	Reviews []Review `json:"reviews"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetBookReviews(ctx context.Context, bookID string, page, limit int) (ReviewsResult, error) {
	list, total, err := s.repo.FindByBook(ctx, bookID, page, limit)
	if err != nil {
		return ReviewsResult{}, err
	}
	return ReviewsResult{Reviews: list, Total: total, Page: page, Limit: limit}, nil
}

func (s *Service) GetMyReview(ctx context.Context, bookID, userID string) (*Review, error) {
	return s.repo.FindByUser(ctx, bookID, userID)
}

func (s *Service) UpsertReview(ctx context.Context, input UpsertReviewInput) (Review, error) {
	return s.repo.Upsert(ctx, input)
}

func (s *Service) DeleteReview(ctx context.Context, bookID, userID string) error {
	return s.repo.Delete(ctx, bookID, userID)
}
