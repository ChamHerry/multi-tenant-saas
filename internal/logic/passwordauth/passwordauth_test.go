package passwordauth

import (
	"context"
	"testing"
	"time"

	"multi-tenant-saas/internal/service"
)

type passwordAuthTestConfig struct{}

func (passwordAuthTestConfig) GetString(_ context.Context, _ string, defaultVal string) string {
	return defaultVal
}
func (passwordAuthTestConfig) GetInt(_ context.Context, _ string, defaultVal int) int {
	return defaultVal
}
func (passwordAuthTestConfig) GetFloat(_ context.Context, _ string, defaultVal float64) float64 {
	return defaultVal
}
func (passwordAuthTestConfig) GetBool(_ context.Context, _ string, defaultVal bool) bool {
	return defaultVal
}
func (passwordAuthTestConfig) GetDuration(_ context.Context, _ string, defaultVal time.Duration) time.Duration {
	return defaultVal
}
func (passwordAuthTestConfig) GetStrings(_ context.Context, keys []string) (map[string]string, error) {
	return make(map[string]string, len(keys)), nil
}
func (passwordAuthTestConfig) Set(_ context.Context, _ *service.ConfigSetParams) error { return nil }
func (passwordAuthTestConfig) Delete(_ context.Context, _ string) error                { return nil }

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  User@Example.COM "); got != "user@example.com" {
		t.Fatalf("normalizeEmail=%q", got)
	}
}

func TestValidatePassword(t *testing.T) {
	service.RegisterConfig(passwordAuthTestConfig{})
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
