//go:build e2e

package frontend_test

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"multi-tenant-saas/test/e2e/harness"
)

func TestSPABuildArtifacts(t *testing.T) {
	root := harness.FindProjectRoot()
	dist := filepath.Join(root, "web", "dist")
	indexPath := filepath.Join(dist, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Skipf("web/dist is not built; run npm run build --prefix web to enable SPA smoke: %v", err)
	}
	html := string(content)
	if !strings.Contains(html, `<div id="root"`) {
		t.Fatalf("index.html does not look like SPA root: %s", indexPath)
	}
	assets := 0
	_ = filepath.WalkDir(filepath.Join(dist, "assets"), func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && (strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css")) {
			assets++
		}
		return nil
	})
	if assets == 0 {
		t.Fatalf("expected built JS/CSS assets under %s", filepath.Join(dist, "assets"))
	}
}

func TestSPALiveFallbackOptional(t *testing.T) {
	if os.Getenv("E2E_FRONTEND_LIVE") != "1" {
		t.Skip("set E2E_FRONTEND_LIVE=1 and E2E_BASE_URL to validate live SPA fallback")
	}
	base := harness.LoadConfig().BaseURL
	if base == "" {
		t.Fatal("E2E_BASE_URL is required when E2E_FRONTEND_LIVE=1")
	}
	resp, err := http.Get(strings.TrimRight(base, "/") + "/admin/system-config")
	if err != nil {
		t.Fatalf("GET SPA fallback: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected SPA fallback 200, got %d", resp.StatusCode)
	}
}
