package config

import (
	"os"
	"strings"

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
}

func Load() (Config, error) {
	_ = godotenv.Load()

	return Config{
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
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
