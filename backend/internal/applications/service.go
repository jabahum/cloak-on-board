package applications

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
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
	req = normalizeCreate(req)
	if err := validateApplication(req.Name, req.Slug, req.AppType, req.OwnerEmail); err != nil {
		return Application{}, err
	}

	if len(req.Roles) == 0 {
		req.Roles = []string{"access", "admin", "viewer"}
	}

	return s.repo.Create(ctx, req)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateApplicationRequest) (Application, error) {
	req, err := s.ValidateUpdate(ctx, id, req)
	if err != nil {
		return Application{}, err
	}
	return s.repo.Update(ctx, id, req)
}

func (s *Service) ValidateUpdate(ctx context.Context, id string, req UpdateApplicationRequest) (UpdateApplicationRequest, error) {
	if strings.TrimSpace(id) == "" {
		return UpdateApplicationRequest{}, fmt.Errorf("%w: application id is required", ErrValidation)
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)
	req.AppType = strings.TrimSpace(req.AppType)
	req.OwnerName = strings.TrimSpace(req.OwnerName)
	req.OwnerEmail = strings.TrimSpace(req.OwnerEmail)
	req.RedirectURIs = cleanValues(req.RedirectURIs)
	req.WebOrigins = cleanValues(req.WebOrigins)
	req.Roles = cleanValues(req.Roles)
	if err := validateApplication(req.Name, req.Slug, req.AppType, req.OwnerEmail); err != nil {
		return UpdateApplicationRequest{}, err
	}
	if err := s.repo.EnsureIdentityAvailable(ctx, id, req.Slug); err != nil {
		return UpdateApplicationRequest{}, err
	}
	return req, nil
}

func (s *Service) Import(ctx context.Context, req CreateApplicationRequest, clientUUID, clientID string, enabled bool) (Application, error) {
	req = normalizeCreate(req)
	if err := validateApplication(req.Name, req.Slug, req.AppType, req.OwnerEmail); err != nil {
		return Application{}, err
	}
	return s.repo.Import(ctx, req, clientUUID, clientID, enabled)
}

func (s *Service) LinkedClientUUIDs(ctx context.Context) (map[string]bool, error) {
	return s.repo.LinkedClientUUIDs(ctx)
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

func normalizeCreate(req CreateApplicationRequest) CreateApplicationRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)
	req.AppType = strings.TrimSpace(req.AppType)
	req.OwnerName = strings.TrimSpace(req.OwnerName)
	req.OwnerEmail = strings.TrimSpace(req.OwnerEmail)
	req.RedirectURIs = cleanValues(req.RedirectURIs)
	req.WebOrigins = cleanValues(req.WebOrigins)
	req.Roles = cleanValues(req.Roles)
	return req
}

func cleanValues(values []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func validateApplication(name, slug, appType, ownerEmail string) error {
	if name == "" {
		return fmt.Errorf("%w: application name is required", ErrValidation)
	}
	if slug == "" {
		return fmt.Errorf("%w: application slug is required", ErrValidation)
	}
	switch appType {
	case "frontend", "backend", "mobile", "machine_to_machine":
	default:
		return fmt.Errorf("%w: unsupported application type %q", ErrValidation, appType)
	}
	if ownerEmail != "" {
		if _, err := mail.ParseAddress(ownerEmail); err != nil {
			return fmt.Errorf("%w: owner email is invalid", ErrValidation)
		}
	}
	return nil
}
