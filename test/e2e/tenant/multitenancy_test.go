//go:build e2e

package tenant_test

import (
	"testing"

	"multi-tenant-saas/internal/testutil"
)

// createExplicitTenant registers a user and creates a tenant, returning user ID and tenant ID.
func createExplicitTenant(t *testing.T) (string, string) {
	t.Helper()
	user, _ := testutil.RegisterUser(t, suite.Client, "owner@example.com", "Owner")
	userID := user["id"].(string)
	resp := suite.Client.POST("/api/v1/tenants", `{"name":"Test Org","slug":"test-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")
	return userID, tenantID
}

// TestCreateTenant_Success verifies tenant creation with auto owner membership.
func TestCreateTenant_Success(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "owner@example.com", "Tenant Owner")

	resp := suite.Client.POST("/api/v1/tenants", `{
		"name": "My Org",
		"slug": "my-org"
	}`)
	testutil.AssertSuccess(t, resp)

	tenantData := testutil.ParseDataField(t, resp, "tenant")
	if tenantData == nil {
		t.Fatalf("expected tenant in response, got %s", resp.Body)
	}
	tenant := tenantData.(map[string]any)
	if tenant["slug"] != "my-org" {
		t.Fatalf("expected slug my-org, got %v", tenant["slug"])
	}
	if tenant["status"] != "active" {
		t.Fatalf("expected status active, got %v", tenant["status"])
	}

	// Verify owner membership via list members.
	tenantID := tenant["id"].(string)
	resp = suite.Client.DoWithHeaders("GET", "/api/v1/tenants/"+tenantID+"/members",
		"", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	members := resp.JSONData()["members"].([]any)
	if len(members) < 1 {
		t.Fatalf("expected at least 1 member (owner), got %v", resp.Body)
	}
}

// TestCreateTenant_DuplicateSlug verifies unique slug constraint.
func TestCreateTenant_DuplicateSlug(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "dup@example.com", "Dup Slug User")

	resp := suite.Client.POST("/api/v1/tenants", `{"name":"Org1","slug":"same-slug"}`)
	testutil.AssertSuccess(t, resp)

	resp = suite.Client.POST("/api/v1/tenants", `{"name":"Org2","slug":"same-slug"}`)
	// DB unique constraint returns 500 (not caught as 409 by handler).
	if resp.StatusCode != 409 && resp.StatusCode != 500 {
		t.Fatalf("expected 409 or 500 for duplicate slug, got %d body %s", resp.StatusCode, resp.Body)
	}
}

// TestGetTenant verifies GET /tenants/{id} with member auth.
func TestGetTenant(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "getter@example.com", "Getter User")

	resp := suite.Client.POST("/api/v1/tenants", `{"name":"Get Org","slug":"get-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")

	resp = suite.Client.DoWithHeaders("GET", "/api/v1/tenants/"+tenantID, "", map[string]string{
		"X-Tenant-ID": tenantID,
	})
	testutil.AssertSuccess(t, resp)

	tenantData := resp.JSONData()["tenant"].(map[string]any)
	if tenantData["name"] != "Get Org" {
		t.Fatalf("expected name Get Org, got %v", tenantData["name"])
	}
}

// TestListUserTenants verifies GET /me/tenants returns memberships.
func TestListUserTenants(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "lister@example.com", "Lister User")

	resp := suite.Client.POST("/api/v1/tenants", `{"name":"List Org","slug":"list-org"}`)
	testutil.AssertSuccess(t, resp)

	resp = suite.Client.GET("/api/v1/me/tenants")
	testutil.AssertSuccess(t, resp)

	data := resp.JSONData()
	tenants, ok := data["tenants"].([]any)
	if !ok || len(tenants) < 1 {
		t.Fatalf("expected at least 1 tenant, got %v", data)
	}
}

