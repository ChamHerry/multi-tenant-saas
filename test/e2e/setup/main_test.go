//go:build e2e

package setup_test

import (
	"os"
	"testing"

	"multi-tenant-saas/internal/testutil"
	"multi-tenant-saas/test/e2e/harness"
)

var suite *testutil.TestSuite

func TestMain(m *testing.M) {
	os.Exit(harness.RunPackage(m, "setup", func(ts *testutil.TestSuite) { suite = ts }))
}
