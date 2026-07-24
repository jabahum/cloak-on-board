package audit

import (
	"reflect"
	"testing"
)

func TestRedactRecursivelyRemovesSecrets(t *testing.T) {
	input := map[string]any{
		"name":          "portal",
		"client_secret": "hidden",
		"nested": map[string]any{
			"access_token": "token",
			"values":       []any{map[string]any{"password": "password"}},
		},
	}
	got := Redact(input).(map[string]any)
	want := map[string]any{
		"name":          "portal",
		"client_secret": "[REDACTED]",
		"nested": map[string]any{
			"access_token": "[REDACTED]",
			"values":       []any{map[string]any{"password": "[REDACTED]"}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Redact() = %#v", got)
	}
}
