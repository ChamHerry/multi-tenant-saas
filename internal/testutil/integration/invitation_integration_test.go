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

// TestCreateInvitation_ExistingMember verifies inviting a user who is already a member fails.
func TestCreateInvitation_ExistingMember(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	// The owner is already a member — try inviting their email.
	// Register the owner with a known email via the main client.
	// owner@example.com is the owner (see createTestTenant).

	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"owner@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	// Should reject because the user is already a member.
	if resp.StatusCode == 200 {
		t.Fatal("expected error when inviting existing member")
	}
	if resp.StatusCode != 400 && resp.StatusCode != 409 && resp.StatusCode != 500 {
		t.Fatalf("expected 400/409/500 for existing member, got %d body %s", resp.StatusCode, resp.Body)
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

// TestAcceptInvitation_WrongUser verifies that a different user cannot accept an invitation.
func TestAcceptInvitation_WrongUser(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	// Invite accept@example.com.
	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"accept@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	token := resp.JSONData()["token"].(string)

	// Register a DIFFERENT user and try to accept.
	wrongClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.RegisterUser(t, wrongClient, "wrong-acceptor@example.com", "Wrong Acceptor")

	resp = wrongClient.POST("/api/v1/invitations/accept",
		`{"token":"`+token+`"}`)
	// Should fail because email doesn't match.
	if resp.StatusCode == 200 {
		t.Fatal("expected wrong user to fail accepting invitation")
	}
	if resp.StatusCode != 400 && resp.StatusCode != 401 && resp.StatusCode != 403 {
		t.Fatalf("expected 400/401/403 for wrong user accept, got %d body %s", resp.StatusCode, resp.Body)
	}
}

// TestAcceptInvitation_Expired verifies that an expired invitation cannot be accepted.
func TestAcceptInvitation_Expired(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	_, tenantID := createTestTenant(t)

	// Create invitation.
	resp := suite.Client.DoWithHeaders("POST", "/api/v1/tenants/"+tenantID+"/invitations",
		`{"invitee_email":"expired@example.com","role":"member"}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	token := resp.JSONData()["token"].(string)
	invitationID := resp.JSONData()["invitation"].(map[string]any)["id"].(string)

	// Expire the invitation directly in DB (bypassing the 7-day TTL).
	err := testutil.ExpireInvitationDB(invitationID)
	if err != nil {
		t.Fatalf("expire invitation: %v", err)
	}

	// Register the invited user.
	expiredClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.RegisterUser(t, expiredClient, "expired@example.com", "Expired User")

	// Try to accept — should fail.
	resp = expiredClient.POST("/api/v1/invitations/accept",
		`{"token":"`+token+`"}`)
	if resp.StatusCode == 200 {
		t.Fatal("expected expired invitation to be rejected")
	}
	if resp.StatusCode != 400 && resp.StatusCode != 404 {
		t.Fatalf("expected 400 or 404 for expired invitation, got %d body %s", resp.StatusCode, resp.Body)
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
