//go:build integration

package integration_test

import (
	"os"
	"testing"

	"multi-tenant-saas/internal/testutil"
)

var suite *testutil.TestSuite

func TestMain(m *testing.M) {
	// Fake T for setup logging — TestMain has no *testing.T.
	t := &testing.T{}
	suite = testutil.NewTestSuite(t)

	code := m.Run()

	suite.TeardownSuite(t)
	os.Exit(code)
}
