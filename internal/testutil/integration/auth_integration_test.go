//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/testutil"
)

// TestRegister_Success verifies the full register chain.
func TestRegister_Success(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	resp := suite.Client.POST("/api/v1/auth/register", `{
		"email": "test@example.com",
		"password": "SuperSecure12345!",
		"display_name": "Test User"
	}`)
	testutil.AssertSuccess(t, resp)

	user := testutil.ParseResponseUser(t, resp)
	if user["email"] != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %v", user["email"])
	}
	if user["status"] != "active" {
		t.Fatalf("expected status active, got %v", user["status"])
	}
	testutil.AssertCookieSet(t, resp, "saas_template_session")
	testutil.AssertCookieSet(t, resp, "saas_template_csrf")

	// Verify first user is super_admin.
	resp = suite.Client.GET("/api/v1/me/access")
	testutil.AssertSuccess(t, resp)
	if pac, ok := resp.JSONData()["platform_admin"].(map[string]any); !ok || pac["role"] != "super_admin" {
		t.Fatalf("expected first user to be super_admin, got %v", resp.JSONData())
	}
}

// TestRegister_DuplicateEmail verifies 409 on duplicate email.
func TestRegister_DuplicateEmail(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	resp := suite.Client.POST("/api/v1/auth/register", `{
		"email": "dup@example.com",
		"password": "SuperSecure12345!",
		"display_name": "First User"
	}`)
	testutil.AssertSuccess(t, resp)

	resp = suite.Client.POST("/api/v1/auth/register", `{
		"email": "dup@example.com",
		"password": "AnotherPassword1!",
		"display_name": "Second User"
	}`)
	testutil.AssertStatus(t, resp, 409)
}

// TestRegister_PasswordTooShort verifies password validation.
func TestRegister_PasswordTooShort(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	resp := suite.Client.POST("/api/v1/auth/register", `{
		"email": "short@example.com",
		"password": "short",
		"display_name": "Short Pwd"
	}`)
	testutil.AssertStatus(t, resp, 400)
}

