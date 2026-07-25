package delivery

import (
	"encoding/json"
	"testing"

	"github.com/jabahum/keycloak-onboarder/backend/internal/keycloak"
)

func TestCanonicalConfigurationOrdering(t *testing.T) {
	first := SnapshotConfiguration{ClientID: "demo", Enabled: true, Roles: sorted([]string{"writer", "reader"}), RedirectURIs: []string{}}
	second := SnapshotConfiguration{ClientID: "demo", Enabled: true, Roles: sorted([]string{"reader", "writer"}), RedirectURIs: []string{}}
	firstHash, _, err := canonicalHash(first)
	if err != nil {
		t.Fatal(err)
	}
	secondHash, _, err := canonicalHash(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstHash != secondHash {
		t.Fatalf("canonical hashes differ: %s != %s", firstHash, secondHash)
	}
}

func TestOverridesAreRestricted(t *testing.T) {
	if err := validateOverrides(json.RawMessage(`{"enabled":false,"redirect_uris":["https://example.test/callback"]}`)); err != nil {
		t.Fatal(err)
	}
	if err := validateOverrides(json.RawMessage(`{"client_id":"takeover"}`)); err == nil {
		t.Fatal("expected client_id override to be rejected")
	}
}

func TestDriftFindingsIgnoreUnmanagedFields(t *testing.T) {
	desired := []byte(`{"client_id":"demo","enabled":true,"roles":["reader"]}`)
	actual := []byte(`{"client_id":"demo","enabled":false,"roles":["reader"],"keycloak_internal":"ignored"}`)
	findings := compareJSON(desired, actual)
	if len(findings) != 1 || findings[0].Path != "/enabled" || findings[0].ChangeType != "changed" {
		t.Fatalf("unexpected findings: %#v", findings)
	}
}

func TestApplyOverridesDoesNotMutateIdentity(t *testing.T) {
	config := SnapshotConfiguration{ClientID: "demo", Enabled: true}
	result, err := applyOverrides(config, json.RawMessage(`{"enabled":false}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.ClientID != "demo" || result.Enabled {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestManagedScopeCanonicalizationIgnoresUnmanagedScopes(t *testing.T) {
	actual := keycloak.ClientScopeAssignments{
		Default: []keycloak.ClientScope{{Name: "roles"}, {Name: "unmanaged"}, {Name: "email"}},
	}
	desired := []ScopeConfiguration{{Name: "email", Type: "default"}, {Name: "roles", Type: "default"}}
	result := managedScopeConfiguration(actual, desired)
	if len(result) != 2 || result[0].Name != "email" || result[1].Name != "roles" {
		t.Fatalf("unexpected managed scopes: %#v", result)
	}
}
