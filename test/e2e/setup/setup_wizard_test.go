//go:build e2e

package setup_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/internal/testutil"
)

const setupPassword = "SetupPassword12345!"

func TestSetupWizard_GuardAndState(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	resetSetupState(t)

	state := suite.Client.GET("/api/v1/setup/state")
	testutil.AssertSuccess(t, state)
	data := state.JSONData()
	if data["requires_setup"] != true || data["initialized"] != false {
		t.Fatalf("expected setup required state, got %s", state.Body)
	}
	missing, _ := data["missing"].([]any)
	for _, key := range []string{"admin", "auth.session.secret", "auth.apiKey.secret"} {
		if !containsAnyString(missing, key) {
			t.Fatalf("expected missing %s in %v", key, missing)
		}
	}

	blocked := suite.Client.POST("/api/v1/auth/register", fmt.Sprintf(`{"email":"blocked@example.com","password":"%s","display_name":"Blocked"}`, testutil.TestPassword()))
	testutil.AssertStatus(t, blocked, http.StatusServiceUnavailable)
	if got := blocked.JSON()["code"]; got != service.SetupCodeRequired {
		t.Fatalf("expected setup required code, got %v body %s", got, blocked.Body)
	}

	ready := suite.Client.GET("/readyz")
	testutil.AssertStatus(t, ready, http.StatusOK)
	if got := ready.JSON()["code"]; got != service.SetupCodeRequired {
		t.Fatalf("expected readyz setup required code, got %v body %s", got, ready.Body)
	}
}

func TestSetupWizard_CompleteCreatesAdminSecretsAndLocks(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	resetSetupState(t)
	defer resetSetupRuntimeConfig(t)

	body := fmt.Sprintf(`{
		"admin":{"email":"setup-admin@example.com","password":"%s","display_name":"Setup Admin"},
		"runtime":{"web_base_url":"http://127.0.0.1:5173","generate_session_secret":true,"generate_api_key_secret":true}
	}`, setupPassword)
	complete := suite.Client.POST("/api/v1/setup/complete", body)
	testutil.AssertSuccess(t, complete)
	if strings.Contains(complete.Body, "change-me") || strings.Contains(complete.Body, setupPassword) {
		t.Fatalf("setup response leaked sensitive data: %s", complete.Body)
	}
	data := complete.JSONData()
	if data["initialized"] != true {
		t.Fatalf("expected initialized response, got %s", complete.Body)
	}
	user := data["user"].(map[string]any)
	if user["email"] != "setup-admin@example.com" || user["email_verified"] != true {
		t.Fatalf("unexpected setup admin user: %v", user)
	}

	if count := setupScalarInt(t, "SELECT COUNT(*) FROM system_setup WHERE id=1 AND status='initialized'"); count != 1 {
		t.Fatalf("expected one initialized system_setup row, got %d", count)
	}
	assertGeneratedSecretConfig(t, "auth.session.secret")
	assertGeneratedSecretConfig(t, "auth.apiKey.secret")

	second := suite.Client.POST("/api/v1/setup/complete", body)
	testutil.AssertStatus(t, second, http.StatusConflict)

	loginClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, loginClient, "setup-admin@example.com", setupPassword)
	adminConfig := loginClient.GET("/api/v1/admin/system-config/auth.session.secret")
	testutil.AssertSuccess(t, adminConfig)
	if strings.Contains(adminConfig.Body, service.Config().GetString(context.Background(), "auth.session.secret", "")) {
		t.Fatalf("admin system config leaked session secret: %s", adminConfig.Body)
	}

	audit := loginClient.GET("/api/v1/admin/audit-logs?action=system_setup.completed&resource_type=system_setup&limit=10")
	testutil.AssertSuccess(t, audit)
	if strings.Contains(audit.Body, service.Config().GetString(context.Background(), "auth.session.secret", "")) || strings.Contains(audit.Body, service.Config().GetString(context.Background(), "auth.apiKey.secret", "")) {
		t.Fatalf("audit response leaked generated secret: %s", audit.Body)
	}
}

