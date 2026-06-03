package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"

	"multi-tenant-saas/internal/controller"
	"multi-tenant-saas/internal/middleware"
)

// TestSuite manages the integration test lifecycle: database, HTTP server, per-test isolation.
//
// Isolation strategy: each test starts with a clean database (all tables truncated).
// Tests commit data to the real test database, and DB assertions work normally.
// The next test's SetupTest truncates everything again.
type TestSuite struct {
	ctx    context.Context
	server *ghttp.Server
	Client *TestClient
}

var (
	suiteIdx int
	suiteMu  sync.Mutex
)

// NewTestSuite creates a TestSuite with a fresh database, migrated schema, and HTTP server.
func NewTestSuite(t *testing.T) *TestSuite {
	t.Helper()

	// Point GoFrame config to the project root's manifest/config directory.
	projectRoot := findProjectRoot()
	configPath := filepath.Join(projectRoot, "manifest/config")
	adapter, err := gcfg.NewAdapterFile(filepath.Join(configPath, "config.yaml"))
	if err != nil {
		t.Fatalf("create config adapter from %s: %v", configPath, err)
	}
	g.Cfg().SetAdapter(adapter)

	ctx := context.Background()
	SetupTestDB(t)

	// Unique server name so parallel packages don't collide.
	suiteMu.Lock()
	suiteIdx++
	name := fmt.Sprintf("itest-%d", suiteIdx)
	suiteMu.Unlock()

	s := g.Server(name)
	s.SetPort(0) // Use OS-assigned random port to avoid conflicts.
	registerTestRoutes(s)

	if err = s.Start(); err != nil {
		t.Fatalf("start test server: %v", err)
	}

	port := s.GetListenedPort()
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	client := NewTestClient(t, baseURL)

	ts := &TestSuite{
		ctx:    ctx,
		server: s,
		Client: client,
	}
	SetCurrentSuite(ts)
	return ts
}

// TeardownSuite stops the HTTP server. Call from TestMain after m.Run().
func (ts *TestSuite) TeardownSuite(t *testing.T) {
	t.Helper()
	if ts.server != nil {
		ts.server.Shutdown()
	}
}

// SetupTest truncates all tables for a clean starting point.
// Returns a context for direct DAO assertions if needed.
func (ts *TestSuite) SetupTest(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()
	TruncateAllTables(ctx, t)
	MarkSystemSetupInitialized(ctx, t)
	return ctx
}

// TeardownTest is a no-op in the truncate-based isolation model.
// Data is cleaned up by the next test's SetupTest.
func (ts *TestSuite) TeardownTest(t *testing.T) {
	t.Helper()
	// No-op: next SetupTest will truncate.
}

// registerTestRoutes mirrors cmd.registerRoutes.
func registerTestRoutes(s *ghttp.Server) {
	// Root routes: healthz, readyz, hello.
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.RequestContext)
		group.Middleware(middleware.HandlerResponse)
		controller.RegisterRootRoutes(group)
	})

	// API v1 routes.
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.RequestContext)
		group.Middleware(middleware.HandlerResponse)
		group.Middleware(middleware.Tracing)
		group.Middleware(middleware.Metrics)
		controller.RegisterRoutes(group)
	})

	// Metrics endpoint (mirrors cmd.registerRoutes).
	s.Group("/metrics", func(group *ghttp.RouterGroup) {
		group.GET("/", middleware.MetricsHandler)
	})
}

// currentSuite is the active TestSuite, set by SetCurrentSuite.
var currentSuite *TestSuite

// SetCurrentSuite makes the suite available for test helpers.
func SetCurrentSuite(ts *TestSuite) {
	currentSuite = ts
}

// findProjectRoot walks up from the working directory to find the directory
// containing go.mod (the project root).
func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}
