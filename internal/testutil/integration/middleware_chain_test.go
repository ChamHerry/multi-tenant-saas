//go:build integration

package integration_test

import (
	"strings"
	"testing"

	"multi-tenant-saas/internal/testutil"
)

// TestHealthCheck verifies /healthz and /readyz endpoints.
func TestHealthCheck(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	resp := suite.Client.GET("/healthz")
	testutil.AssertStatus(t, resp, 200)

	resp = suite.Client.GET("/readyz")
	testutil.AssertStatus(t, resp, 200)
	if !strings.Contains(resp.Body, `"ok":true`) {
		t.Fatalf("expected readyz ok=true, got %s", resp.Body)
	}
}

// TestAuthRequired verifies that /api/v1/me returns 401 without authentication.
func TestAuthRequired(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	client := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp := client.GET("/api/v1/me")
	testutil.AssertStatus(t, resp, 401)
}

// TestTenantResolverWithHeader verifies X-Tenant-ID header resolves tenant context.
func TestTenantResolverWithHeader(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "tenantr@example.com", "Tenant R User")
	resp := suite.Client.POST("/api/v1/tenants", `{"name":"TR Org","slug":"tr-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")

	// With X-Tenant-ID header → should succeed.
	resp = suite.Client.DoWithHeaders("GET", "/api/v1/tenants/"+tenantID+"/members",
		"", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
}

// TestRBAC_OwnerCanDelete verifies owner can remove members.
func TestRBAC_OwnerCanDelete(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	user3 := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "rbac3@example.com", "RBAC Target")
	userID3 := user3["id"].(string)
	suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/members",
		`{"user_id":"`+userID3+`","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})

	resp := suite.Client.DoWithHeaders("DELETE",
		"/api/v1/tenants/"+tenantID+"/members/"+userID3,
		"", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
}

// TestFullChain_PublicRegister verifies public routes don't require auth/CSRF.
func TestFullChain_PublicRegister(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	freshClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp := freshClient.POST("/api/v1/auth/register", `{
		"email": "chain@example.com",
		"password": "ChainPassword123!",
		"display_name": "Chain User"
	}`)
	testutil.AssertSuccess(t, resp)
}

// TestMetricsEndpoint verifies /metrics returns Prometheus metrics.
func TestMetricsEndpoint(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	// GoFrame mounts metrics handler at /metrics/ (with trailing slash).
	resp := suite.Client.GET("/metrics")
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 from /metrics, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Body, "http_requests_total") {
		bodyPreview := resp.Body
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200]
		}
		t.Fatalf("expected http_requests_total metric, got %s", bodyPreview)
	}
}
