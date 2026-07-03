package settings

import (
	"context"
	"errors"
	"strings"

	"github.com/jabahum/keycloak-onboarder/backend/internal/config"
)

type Service struct {
	repo *Repository
	cfg  config.Config
}

func NewService(repo *Repository, cfg config.Config) *Service {
	return &Service{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *Service) Get(ctx context.Context) (Settings, error) {
	item, err := s.repo.Get(ctx)
	if err == nil {
		return item, nil
	}

	return Settings{
		KeycloakBaseURL:       s.cfg.KeycloakBaseURL,
		KeycloakRealm:         s.cfg.KeycloakRealm,
		KeycloakAdminClientID: s.cfg.KeycloakAdminClientID,
	}, nil
}

func (s *Service) Save(ctx context.Context, req SaveSettingsRequest) (Settings, error) {
	req.KeycloakBaseURL = strings.TrimSpace(req.KeycloakBaseURL)
	req.KeycloakRealm = strings.TrimSpace(req.KeycloakRealm)
	req.KeycloakAdminClientID = strings.TrimSpace(req.KeycloakAdminClientID)

	if req.KeycloakBaseURL == "" {
		return Settings{}, errors.New("keycloak base url is required")
	}

	if req.KeycloakRealm == "" {
		return Settings{}, errors.New("keycloak realm is required")
	}

	if req.KeycloakAdminClientID == "" {
		return Settings{}, errors.New("keycloak admin client id is required")
	}

	return s.repo.Save(ctx, req)
}
