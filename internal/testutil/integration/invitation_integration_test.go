//go:build integration

package integration_test

import (
	"testing"

	"multi-tenant-saas/internal/testutil"
)

// TestCreateInvitation_Success verifies creating an invitation.
func TestCreateInvitation_Success(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"invite@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)

	data := resp.JSONData()
	invitation, ok := data["invitation"].(map[string]any)
	if !ok {
		t.Fatalf("expected invitation in response, got %s", resp.Body)
	}
	if invitation["status"] != "pending" {
		t.Fatalf("expected status pending, got %v", invitation["status"])
	}
	if data["token"] == nil || data["token"] == "" {
		t.Fatal("expected token to be returned")
	}
}

// TestAcceptInvitation_Success verifies accepting creates a membership.
func TestAcceptInvitation_Success(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	// Create invitation.
	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"accept@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	token := resp.JSONData()["token"].(string)

	// Register the invited user with a separate client.
	acceptClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.RegisterUser(t, acceptClient, "accept@example.com", "Accept User")

	// Accept the invitation.
	resp = acceptClient.POST("/api/v1/invitations/accept",
		`{"token":"`+token+`"}`)
	testutil.AssertSuccess(t, resp)

	member := resp.JSONData()["member"].(map[string]any)
	if member["role"] != "member" {
		t.Fatalf("expected role member, got %v", member["role"])
	}
}

// TestDeclineInvitation verifies declining an invitation.
func TestDeclineInvitation(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	// Create invitation.
	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"decline@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	invitationID := resp.JSONData()["invitation"].(map[string]any)["id"].(string)

	// Register and decline with separate client.
	declineClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.RegisterUser(t, declineClient, "decline@example.com", "Decline User")

	resp = declineClient.POST("/api/v1/me/invitations/"+invitationID+"/decline", "")
	testutil.AssertSuccess(t, resp)
}

// TestRevokeInvitation verifies revoking an invitation.
func TestRevokeInvitation(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"revoke@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	invitationID := resp.JSONData()["invitation"].(map[string]any)["id"].(string)

	resp = suite.Client.DoWithHeaders("POST",
		"/api/v1/tenants/"+tenantID+"/invitations/"+invitationID+"/revoke",
		"", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
}

// TestResendInvitation verifies resending generates a new token.
func TestResendInvitation(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"resend@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	invitationID := resp.JSONData()["invitation"].(map[string]any)["id"].(string)

	resp = suite.Client.DoWithHeaders("POST",
		"/api/v1/tenants/"+tenantID+"/invitations/"+invitationID+"/resend",
		"", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)

	data := resp.JSONData()
	if data["token"] == nil || data["token"] == "" {
		t.Fatal("expected new token from resend")
	}
}

// TestListTenantInvitations verifies listing invitations for a tenant.
func TestListTenantInvitations(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"list1@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"list2@example.com","role":"viewer"}`,
		map[string]string{"X-Tenant-ID": tenantID})

	resp := suite.Client.DoWithHeaders("GET", "/api/v1/tenants/"+tenantID+"/invitations",
		"", map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)

	data := resp.JSONData()
	invitations, ok := data["invitations"].([]any)
	if !ok || len(invitations) != 2 {
		t.Fatalf("expected 2 invitations, got %v", data)
	}
}
