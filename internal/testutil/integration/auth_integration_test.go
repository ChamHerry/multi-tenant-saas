//go:build integration

package integration_test

import (
	"fmt"
	"testing"

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
