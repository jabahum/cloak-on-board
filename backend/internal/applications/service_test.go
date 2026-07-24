package applications

import (
	"errors"
	"reflect"
	"testing"
)

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
