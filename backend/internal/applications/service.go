package applications

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

func (s *Service) List(ctx context.Context) ([]Application, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (Application, error) {
	if strings.TrimSpace(id) == "" {
		return Application{}, errors.New("application id is required")
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req CreateApplicationRequest) (Application, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)
	req.AppType = strings.TrimSpace(req.AppType)

	if req.Name == "" {
		return Application{}, errors.New("application name is required")
	}

	if req.Slug == "" {
		return Application{}, errors.New("application slug is required")
	}

	if req.AppType == "" {
		return Application{}, errors.New("application type is required")
	}

	if len(req.Roles) == 0 {
		req.Roles = []string{"access", "admin", "viewer"}
	}

	return s.repo.Create(ctx, req)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("application id is required")
	}

	if strings.TrimSpace(status) == "" {
		return errors.New("status is required")
	}

	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("application id is required")
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) MarkProvisioned(
	ctx context.Context,
	id string,
	clientUUID string,
	clientID string,
	clientSecret string,
) error {
	return s.repo.MarkProvisioned(ctx, id, clientUUID, clientID, clientSecret)
}
