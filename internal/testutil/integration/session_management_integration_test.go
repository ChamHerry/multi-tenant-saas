//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"multi-tenant-saas/internal/testutil"
)

func TestSessionManagement_ListAndRevokeOtherSession(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "sessions-list@example.com"
	testutil.RegisterUser(t, suite.Client, email, "Sessions List")

	secondClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, secondClient, email, testutil.TestPassword())

	resp := suite.Client.GET("/api/v1/me/sessions")
	testutil.AssertSuccess(t, resp)
	sessions := sessionItems(t, resp)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 active sessions, got %d body=%s", len(sessions), resp.Body)
	}
	currentID := currentSessionID(t, sessions)
	otherID := otherSessionID(t, sessions)
	assertSessionItemSanitized(t, sessions[0])

	resp = suite.Client.DELETE("/api/v1/me/sessions/" + currentID)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)
	testutil.AssertSuccess(t, suite.Client.GET("/api/v1/auth/session"))

	resp = suite.Client.DELETE("/api/v1/me/sessions/" + otherID)
	testutil.AssertSuccess(t, resp)
	testutil.AssertStatus(t, secondClient.GET("/api/v1/auth/session"), http.StatusUnauthorized)
	testutil.AssertSuccess(t, suite.Client.GET("/api/v1/auth/session"))
}

func TestSessionManagement_RevokeOtherSessionsKeepsCurrent(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "sessions-revoke-others@example.com"
	testutil.RegisterUser(t, suite.Client, email, "Sessions Batch")

	secondClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, secondClient, email, testutil.TestPassword())
	thirdClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, thirdClient, email, testutil.TestPassword())

	resp := suite.Client.DELETE("/api/v1/me/sessions")
	testutil.AssertSuccess(t, resp)
	data := resp.JSONData()
	if got := int(data["revoked_count"].(float64)); got != 2 {
		t.Fatalf("expected revoked_count=2, got %d body=%s", got, resp.Body)
	}
	testutil.AssertSuccess(t, suite.Client.GET("/api/v1/auth/session"))
	testutil.AssertStatus(t, secondClient.GET("/api/v1/auth/session"), http.StatusUnauthorized)
	testutil.AssertStatus(t, thirdClient.GET("/api/v1/auth/session"), http.StatusUnauthorized)
}

func TestSessionManagement_CannotRevokeForeignSession(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	testutil.RegisterUser(t, suite.Client, "session-owner@example.com", "Session Owner")

	foreignClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.RegisterUser(t, foreignClient, "session-foreign@example.com", "Session Foreign")
	foreignSessions := sessionItems(t, foreignClient.GET("/api/v1/me/sessions"))
	foreignID := currentSessionID(t, foreignSessions)

	resp := suite.Client.DELETE("/api/v1/me/sessions/" + foreignID)
	testutil.AssertStatus(t, resp, http.StatusForbidden)
	testutil.AssertSuccess(t, foreignClient.GET("/api/v1/auth/session"))
}

func TestSessionManagement_DeleteRequiresCSRF(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "sessions-csrf@example.com"
	testutil.RegisterUser(t, suite.Client, email, "Sessions CSRF")
	secondClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, secondClient, email, testutil.TestPassword())

	sessions := sessionItems(t, suite.Client.GET("/api/v1/me/sessions"))
	otherID := otherSessionID(t, sessions)
	suite.Client.ClearCSRF()

	resp := suite.Client.DELETE("/api/v1/me/sessions/" + otherID)
	testutil.AssertStatus(t, resp, http.StatusForbidden)
	testutil.AssertSuccess(t, secondClient.GET("/api/v1/auth/session"))
}

func sessionItems(t *testing.T, resp *testutil.TestResponse) []map[string]any {
	t.Helper()
	testutil.AssertSuccess(t, resp)
	raw, ok := resp.JSONData()["sessions"].([]any)
	if !ok {
		t.Fatalf("sessions response missing array: %s", resp.Body)
	}
	items := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("session item is not object: %#v", item)
		}
		items = append(items, m)
	}
	return items
}

func currentSessionID(t *testing.T, sessions []map[string]any) string {
	t.Helper()
	for _, session := range sessions {
		if current, _ := session["is_current"].(bool); current {
			return stringField(t, session, "session_id")
		}
	}
	t.Fatalf("current session not found in %#v", sessions)
	return ""
}

func otherSessionID(t *testing.T, sessions []map[string]any) string {
	t.Helper()
	for _, session := range sessions {
		if current, _ := session["is_current"].(bool); !current {
			return stringField(t, session, "session_id")
		}
	}
	t.Fatalf("other session not found in %#v", sessions)
	return ""
}

func assertSessionItemSanitized(t *testing.T, session map[string]any) {
	t.Helper()
	for _, forbidden := range []string{"secret_hash", "csrf_hash", "token"} {
		if _, ok := session[forbidden]; ok {
			t.Fatalf("session response leaked %s: %#v", forbidden, session)
		}
	}
	if stringField(t, session, "session_id") == "" {
		t.Fatal("session_id is required")
	}
	if stringField(t, session, "last_active_at") == "" || stringField(t, session, "created_at") == "" || stringField(t, session, "expires_at") == "" {
		t.Fatalf("session timestamps are required: %#v", session)
	}
	if ip := stringField(t, session, "ip"); ip == "" || ip == "127.0.0.1" {
		t.Fatalf("expected masked ip, got %q", ip)
	}
	device, ok := session["device"].(map[string]any)
	if !ok {
		t.Fatalf("device is required: %#v", session)
	}
	if stringField(t, device, "browser") == "" || stringField(t, device, "os") == "" {
		t.Fatalf("device browser/os are required: %#v", device)
	}
}

func stringField(t *testing.T, object map[string]any, key string) string {
	t.Helper()
	value, ok := object[key].(string)
	if !ok {
		t.Fatalf("%s is not a string in %s", key, fmt.Sprint(object))
	}
	return value
}
