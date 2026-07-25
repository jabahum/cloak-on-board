package approvals

import "testing"

func TestSecretRotationIsNeverRetried(t *testing.T) {
	if retrySafe("rotate_client_secret") {
		t.Fatal("secret rotation must not be retryable")
	}
	for _, action := range []string{"deploy_application", "promote_application", "reconcile_drift"} {
		if !retrySafe(action) {
			t.Fatalf("%s should be retryable", action)
		}
	}
}
