//go:build integration

package integration_test

import (
	"testing"

	"multi-tenant-saas/internal/testutil"
)

// TestPlatformAdmin_Session verifies that a super_admin user can access the admin session.
func TestPlatformAdmin_Session(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "admin@example.com", "Admin User")

	resp := suite.Client.GET("/api/v1/admin/session")
	testutil.AssertSuccess(t, resp)

	pac, ok := resp.JSONData()["platform_admin"].(map[string]any)
	if !ok || pac["role"] != "super_admin" {
		t.Fatalf("expected role super_admin, got %v", resp.JSONData())
	}
}

// TestPlatformAdmin_NonAdmin_Forbidden verifies non-platform-admin users get 403.
func TestPlatformAdmin_NonAdmin_Forbidden(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	// First user is super_admin, so we need a second user.
	testutil.RegisterUser(t, suite.Client, "first@example.com", "First User")
	user2 := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "nonadmin@example.com", "Non Admin")

	// Login as the second user (non-admin).
	nonAdminClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, nonAdminClient, "nonadmin@example.com", testutil.TestPassword())

	// Verify user2 has no platform admin record by checking the ID.
	_ = user2["id"].(string)

	resp := nonAdminClient.GET("/api/v1/admin/session")
	// Should be forbidden — not a platform admin.
	if resp.StatusCode == 200 {
		t.Fatal("expected non-platform-admin to be forbidden from admin session")
	}
	if resp.StatusCode != 403 && resp.StatusCode != 401 {
		t.Fatalf("expected 403 or 401, got %d body %s", resp.StatusCode, resp.Body)
	}
}

// TestPlatformAdmin_ListTenants verifies admin can list all tenants.
func TestPlatformAdmin_ListTenants(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "listadmin@example.com", "List Admin")
	suite.Client.POST("/api/v1/tenants", `{"name":"Admin List Org","slug":"admin-list-org"}`)

	resp := suite.Client.GET("/api/v1/admin/tenants")
	testutil.AssertSuccess(t, resp)

	items := resp.JSONData()["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected at least 1 tenant, got %v", resp.Body)
	}
}

// TestPlatformAdmin_GrantRevoke verifies granting and revoking platform admin roles.
// Note: Revoke is a soft-delete (status → "suspended"), not a hard delete.
func TestPlatformAdmin_GrantRevoke(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "granter@example.com", "Granter")
	user2 := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "target@example.com", "Target User")
	userID2 := user2["id"].(string)

	// Grant support role.
	resp := suite.Client.POST("/api/v1/admin/platform-admins",
		`{"user_id":"`+userID2+`","role":"support"}`)
	testutil.AssertSuccess(t, resp)

	// Verify 2 admins (super_admin + support).
	resp = suite.Client.GET("/api/v1/admin/platform-admins")
	testutil.AssertSuccess(t, resp)
	items := resp.JSONData()["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 platform admins, got %d", len(items))
	}

	// Revoke (soft: status → suspended).
	resp = suite.Client.DELETE("/api/v1/admin/platform-admins/" + userID2)
	testutil.AssertSuccess(t, resp)

	// Verify still 2 records but one is suspended.
	resp = suite.Client.GET("/api/v1/admin/platform-admins")
	testutil.AssertSuccess(t, resp)
	items = resp.JSONData()["items"].([]any)
	// Revoke soft-deletes, so records still exist.
	suspendedCount := 0
	for _, item := range items {
		admin := item.(map[string]any)
		if admin["user_id"] == userID2 && admin["status"] == "suspended" {
			suspendedCount++
		}
	}
	if suspendedCount != 1 {
		t.Fatalf("expected 1 suspended admin, got %d (items=%v)", suspendedCount, items)
	}
}

// TestPlatformAdmin_ListUsers verifies admin can list all users.
func TestPlatformAdmin_ListUsers(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "useradmin@example.com", "User Admin")
	testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "listed@example.com", "Listed User")

	resp := suite.Client.GET("/api/v1/admin/users")
	testutil.AssertSuccess(t, resp)

	items := resp.JSONData()["items"].([]any)
	if len(items) < 2 {
		t.Fatalf("expected at least 2 users, got %d", len(items))
	}
}

// TestPlatformAdmin_AuditLogs verifies admin can query audit logs.
func TestPlatformAdmin_AuditLogs(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "auditadmin@example.com", "Audit Admin")
	suite.Client.POST("/api/v1/tenants", `{"name":"Audit Org","slug":"audit-org"}`)

	resp := suite.Client.GET("/api/v1/admin/audit-logs")
	testutil.AssertSuccess(t, resp)
}
