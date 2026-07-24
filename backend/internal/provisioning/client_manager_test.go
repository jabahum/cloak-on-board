package provisioning

import (
	"errors"
	"testing"

	"github.com/jabahum/keycloak-onboarder/backend/internal/applications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/keycloak"
)

func TestInferApplicationType(t *testing.T) {
	cases := []struct {
		client keycloak.KeycloakClient
		want   string
	}{
		{keycloak.KeycloakClient{ServiceAccountsEnabled: true}, "machine_to_machine"},
		{keycloak.KeycloakClient{PublicClient: true, Attributes: map[string]string{"pkce.code.challenge.method": "S256"}}, "frontend"},
		{keycloak.KeycloakClient{PublicClient: true}, "mobile"},
		{keycloak.KeycloakClient{}, "backend"},
	}
	for _, tc := range cases {
		if got := inferApplicationType(tc.client); got != tc.want {
			t.Fatalf("inferApplicationType(%#v) = %q, want %q", tc.client, got, tc.want)
		}
	}
}

func TestValidateMapper(t *testing.T) {
	valid := keycloak.UserAttributeMapper("department", "department", "department")
	if err := validateMapper(valid); err != nil {
		t.Fatalf("valid mapper: %v", err)
	}
	invalid := valid
	invalid.Config = map[string]string{}
	if err := validateMapper(invalid); !errors.Is(err, applications.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
