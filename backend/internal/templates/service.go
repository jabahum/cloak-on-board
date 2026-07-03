package templates

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Template, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (Template, error) {
	if strings.TrimSpace(id) == "" {
		return Template{}, errors.New("template id is required")
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) SeedDefaults(ctx context.Context) error {
	return s.repo.SeedDefaults(ctx)
}
