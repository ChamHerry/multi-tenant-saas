package passwordauth

import (
	"context"
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  User@Example.COM "); got != "user@example.com" {
		t.Fatalf("normalizeEmail=%q", got)
	}
}

func TestValidatePassword(t *testing.T) {
	ctx := context.Background()
	if err := validatePassword(ctx, "short"); err == nil {
		t.Fatal("validatePassword short expected error")
	}
	if err := validatePassword(ctx, "0123456789abcde"); err != nil {
		t.Fatalf("validatePassword valid error = %v", err)
	}
	long := "0123456789abcde0123456789abcde0123456789abcde0123456789abcde0123456789abcde"
	if err := validatePassword(ctx, long); err == nil {
		t.Fatal("validatePassword >72 bytes expected error")
	}
}

func TestShouldBootstrapFirstPlatformAdmin(t *testing.T) {
	if !shouldBootstrapFirstPlatformAdmin(0) {
		t.Fatal("empty users table should trigger first platform admin bootstrap")
	}
	for _, count := range []int{-1, 1, 2, 100} {
		if shouldBootstrapFirstPlatformAdmin(count) {
			t.Fatalf("user count %d should not trigger bootstrap", count)
		}
	}
}