func TestSetupWizard_LegacySelfHeal(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	resetSetupState(t)
	defer resetSetupRuntimeConfig(t)

	user, err := service.PasswordAuth().CreatePasswordUser(context.Background(), service.CreatePasswordUserInput{
		Email:         "legacy-admin@example.com",
		Password:      setupPassword,
		DisplayName:   "Legacy Admin",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("create legacy admin: %v", err)
	}
	if err = service.PlatformAdminService().Grant(context.Background(), user.ID, user.ID, "super_admin"); err != nil {
		t.Fatalf("grant legacy admin: %v", err)
	}
	if err = service.Config().Set(context.Background(), &service.ConfigSetParams{Key: "auth.session.secret", Value: "0123456789abcdef0123456789abcdef", ValueType: "secret", Description: "legacy safe session"}); err != nil {
		t.Fatalf("set legacy session secret: %v", err)
	}
	if err = service.Config().Set(context.Background(), &service.ConfigSetParams{Key: "auth.apiKey.secret", Value: "fedcba9876543210fedcba9876543210", ValueType: "secret", Description: "legacy safe api key"}); err != nil {
		t.Fatalf("set legacy api secret: %v", err)
	}

	if err = service.SystemSetup().EnsureLegacyState(context.Background()); err != nil {
		t.Fatalf("ensure legacy setup state: %v", err)
	}
	state, err := service.SystemSetup().State(context.Background())
	if err != nil {
		t.Fatalf("state after legacy self-heal: %v", err)
	}
	if !state.Initialized || state.RequiresSetup {
		t.Fatalf("expected initialized legacy state, got %+v", state)
	}
	if count := setupScalarInt(t, "SELECT COUNT(*) FROM system_setup WHERE id=1 AND metadata->>'legacy_detected'='true'"); count != 1 {
		t.Fatalf("expected legacy system_setup row, got %d", count)
	}
}

func resetSetupState(t *testing.T) {
	t.Helper()
	resetSetupRuntimeConfig(t)
	if _, err := g.DB().Exec(context.Background(), "DELETE FROM system_setup"); err != nil {
		t.Fatalf("delete setup row: %v", err)
	}
}

func resetSetupRuntimeConfig(t *testing.T) {
	t.Helper()
	statements := []struct{ key, value, valueType, description string }{
		{"web.baseUrl", "http://127.0.0.1:5173", "string", "Frontend base URL for invitation links"},
		{"auth.session.cookie.secure", "false", "bool", "Integration test HTTP cookie override"},
	}
	for _, item := range statements {
		_, err := g.DB().Exec(context.Background(), `
			INSERT INTO system_config(key, value, value_type, description, is_encrypted, created_at, updated_at)
			VALUES($1,$2,$3,$4,false,now(),now())
			ON CONFLICT(key) DO UPDATE SET value=$2, value_type=$3, description=$4, is_encrypted=false, updated_at=now()`, item.key, item.value, item.valueType, item.description)
		if err != nil {
			t.Fatalf("reset system config %s: %v", item.key, err)
		}
	}
	if _, err := g.DB().Exec(context.Background(), "DELETE FROM system_config WHERE key IN ('auth.session.secret', 'auth.apiKey.secret')"); err != nil {
		t.Fatalf("reset system secret config: %v", err)
	}
}

func assertGeneratedSecretConfig(t *testing.T, key string) {
	t.Helper()
	row, err := g.DB().GetOne(context.Background(), "SELECT value, value_type, is_encrypted FROM system_config WHERE key=$1", key)
	if err != nil || row.IsEmpty() {
		t.Fatalf("select generated secret %s: %v", key, err)
	}
	if row["value_type"].String() != "secret" || !row["is_encrypted"].Bool() {
		t.Fatalf("expected encrypted secret config for %s, got %v", key, row)
	}
	plain := service.Config().GetString(context.Background(), key, "")
	if len(plain) < 32 || strings.Contains(strings.ToLower(plain), "change-me") {
		t.Fatalf("expected safe plaintext for %s, got %q", key, plain)
	}
	if row["value"].String() == plain || strings.Contains(row["value"].String(), "change-me") {
		t.Fatalf("expected encrypted stored value for %s, got %q", key, row["value"].String())
	}
}

func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func setupScalarInt(t *testing.T, query string, args ...any) int {
	t.Helper()
	value, err := g.DB().GetValue(context.Background(), query, args...)
	if err != nil {
		t.Fatalf("scalar int query failed: %v", err)
	}
	return value.Int()
}
