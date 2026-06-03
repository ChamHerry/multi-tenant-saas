//go:build e2e

package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"multi-tenant-saas/internal/testutil"
)

type Runtime struct {
	Suite   *testutil.TestSuite
	cfg     Config
	lock    *os.File
	pkgName string
}

func StartPackage(t *testing.T, pkgName string) *Runtime {
	t.Helper()
	cfg := LoadConfig()
	if err := os.MkdirAll(filepath.Dir(cfg.LockPath), 0o755); err != nil {
		t.Fatalf("create e2e lock dir: %v", err)
	}
	lock, err := os.OpenFile(cfg.LockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("open e2e lock %s: %v", cfg.LockPath, err)
	}
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		_ = lock.Close()
		t.Fatalf("acquire e2e package lock: %v", err)
	}
	if err = os.MkdirAll(filepath.Join(cfg.ArtifactsDir, pkgName), 0o755); err != nil {
		_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		_ = lock.Close()
		t.Fatalf("create artifacts dir: %v", err)
	}
	rt := &Runtime{Suite: testutil.NewTestSuite(t), cfg: cfg, lock: lock, pkgName: pkgName}
	return rt
}

func (r *Runtime) Finish(t *testing.T) {
	t.Helper()
	if r == nil {
		return
	}
	if r.Suite != nil {
		r.Suite.TeardownSuite(t)
	}
	if r.lock != nil {
		_ = syscall.Flock(int(r.lock.Fd()), syscall.LOCK_UN)
		_ = r.lock.Close()
	}
}

func RunPackage(m interface{ Run() int }, pkgName string, assign func(*testutil.TestSuite)) int {
	t := &testing.T{}
	rt := StartPackage(t, pkgName)
	if assign != nil {
		assign(rt.Suite)
	}
	code := m.Run()
	rt.Finish(t)
	return code
}

func (r *Runtime) ArtifactPath(name string) string {
	safe := filepath.Base(name)
	return filepath.Join(r.cfg.ArtifactsDir, r.pkgName, safe)
}

func (r *Runtime) String() string {
	if r == nil || r.Suite == nil || r.Suite.Client == nil {
		return "e2e runtime <nil>"
	}
	return fmt.Sprintf("e2e runtime package=%s base=%s", r.pkgName, r.Suite.Client.BaseURL())
}