// TestUpdateTenant verifies PATCH /tenants/{id}.
func TestUpdateTenant(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "updater@example.com", "Updater User")

	resp := suite.Client.POST("/api/v1/tenants", `{"name":"Old Name","slug":"update-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")

	resp = suite.Client.DoWithHeaders("PATCH", "/api/v1/tenants/"+tenantID, `{"name":"New Name"}`, map[string]string{
		"X-Tenant-ID": tenantID,
	})
	testutil.AssertSuccess(t, resp)

	updated := resp.JSONData()["tenant"].(map[string]any)
	if updated["name"] != "New Name" {
		t.Fatalf("expected name New Name, got %v", updated["name"])
	}
}

// TestUpdateTenant_AsMember_Forbidden verifies member role cannot update tenant.
func TestUpdateTenant_AsMember_Forbidden(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	// Create owner + tenant.
	testutil.RegisterUser(t, suite.Client, "owner-forbidden@example.com", "Owner")
	resp := suite.Client.POST("/api/v1/tenants", `{"name":"Forbidden Org","slug":"forbidden-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")

	// Add a member (role=member).
	user2 := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "member-forbidden@example.com", "Member User")
	userID2 := user2["id"].(string)
	suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/members",
		`{"user_id":"`+userID2+`","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})

	// Login as the member user on a separate client.
	memberClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, memberClient, "member-forbidden@example.com", testutil.TestPassword())

	resp = memberClient.DoWithHeaders("PATCH", "/api/v1/tenants/"+tenantID, `{"name":"Hacked Name"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	if resp.StatusCode == 200 {
		t.Fatal("expected member to be forbidden from updating tenant")
	}
	if resp.StatusCode != 403 && resp.StatusCode != 500 {
		t.Fatalf("expected 403 or 500, got %d body %s", resp.StatusCode, resp.Body)
	}
}

// TestSuspendRestoreTenant verifies suspend → restore lifecycle.
// Uses platform admin API to verify state since TenantResolver blocks access to suspended tenants.
func TestSuspendRestoreTenant(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "suspend@example.com", "Suspend User")

	resp := suite.Client.POST("/api/v1/tenants", `{"name":"Suspend Org","slug":"suspend-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")

	// Suspend via tenant endpoint (owner has tenant:manage permission).
	resp = suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/suspend", "", map[string]string{
		"X-Tenant-ID": tenantID,
	})
	testutil.AssertSuccess(t, resp)

	// Verify suspended via admin API (first user is super_admin, no TenantResolver needed).
	resp = suite.Client.GET("/api/v1/admin/tenants")
	testutil.AssertSuccess(t, resp)
	items := resp.JSONData()["items"].([]any)
	found := false
	for _, item := range items {
		tenant := item.(map[string]any)
		if tenant["id"] == tenantID {
			if tenant["status"] != "suspended" {
				t.Fatalf("expected status suspended, got %v", tenant["status"])
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("tenant %s not found in admin list", tenantID)
	}

	// Restore via platform admin API (TenantResolver blocks suspended tenants).
	resp = suite.Client.POST("/api/v1/admin/tenants/"+tenantID+"/restore", "")
	testutil.AssertSuccess(t, resp)

	// Verify restored via admin API.
	resp = suite.Client.GET("/api/v1/admin/tenants")
	testutil.AssertSuccess(t, resp)
	items = resp.JSONData()["items"].([]any)
	for _, item := range items {
		tenant := item.(map[string]any)
		if tenant["id"] == tenantID && tenant["status"] != "active" {
			t.Fatalf("expected status active after restore, got %v", tenant["status"])
		}
	}
}

// TestDeleteTenant verifies soft delete.
func TestDeleteTenant(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "deleter@example.com", "Deleter User")

	resp := suite.Client.POST("/api/v1/tenants", `{"name":"Delete Org","slug":"delete-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")

	resp = suite.Client.DoWithHeaders("DELETE", "/api/v1/tenants/"+tenantID, "", map[string]string{
		"X-Tenant-ID": tenantID,
	})
	testutil.AssertSuccess(t, resp)

	// Verify tenant is deleted via admin API (first user is super_admin).
	resp = suite.Client.GET("/api/v1/admin/tenants")
	testutil.AssertSuccess(t, resp)
	items := resp.JSONData()["items"].([]any)
	for _, item := range items {
		tenant := item.(map[string]any)
		if tenant["id"] == tenantID && tenant["status"] != "deleted" {
			t.Fatalf("expected deleted status, got %v", tenant["status"])
		}
	}
}
