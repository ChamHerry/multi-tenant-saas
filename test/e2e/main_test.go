//go:build e2e

package e2e_test

import (
	"testing"

	"multi-tenant-saas/test/e2e/harness"
)

func TestHarnessSmoke(t *testing.T) {
	cfg := harness.LoadConfig()
	if cfg.ProjectRoot == "" {
		t.Fatal("expected project root")
	}
	if cfg.LockPath == "" {
		t.Fatal("expected lock path")
	}
}
