package series

import (
	"context"
	"fmt"
	"strings"

	"github.com/AgataPalma/biblios/internal/apictx"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateSeriesInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (s *Service) Create(ctx context.Context, input CreateSeriesInput, role apictx.Role) (Series, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Series{}, fmt.Errorf("name is required")
	}
	autoApprove := role == apictx.RoleModerator || role == apictx.RoleAdmin
	return s.repo.Create(ctx, input.Name, input.Description, autoApprove)
}

func (s *Service) Search(ctx context.Context, query string, page, limit int) ([]Series, int, error) {
	return s.repo.Search(ctx, query, page, limit)
}

func (s *Service) GetDetail(ctx context.Context, id string) (*SeriesDetail, error) {
	ser, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ser == nil {
		return nil, nil
	}
	bks, err := s.repo.GetBooks(ctx, id)
	if err != nil {
		return nil, err
	}
	if bks == nil {
		bks = []SeriesBook{}
	}
	return &SeriesDetail{Series: *ser, Books: bks}, nil
}
