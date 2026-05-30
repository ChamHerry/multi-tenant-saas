//go:build integration

package integration_test

import (
	"testing"

	"multi-tenant-saas/internal/testutil"
)

// TestCreateAPIKey_Success verifies API key creation returns raw_key.
func TestCreateAPIKey_Success(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/api-keys",
		`{"name":"Test Key","scopes":["tenant:read","member:read"]}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)

	data := resp.JSONData()
	if data["raw_key"] == nil || data["raw_key"] == "" {
		t.Fatal("expected raw_key in response")
	}

	apiKey, ok := data["api_key"].(map[string]any)
	if !ok || apiKey["name"] != "Test Key" {
		t.Fatalf("expected api_key with name Test Key, got %v", data)
	}
}

// TestAPIKeyAuthenticate verifies using an API key for authentication.
func TestAPIKeyAuthenticate(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/api-keys",
		`{"name":"Auth Key","scopes":["tenant:read","member:read"]}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	rawKey := resp.JSONData()["raw_key"].(string)

	// Use a fresh client with the API key as Bearer token.
	// GET /me requires auth scope; include tenant header with matching scope.
	apiClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp = apiClient.DoWithHeaders("GET", "/api/v1/me", "",
		map[string]string{"Authorization": "Bearer " + rawKey})
	// API key auth may not have user:read scope — just verify it authenticates.
	if resp.StatusCode == 401 {
		t.Fatalf("API key failed to authenticate: %s", resp.Body)
	}
	// 200 or 403 (scope mismatch) both mean auth succeeded.
	if resp.StatusCode != 200 && resp.StatusCode != 403 {
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, resp.Body)
	}
}

// TestAPIKeyRevoke verifies that a revoked key cannot be used.
func TestAPIKeyRevoke(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/api-keys",
		`{"name":"Revoke Key","scopes":["tenant:read"]}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	rawKey := resp.JSONData()["raw_key"].(string)
	keyID := resp.JSONData()["api_key"].(map[string]any)["id"].(string)

	// Revoke.
	resp = suite.Client.DoWithHeaders("DELETE",
		"/api/v1/tenants/"+tenantID+"/api-keys/"+keyID,
		"", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)

	// Try using the revoked key.
	apiClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp = apiClient.DoWithHeaders("GET", "/api/v1/me", "",
		map[string]string{"Authorization": "Bearer " + rawKey})
	if resp.StatusCode == 200 {
		t.Fatal("expected revoked key to fail authentication")
	}
}

// TestListAPIKeys verifies listing keys for a tenant.
func TestListAPIKeys(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/api-keys",
		`{"name":"Key 1","scopes":["tenant:read"]}`,
		map[string]string{"X-Tenant-ID": tenantID})
	suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/api-keys",
		`{"name":"Key 2","scopes":["member:read"]}`,
		map[string]string{"X-Tenant-ID": tenantID})

	resp := suite.Client.DoWithHeaders("GET", "/api/v1/tenants/"+tenantID+"/api-keys",
		"", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)

	keys, ok := resp.JSONData()["api_keys"].([]any)
	if !ok || len(keys) != 2 {
		t.Fatalf("expected 2 api keys, got %v", resp.Body)
	}
}
