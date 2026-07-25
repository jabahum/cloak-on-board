package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jabahum/keycloak-onboarder/backend/internal/credentials"
	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	AuthMode string

	AllowedOrigins []string
	APIAuthToken   string

	DatabaseURL string

	KeycloakPublicURL   string
	KeycloakInternalURL string

	KeycloakBaseURL           string
	KeycloakRealm             string
	KeycloakAdminClientID     string
	KeycloakAdminClientSecret string
	KeycloakAudience          string

	CredentialEncryptionKeys  string
	SecretDeliveryTTLMinutes  int
	DriftCheckIntervalMinutes int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppName:  getEnv("APP_NAME", "keycloak-onboarder"),
		AppEnv:   getEnv("APP_ENV", "development"),
		AppPort:  getEnv("APP_PORT", "9000"),
		AuthMode: getEnv("AUTH_MODE", "api_key"),

		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		APIAuthToken:   getEnv("API_AUTH_TOKEN", ""),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		KeycloakPublicURL:   getEnv("KEYCLOAK_PUBLIC_URL", ""),
		KeycloakInternalURL: getEnv("KEYCLOAK_INTERNAL_URL", ""),

		KeycloakBaseURL:           getEnv("KEYCLOAK_BASE_URL", ""),
		KeycloakRealm:             getEnv("KEYCLOAK_REALM", ""),
		KeycloakAdminClientID:     getEnv("KEYCLOAK_ADMIN_CLIENT_ID", ""),
		KeycloakAdminClientSecret: getEnv("KEYCLOAK_ADMIN_CLIENT_SECRET", ""),
		KeycloakAudience:          getEnv("KEYCLOAK_AUDIENCE", "keycloak-onboarder-ui"),
		CredentialEncryptionKeys:  getEnv("CREDENTIAL_ENCRYPTION_KEYS", ""),
		SecretDeliveryTTLMinutes:  getEnvInt("SECRET_DELIVERY_TTL_MINUTES", 10),
		DriftCheckIntervalMinutes: getEnvIntAllowZero("DRIFT_CHECK_INTERVAL_MINUTES", 0),
	}
	if cfg.AuthMode == "keycloak" &&
		(cfg.KeycloakPublicURL == "" || cfg.KeycloakInternalURL == "" || cfg.KeycloakRealm == "" || cfg.KeycloakAudience == "") {
		return Config{}, errors.New("keycloak authentication requires public URL, internal URL, realm, and audience")
	}
	if cfg.AppEnv == "production" && cfg.AuthMode != "keycloak" {
		return Config{}, errors.New("production requires AUTH_MODE=keycloak")
	}
	if cfg.AppEnv == "production" && cfg.CredentialEncryptionKeys == "" {
		return Config{}, errors.New("production requires CREDENTIAL_ENCRYPTION_KEYS")
	}
	if cfg.CredentialEncryptionKeys != "" {
		if _, err := credentials.Parse(cfg.CredentialEncryptionKeys); err != nil {
			return Config{}, fmt.Errorf("invalid CREDENTIAL_ENCRYPTION_KEYS: %w", err)
		}
	}
	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func getEnvIntAllowZero(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}
