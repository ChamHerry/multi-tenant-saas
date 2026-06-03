//go:build e2e

package harness

import "testing"

func RequireOneOfStatus(t *testing.T, got int, allowed ...int) {
	t.Helper()
	for _, status := range allowed {
		if got == status {
			return
		}
	}
	t.Fatalf("status %d not in allowed set %v", got, allowed)
}
