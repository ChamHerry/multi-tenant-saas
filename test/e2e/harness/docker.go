//go:build e2e

package harness

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"
)

func ManageDocker(t *testing.T, action string) {
	t.Helper()
	if action == "" || action == "skip" {
		return
	}
	switch action {
	case "start", "restart", "ready":
	default:
		t.Fatalf("unsupported E2E_DOCKER_ACTION=%q", action)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "./scripts/manage-docker.sh", action)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("manage docker %s failed: %v\n%s", action, err, string(out))
	}
	t.Logf("manage docker %s: %s", action, string(out))
}

func DockerLogs(t *testing.T, service string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	args := []string{"compose", "logs", "--tail=120"}
	if service != "" {
		args = append(args, service)
	}
	out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("docker logs failed: %v\n%s", err, string(out))
	}
	return string(out)
}
