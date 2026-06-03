//go:build e2e

package harness

import (
	"os"
	"testing"
)

func WriteArtifact(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artifact %s: %v", path, err)
	}
}
