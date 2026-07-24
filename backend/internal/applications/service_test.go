package applications

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestApplicationJSONDoesNotExposeClientSecret(t *testing.T) {
	payload, err := json.Marshal(Application{KeycloakClientSecret: "sensitive"})
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) == "" || containsText(string(payload), "sensitive") || containsText(string(payload), "keycloak_client_secret") {
		t.Fatalf("application JSON exposed client secret: %s", payload)
	}
}

func containsText(value, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}

func TestValidateApplication(t *testing.T) {
	if err := validateApplication("Portal", "portal", "frontend", "owner@example.org"); err != nil {
		t.Fatalf("valid application: %v", err)
	}
	for _, tc := range []struct {
		name, slug, appType, email string
	}{
		{"", "portal", "frontend", ""},
		{"Portal", "", "frontend", ""},
		{"Portal", "portal", "unknown", ""},
		{"Portal", "portal", "frontend", "not-an-email"},
	} {
		if err := validateApplication(tc.name, tc.slug, tc.appType, tc.email); !errors.Is(err, ErrValidation) {
			t.Fatalf("expected validation error for %#v, got %v", tc, err)
		}
	}
}

func TestCleanValuesTrimsDropsEmptyAndDeduplicates(t *testing.T) {
	got := cleanValues([]string{" admin ", "", "viewer", "admin"})
	want := []string{"admin", "viewer"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cleanValues() = %#v, want %#v", got, want)
	}
}
