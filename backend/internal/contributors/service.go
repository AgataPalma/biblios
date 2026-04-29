package contributors

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

type CreateContributorInput struct {
	Name        string  `json:"name"`
	Bio         *string `json:"bio"`
	PhotoURL    *string `json:"photo_url"`
	Website     *string `json:"website"`
	Nationality *string `json:"nationality"`
}

func (s *Service) Create(ctx context.Context, input CreateContributorInput, role apictx.Role) (Contributor, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Contributor{}, fmt.Errorf("name is required")
	}
	autoApprove := role == apictx.RoleModerator || role == apictx.RoleAdmin
	return s.repo.Create(ctx, input.Name, input.Bio, input.PhotoURL, input.Website, input.Nationality, autoApprove)
}

func (s *Service) GetDetail(ctx context.Context, id string) (*ContributorDetail, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	awards, err := s.repo.GetAwards(ctx, id)
	if err != nil {
		return nil, err
	}
	if awards == nil {
		awards = []ContributorAward{}
	}
	return &ContributorDetail{Contributor: *c, Awards: awards}, nil
}

func (s *Service) Search(ctx context.Context, query string, page, limit int) ([]Contributor, int, error) {
	return s.repo.Search(ctx, query, page, limit)
}