// TestLogin_Success verifies the login chain.
func TestLogin_Success(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "login@example.com"
	// Register with a temp client so main client stays clean.
	testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), email, "Login User")

	// Login with the main client (no prior session).
	loginClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp := loginClient.POST("/api/v1/auth/login", fmt.Sprintf(`{
		"email": "%s",
		"password": "%s"
	}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, resp)

	user := testutil.ParseResponseUser(t, resp)
	if user["email"] != email {
		t.Fatalf("expected email %s, got %v", email, user["email"])
	}

	// Session should be accessible.
	resp = loginClient.GET("/api/v1/auth/session")
	testutil.AssertSuccess(t, resp)
}

// TestLogin_WrongPassword verifies failed login returns error.
func TestLogin_WrongPassword(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "wrong@example.com", "Wrong Pwd User")

	resp := suite.Client.POST("/api/v1/auth/login", `{
		"email": "wrong@example.com",
		"password": "WrongPassword12345!"
	}`)
	// GoFrame maps CodeNotAuthorized to 403.
	if resp.StatusCode != 401 && resp.StatusCode != 403 {
		t.Fatalf("expected 401 or 403 for wrong password, got %d body %s", resp.StatusCode, resp.Body)
	}
}

// TestLogin_RateLimited verifies rate limiting kicks in after burst exhaustion.
func TestLogin_RateLimited(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	// Use a fresh client so the rate limit bucket starts from scratch.
	// Rate limiter is keyed by IP (no auth), so localhost gets a single bucket.
	rlClient := testutil.NewTestClient(t, suite.Client.BaseURL())

	// Send requests to exhaust the burst bucket (config: burst=40, rps=20).
	// Using GET /api/v1/me which is cheap but goes through RateLimit middleware.
	got429 := false
	for i := 0; i < 60; i++ {
		resp := rlClient.GET("/api/v1/me")
		if resp.StatusCode == 429 {
			got429 = true
			if resp.Headers.Get("Retry-After") == "" {
				t.Fatal("expected Retry-After header on 429 response")
			}
			break
		}
	}
	if !got429 {
		t.Fatal("expected 429 rate limited response after 60 requests, never got one")
	}
}

// TestLogout_Success verifies the logout chain.
func TestLogout_Success(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "logout@example.com", "Logout User")

	resp := suite.Client.GET("/api/v1/auth/session")
	testutil.AssertSuccess(t, resp)

	resp = suite.Client.POST("/api/v1/auth/logout", "")
	testutil.AssertSuccess(t, resp)

	resp = suite.Client.GET("/api/v1/auth/session")
	if resp.StatusCode == 200 {
		data := resp.JSONData()
		if data != nil && data["user"] != nil {
			t.Fatal("expected session to be revoked after logout")
		}
	}
}

// TestSession_AfterLogout verifies unauthenticated access returns error.
func TestSession_AfterLogout(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "session@example.com", "Session User")
	suite.Client.POST("/api/v1/auth/logout", "")
	suite.Client.ResetCookies()

	resp := suite.Client.GET("/api/v1/auth/session")
	testutil.AssertStatus(t, resp, 401)
}

// TestMeEndpoint verifies GET /me returns the authenticated user.
func TestMeEndpoint(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "me@example.com", "Me User")
	resp := suite.Client.GET("/api/v1/me")
	testutil.AssertSuccess(t, resp)

	user := resp.JSONData()["user"].(map[string]any)
	if user["email"] != "me@example.com" {
		t.Fatalf("expected email me@example.com, got %v", user["email"])
	}
}

// TestMe_Unauthenticated verifies GET /me returns 401 without auth.
func TestMe_Unauthenticated(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	resp := suite.Client.GET("/api/v1/me")
	testutil.AssertStatus(t, resp, 401)
}

// TestChangePassword verifies password change flow:
// old password stops working, new password works.
func TestChangePassword(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "changepw@example.com"
	testutil.RegisterUser(t, suite.Client, email, "Change PW User")

	// Change password via the authenticated endpoint.
	newPassword := "BrandNewPassword123!"
	resp := suite.Client.POST("/api/v1/auth/password/change", fmt.Sprintf(`{
		"old_password": "%s",
		"new_password": "%s"
	}`, testutil.TestPassword(), newPassword))
	testutil.AssertSuccess(t, resp)

	// Old session is still valid after password change (session-based auth).
	// Verify by accessing /me.
	resp = suite.Client.GET("/api/v1/me")
	testutil.AssertSuccess(t, resp)

	// Old password should no longer work.
	suite.Client.ResetCookies()
	oldPwClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp = oldPwClient.POST("/api/v1/auth/login", fmt.Sprintf(`{
		"email": "%s",
		"password": "%s"
	}`, email, testutil.TestPassword()))
	if resp.StatusCode == 200 {
		t.Fatal("expected login with old password to fail after change")
	}

	// New password should work.
	newPwClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp = newPwClient.POST("/api/v1/auth/login", fmt.Sprintf(`{
		"email": "%s",
		"password": "%s"
	}`, email, newPassword))
	testutil.AssertSuccess(t, resp)
}

// TestLogin_Lockout verifies account lockout after N failed attempts.
func TestLogin_Lockout(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "lockout@example.com"
	testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), email, "Lockout User")

	// Fail 5 times (LOCKOUT_MAX_ATTEMPTS = 5)
	for i := 0; i < 5; i++ {
		resp := suite.Client.POST("/api/v1/auth/login", fmt.Sprintf(`{
			"email": "%s",
			"password": "WrongPassword12345!"
		}`, email))
		if resp.StatusCode != 401 && resp.StatusCode != 403 {
			t.Fatalf("attempt %d: expected 401/403, got %d body=%s", i+1, resp.StatusCode, resp.Body)
		}
	}

	// Now even the correct password should fail (lock is active)
	resp := suite.Client.POST("/api/v1/auth/login", fmt.Sprintf(`{
		"email": "%s",
		"password": "%s"
	}`, email, testutil.TestPassword()))
	if resp.StatusCode != 401 && resp.StatusCode != 403 {
		t.Fatalf("expected lockout to reject correct password, got %d body=%s", resp.StatusCode, resp.Body)
	}
}

// TestLogin_WindowReset verifies failures outside the window are ignored.
func TestLogin_WindowReset(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "window@example.com"
	testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), email, "Window User")

	// Fail once
	resp := suite.Client.POST("/api/v1/auth/login", fmt.Sprintf(`{
		"email": "%s",
		"password": "WrongPassword12345!"
	}`, email))
	if resp.StatusCode != 401 && resp.StatusCode != 403 {
		t.Fatal("expected first failure to return 401/403")
	}

	// Manually set last_failed_at to 30 minutes ago so window has expired
	_, err := g.DB().Exec(context.Background(),
		"UPDATE auth_login_attempts SET last_failed_at = NOW() - INTERVAL '30 minutes' WHERE login_key = $1",
		email,
	)
	if err != nil {
		t.Fatalf("failed to backdate last_failed_at: %v", err)
	}

	// Now 4 more failures should NOT cause lockout (counter resets to 1 after window)
	for i := 0; i < 4; i++ {
		resp := suite.Client.POST("/api/v1/auth/login", fmt.Sprintf(`{
			"email": "%s",
			"password": "WrongPassword12345!"
		}`, email))
		if resp.StatusCode != 401 && resp.StatusCode != 403 {
			t.Fatalf("expected 401/403 for failure, got %d", resp.StatusCode)
		}
	}

	// Correct password should still work (only 4 failures in window + 1 old)
	resp = suite.Client.POST("/api/v1/auth/login", fmt.Sprintf(`{
		"email": "%s",
		"password": "%s"
	}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, resp)
}

// TestAdmin_UnlockUser verifies the admin unlock endpoint.
func TestAdmin_UnlockUser(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	// Register the admin user first (first user gets super_admin automatically)
	testutil.RegisterUser(t, suite.Client, "admin@example.com", "Admin User")

	// Register a target user with a separate client
	targetEmail := "unlockme@example.com"
	targetUser := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), targetEmail, "Unlock Me")
	targetID := targetUser["id"].(string)

	// Lock the target account by failing 5 times
	targetClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	for i := 0; i < 5; i++ {
		targetClient.POST("/api/v1/auth/login", fmt.Sprintf(`{
			"email": "%s",
			"password": "WrongPassword12345!"
		}`, targetEmail))
	}

	// Verify target is locked (correct password fails)
	resp := targetClient.POST("/api/v1/auth/login", fmt.Sprintf(`{
		"email": "%s",
		"password": "%s"
	}`, targetEmail, testutil.TestPassword()))
	if resp.StatusCode != 401 && resp.StatusCode != 403 {
		t.Fatalf("expected lockout to reject correct password, got %d", resp.StatusCode)
	}

	// Unlock via admin endpoint (main client is the super_admin)
	resp = suite.Client.POST(fmt.Sprintf("/api/v1/admin/users/%s/unlock", targetID), "")
	testutil.AssertSuccess(t, resp)

	// Now target user should be able to log in
	resp = targetClient.POST("/api/v1/auth/login", fmt.Sprintf(`{
		"email": "%s",
		"password": "%s"
	}`, targetEmail, testutil.TestPassword()))
	testutil.AssertSuccess(t, resp)
}
