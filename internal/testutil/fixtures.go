package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
)

const testPassword = "TestPassword12345!" // Satisfies minLength=15

// TestPassword returns the default test password used by RegisterUser/RegisterAdditionalUser.
func TestPassword() string { return testPassword }

// RegisterUser creates a user via the register endpoint on the given client.
// The client's session cookies are updated to the new user.
func RegisterUser(t *testing.T, client *TestClient, email, displayName string) (map[string]any, string) {
	t.Helper()
	body := fmt.Sprintf(`{"email":"%s","password":"%s","display_name":"%s"}`,
		email, testPassword, displayName)
	resp := client.POST("/api/v1/auth/register", body)
	AssertSuccess(t, resp)
	data := resp.JSONData()
	if data == nil || data["user"] == nil {
		t.Fatalf("register user: unexpected response %s", resp.Body)
	}
	user, _ := data["user"].(map[string]any)
	return user, testPassword
}

// RegisterAdditionalUser creates a user with a fresh client so the main client's
// session is not affected. Returns the user map from the response.
// Use this to create users that will be added as members, invitees, etc.
func RegisterAdditionalUser(t *testing.T, baseURL, email, displayName string) map[string]any {
	t.Helper()
	tmpClient := NewTestClient(t, baseURL)
	body := fmt.Sprintf(`{"email":"%s","password":"%s","display_name":"%s"}`,
		email, testPassword, displayName)
	resp := tmpClient.POST("/api/v1/auth/register", body)
	AssertSuccess(t, resp)
	data := resp.JSONData()
	if data == nil || data["user"] == nil {
		t.Fatalf("register additional user: unexpected response %s", resp.Body)
	}
	user, _ := data["user"].(map[string]any)
	return user
}

// LoginUser authenticates via the login endpoint and returns the parsed response data.
func LoginUser(t *testing.T, client *TestClient, email, password string) map[string]any {
	t.Helper()
	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	resp := client.POST("/api/v1/auth/login", body)
	AssertSuccess(t, resp)
	data := resp.JSONData()
	if data == nil {
		t.Fatalf("login user: unexpected response %s", resp.Body)
	}
	return data
}

// CreateTenant calls the service directly (bypassing HTTP) to create a tenant.
func CreateTenant(ctx context.Context, t *testing.T, ownerUserID, name, slug string) *service.Tenant {
	t.Helper()
	tenant, err := service.TenantProvision().CreateTenant(ctx, service.CreateTenantInput{
		Name:        name,
		Slug:        slug,
		OwnerUserID: ownerUserID,
	})
	if err != nil {
		t.Fatalf("create tenant %s: %v", slug, err)
	}
	return tenant
}

// AddMember calls the service directly to add a member to a tenant.
func AddMember(ctx context.Context, t *testing.T, tenantID, userID, role string) *service.TenantMembership {
	t.Helper()
	m, err := service.TenantMembershipService().AddMember(ctx, service.AddTenantMemberInput{
		TenantID: tenantID,
		UserID:   userID,
		Role:     role,
		Status:   "active",
	})
	if err != nil {
		t.Fatalf("add member %s to %s: %v", userID, tenantID, err)
	}
	return m
}

// ExpireInvitationDB directly sets an invitation's expires_at to the past
// and status to expired. Uses parameterized query to prevent SQL injection.
func ExpireInvitationDB(invitationID string) error {
	_, err := g.DB().Exec(context.Background(),
		"UPDATE tenant_invitations SET expires_at = NOW() - INTERVAL '1 day', status = 'expired', updated_at = NOW() WHERE id = $1",
		invitationID,
	)
	return err
}
