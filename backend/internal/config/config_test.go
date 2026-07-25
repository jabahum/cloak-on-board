package config

import "testing"

func setProductionEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_MODE", "keycloak")
	t.Setenv("KEYCLOAK_PUBLIC_URL", "https://auth.example.test")
	t.Setenv("KEYCLOAK_INTERNAL_URL", "http://keycloak:8080")
	t.Setenv("KEYCLOAK_REALM", "onboarder")
	t.Setenv("KEYCLOAK_AUDIENCE", "keycloak-onboarder-ui")
}

func TestProductionRequiresCredentialEncryptionKeys(t *testing.T) {
	setProductionEnvironment(t)
	t.Setenv("CREDENTIAL_ENCRYPTION_KEYS", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected production encryption key requirement")
	}
}

func TestInvalidCredentialEncryptionKeyFailsStartup(t *testing.T) {
	setProductionEnvironment(t)
	t.Setenv("CREDENTIAL_ENCRYPTION_KEYS", "v1:c2hvcnQ=")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid key to fail startup")
	}
}
