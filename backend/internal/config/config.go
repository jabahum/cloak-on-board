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

	AllowedOrigins []string
	APIAuthToken   string

	DatabaseURL string

	KeycloakBaseURL           string
	KeycloakRealm             string
	KeycloakAdminClientID     string
	KeycloakAdminClientSecret string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	return Config{
		AppName: getEnv("APP_NAME", "keycloak-onboarder"),
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "9000"),

		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		APIAuthToken:   getEnv("API_AUTH_TOKEN", ""),

		DatabaseURL: getEnv("DATABASE_URL", ""),

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
