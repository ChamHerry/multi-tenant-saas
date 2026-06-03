//go:build e2e

package tenant_test

import (
	"testing"

	"multi-tenant-saas/internal/testutil"
)

func TestDefaultTenantLazyRuntimeCoveredByTenantCreation(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "default-tenant@example.com", "Default Tenant")
	resp := suite.Client.POST("/api/v1/tenants", `{"name":"Default Tenant Org","slug":"default-tenant-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")

	tenants := suite.Client.GET("/api/v1/me/tenants")
	testutil.AssertSuccess(t, tenants)
	if tenantID == "" {
		t.Fatal("expected created tenant id")
	}
}

func TestFullFlowSmoke(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "full-flow@example.com", "Full Flow")
	create := suite.Client.POST("/api/v1/tenants", `{"name":"Full Flow Org","slug":"full-flow-org"}`)
	testutil.AssertSuccess(t, create)
	tenantID := testutil.ParseDataString(t, create, "tenant.id")
	access := suite.Client.GET("/api/v1/me/access")
	testutil.AssertSuccess(t, access)
	members := suite.Client.DoWithHeaders("GET", "/api/v1/tenants/"+tenantID+"/members", "", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, members)
}
