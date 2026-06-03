//go:build e2e

package harness

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	ProjectRoot  string
	BaseURL      string
	DockerAction string
	LockPath     string
	ArtifactsDir string
}

func LoadConfig() Config {
	root := FindProjectRoot()
	artifacts := envOrDefault("E2E_ARTIFACTS_DIR", filepath.Join(root, ".local", "e2e-artifacts"))
	return Config{
		ProjectRoot:  root,
		BaseURL:      envOrDefault("E2E_BASE_URL", ""),
		DockerAction: envOrDefault("E2E_DOCKER_ACTION", "skip"),
		LockPath:     envOrDefault("E2E_LOCK_PATH", filepath.Join(root, ".local", "e2e-go.lock")),
		ArtifactsDir: artifacts,
	}
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func FindProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
