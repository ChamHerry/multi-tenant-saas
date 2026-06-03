//go:build e2e

package harness_test

import (
	"net/http"
	"testing"

	"multi-tenant-saas/internal/testutil"
	"multi-tenant-saas/test/e2e/harness"
)

func TestHarnessConfigAndConcurrency(t *testing.T) {
	cfg := harness.LoadConfig()
	if cfg.ProjectRoot == "" || cfg.LockPath == "" || cfg.ArtifactsDir == "" {
		t.Fatalf("incomplete config: %+v", cfg)
	}
	seen := make(chan int, 2)
	harness.RunConcurrent(2, func(i int) { seen <- i })
	close(seen)
	if len(seen) != 2 {
		t.Fatalf("expected two concurrent callbacks, got %d", len(seen))
	}
}

func TestHarnessSuiteSmoke(t *testing.T) {
	rt := harness.StartPackage(t, "harness")
	defer rt.Finish(t)
	rt.Suite.SetupTest(t)
	defer rt.Suite.TeardownTest(t)
	resp := rt.Suite.Client.GET("/readyz")
	testutil.AssertStatus(t, resp, http.StatusOK)
}
