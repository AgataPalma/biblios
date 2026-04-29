package genres

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

func autoApprove(role apictx.Role) bool {
	return role == apictx.RoleModerator || role == apictx.RoleAdmin
}

func (s *Service) ListGenres(ctx context.Context) ([]Genre, error) {
	return s.repo.ListGenres(ctx)
}

func (s *Service) CreateGenre(ctx context.Context, name string, role apictx.Role) (Genre, error) {
	if strings.TrimSpace(name) == "" {
		return Genre{}, fmt.Errorf("name is required")
	}
	return s.repo.CreateGenre(ctx, name, autoApprove(role))
}

func (s *Service) ListMoods(ctx context.Context) ([]Mood, error) {
	return s.repo.ListMoods(ctx)
}

func (s *Service) CreateMood(ctx context.Context, name string, role apictx.Role) (Mood, error) {
	if strings.TrimSpace(name) == "" {
		return Mood{}, fmt.Errorf("name is required")
	}
	return s.repo.CreateMood(ctx, name, autoApprove(role))
}
