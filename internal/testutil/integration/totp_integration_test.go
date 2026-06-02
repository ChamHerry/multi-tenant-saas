//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	ptotp "github.com/pquerna/otp/totp"

	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/internal/testutil"
)

func TestTOTP_FullLoginFlow(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp@example.com"
	testutil.RegisterUser(t, suite.Client, email, "TOTP User")

	setupResp := suite.Client.POST("/api/v1/me/totp/setup", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	testutil.AssertSuccess(t, setupResp)
	secret := testutil.ParseDataString(t, setupResp, "secret")
	if secret == "" {
		t.Fatal("expected setup secret")
	}

	setupCode, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate setup TOTP code: %v", err)
	}
	enableResp := suite.Client.POST("/api/v1/me/totp/enable", fmt.Sprintf(`{"code":"%s"}`, setupCode))
	testutil.AssertSuccess(t, enableResp)
	backupCodes, ok := enableResp.JSONData()["backup_codes"].([]any)
	if !ok || len(backupCodes) != 10 {
		t.Fatalf("expected 10 backup codes, got %v", enableResp.Body)
	}

	loginClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	loginResp := loginClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	loginData := loginResp.JSONData()
	if loginData["requires_2fa"] != true {
		t.Fatalf("expected requires_2fa=true, got %v", loginData)
	}
	token, ok := loginData["totp_token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected totp_token, got %v", loginData)
	}
	preVerifySession := loginClient.GET("/api/v1/auth/session")
	testutil.AssertStatus(t, preVerifySession, 401)

	verifyCode, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate verify TOTP code: %v", err)
	}
	verifyResp := loginClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"%s"}`, token, verifyCode))
	testutil.AssertSuccess(t, verifyResp)
	testutil.AssertCookieSet(t, verifyResp, "saas_template_session")
	testutil.AssertCookieSet(t, verifyResp, "saas_template_csrf")

	sessionResp := loginClient.GET("/api/v1/auth/session")
	testutil.AssertSuccess(t, sessionResp)
}

func TestTOTP_BackupCodeLogin(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp-backup@example.com"
	testutil.RegisterUser(t, suite.Client, email, "TOTP Backup User")

	setupResp := suite.Client.POST("/api/v1/me/totp/setup", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	testutil.AssertSuccess(t, setupResp)
	secret := testutil.ParseDataString(t, setupResp, "secret")
	setupCode, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate setup TOTP code: %v", err)
	}
	enableResp := suite.Client.POST("/api/v1/me/totp/enable", fmt.Sprintf(`{"code":"%s"}`, setupCode))
	testutil.AssertSuccess(t, enableResp)
	backupCodes := enableResp.JSONData()["backup_codes"].([]any)
	backupCode := backupCodes[0].(string)

	backupClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	loginResp := backupClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	token := loginResp.JSONData()["totp_token"].(string)

	verifyResp := backupClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"%s"}`, token, backupCode))
	testutil.AssertSuccess(t, verifyResp)
	remaining, ok := verifyResp.JSONData()["backup_codes_remaining"].(float64)
	if !ok || remaining != 9 {
		t.Fatalf("expected 9 remaining backup codes, got %v", verifyResp.Body)
	}
}

func TestTOTP_BackupCodeReuseAfterSuccessFails(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp-backup-reuse@example.com"
	_, _, backupCodes := setupAndEnableTOTP(t, email)
	backupCode := backupCodes[0]

	firstClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	loginResp := firstClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	token := loginResp.JSONData()["totp_token"].(string)
	verifyResp := firstClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"%s"}`, token, backupCode))
	testutil.AssertSuccess(t, verifyResp)

	secondClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	loginResp = secondClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	token = loginResp.JSONData()["totp_token"].(string)
	reuseResp := secondClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"%s"}`, token, backupCode))
	if reuseResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected reused backup code to fail with 403, got %d body=%s", reuseResp.StatusCode, reuseResp.Body)
	}
}

func TestTOTP_APIKeyCannotWriteTOTPSettings(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp-api-key-denied@example.com"
	testutil.RegisterUser(t, suite.Client, email, "TOTP API Key User")
	rawKey := createUserSecurityReadAPIKey(t)

	apiClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	bearer := map[string]string{"Authorization": "Bearer " + rawKey}

	statusResp := apiClient.DoWithHeaders(http.MethodGet, "/api/v1/me/totp/status", "", bearer)
	testutil.AssertSuccess(t, statusResp)

	writeRequests := []struct {
		name string
		path string
		body string
	}{
		{name: "setup", path: "/api/v1/me/totp/setup", body: fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword())},
		{name: "enable", path: "/api/v1/me/totp/enable", body: `{"code":"123456"}`},
		{name: "disable", path: "/api/v1/me/totp/disable", body: fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword())},
		{name: "regenerate", path: "/api/v1/me/totp/backup-codes/regenerate", body: fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword())},
	}
	for _, tt := range writeRequests {
		resp := apiClient.DoWithHeaders(http.MethodPost, tt.path, tt.body, bearer)
		if resp.StatusCode == http.StatusOK {
			t.Fatalf("%s: expected API-key TOTP write to be denied, got 200 body=%s", tt.name, resp.Body)
		}
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401/403, got %d body=%s", tt.name, resp.StatusCode, resp.Body)
		}
	}
}

func TestTOTP_EnableConcurrentRequestsAreSerialized(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp-enable-race@example.com"
	testutil.RegisterUser(t, suite.Client, email, "TOTP Enable Race")

	setupResp := suite.Client.POST("/api/v1/me/totp/setup", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	testutil.AssertSuccess(t, setupResp)
	secret := testutil.ParseDataString(t, setupResp, "secret")

	secondSession := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, secondSession, email, testutil.TestPassword())

	code, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate enable TOTP code: %v", err)
	}

	clients := []*testutil.TestClient{suite.Client, secondSession}
	statuses := runConcurrentPosts(clients, "/api/v1/me/totp/enable", fmt.Sprintf(`{"code":"%s"}`, code))
	assertExactlyOneStatus(t, statuses, http.StatusOK)

	unused, err := g.DB().Model("user_totp_backup_codes").Ctx(context.Background()).Where("used_at IS NULL").Count()
	if err != nil {
		t.Fatalf("count unused backup codes: %v", err)
	}
	if unused != 10 {
		t.Fatalf("expected exactly 10 unused backup codes after concurrent enable, got %d (statuses=%v)", unused, statuses)
	}
}

func TestTOTP_BackupCodeConcurrentReuseRejected(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp-backup-race@example.com"
	_, _, backupCodes := setupAndEnableTOTP(t, email)
	backupCode := backupCodes[0]

	clients := []*testutil.TestClient{
		testutil.NewTestClient(t, suite.Client.BaseURL()),
		testutil.NewTestClient(t, suite.Client.BaseURL()),
	}
	bodies := make([]string, 0, len(clients))
	for _, client := range clients {
		loginResp := client.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
		testutil.AssertSuccess(t, loginResp)
		token := loginResp.JSONData()["totp_token"].(string)
		bodies = append(bodies, fmt.Sprintf(`{"totp_token":"%s","code":"%s"}`, token, backupCode))
	}

	statuses := runConcurrentClientPosts(clients, "/api/v1/auth/verify-totp", bodies)
	assertExactlyOneStatus(t, statuses, http.StatusOK)

	remaining, err := g.DB().Model("user_totp_backup_codes").Ctx(context.Background()).Where("used_at IS NULL").Count()
	if err != nil {
		t.Fatalf("count remaining backup codes: %v", err)
	}
	if remaining != 9 {
		t.Fatalf("expected backup-code remaining count 9 after one successful reuse, got %d (statuses=%v)", remaining, statuses)
	}

	sessions, err := g.DB().Model("auth_sessions").Ctx(context.Background()).Count()
	if err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	// Register/setup/enable creates one original browser session; backup-code verification should add exactly one.
	if sessions != 2 {
		t.Fatalf("expected exactly 2 sessions total after concurrent backup-code verify, got %d (statuses=%v)", sessions, statuses)
	}
}

func TestTOTP_PasswordStepDoesNotRecordFinalLoginSideEffects(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp-login-side-effects@example.com"
	userID, _, _ := setupAndEnableTOTP(t, email)

	beforeUserLastLogin := scalarString(t, "SELECT COALESCE(last_login_at::text, '') FROM users WHERE id = $1", userID)
	beforeIdentityLastLogin := scalarString(t, "SELECT COALESCE(last_login_at::text, '') FROM user_identities WHERE user_id = $1 AND provider = 'password'", userID)
	beforeSuccessAudits := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE user_id = $1 AND action = 'auth.login.success'", userID)

	loginClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	loginResp := loginClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	if loginResp.JSONData()["requires_2fa"] != true {
		t.Fatalf("expected requires_2fa response, got %s", loginResp.Body)
	}

	afterUserLastLogin := scalarString(t, "SELECT COALESCE(last_login_at::text, '') FROM users WHERE id = $1", userID)
	afterIdentityLastLogin := scalarString(t, "SELECT COALESCE(last_login_at::text, '') FROM user_identities WHERE user_id = $1 AND provider = 'password'", userID)
	afterSuccessAudits := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE user_id = $1 AND action = 'auth.login.success'", userID)

	if afterUserLastLogin != beforeUserLastLogin {
		t.Fatalf("users.last_login_at changed before TOTP verification: before=%q after=%q", beforeUserLastLogin, afterUserLastLogin)
	}
	if afterIdentityLastLogin != beforeIdentityLastLogin {
		t.Fatalf("identity.last_login_at changed before TOTP verification: before=%q after=%q", beforeIdentityLastLogin, afterIdentityLastLogin)
	}
	if afterSuccessAudits != beforeSuccessAudits {
		t.Fatalf("auth.login.success audit changed before TOTP verification: before=%d after=%d", beforeSuccessAudits, afterSuccessAudits)
	}
}

func TestTOTP_VerifyRateLimitAppliesAcrossInvalidCodes(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp-rate-limit@example.com"
	setupAndEnableTOTP(t, email)

	loginClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	loginResp := loginClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	token := loginResp.JSONData()["totp_token"].(string)

	for i := 0; i < 5; i++ {
		resp := loginClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"000000"}`, token))
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("attempt %d: expected invalid TOTP to be 403 before limiter, got %d body=%s", i+1, resp.StatusCode, resp.Body)
		}
	}

	limited := loginClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"000000"}`, token))
	if limited.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected TOTP verify limiter to return 429, got %d body=%s", limited.StatusCode, limited.Body)
	}
	if limited.Headers.Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header on TOTP verify 429")
	}
}

func TestTOTP_SessionRevocationOnChangeWhenConfigured(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	defer setBoolConfig(t, "auth.totp.revokeSessionsOnChange", false, "Revoke all active sessions after TOTP changes")

	email := "totp-revoke-sessions@example.com"
	testutil.RegisterUser(t, suite.Client, email, "TOTP Revoke Sessions")

	secondSession := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, secondSession, email, testutil.TestPassword())
	testutil.AssertSuccess(t, secondSession.GET("/api/v1/auth/session"))

	setBoolConfig(t, "auth.totp.revokeSessionsOnChange", true, "Revoke all active sessions after TOTP changes")

	setupResp := suite.Client.POST("/api/v1/me/totp/setup", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	testutil.AssertSuccess(t, setupResp)
	secret := testutil.ParseDataString(t, setupResp, "secret")
	setupCode, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate setup TOTP code: %v", err)
	}
	enableResp := suite.Client.POST("/api/v1/me/totp/enable", fmt.Sprintf(`{"code":"%s"}`, setupCode))
	testutil.AssertSuccess(t, enableResp)

	testutil.AssertStatus(t, suite.Client.GET("/api/v1/auth/session"), http.StatusUnauthorized)
	testutil.AssertStatus(t, secondSession.GET("/api/v1/auth/session"), http.StatusUnauthorized)
	revoked := scalarInt(t, "SELECT COUNT(*) FROM auth_sessions WHERE user_id = (SELECT id FROM users WHERE email = $1) AND revoked_at IS NOT NULL", email)
	if revoked < 2 {
		t.Fatalf("expected both active sessions to be revoked after TOTP enable, got %d", revoked)
	}
}

func TestTOTP_AuditEventsAreRecorded(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "totp-audit@example.com"
	user, _ := testutil.RegisterUser(t, suite.Client, email, "TOTP Audit")
	userID := user["id"].(string)

	setupResp := suite.Client.POST("/api/v1/me/totp/setup", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	testutil.AssertSuccess(t, setupResp)
	secret := testutil.ParseDataString(t, setupResp, "secret")
	setupCode, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate setup TOTP code: %v", err)
	}
	enableResp := suite.Client.POST("/api/v1/me/totp/enable", fmt.Sprintf(`{"code":"%s"}`, setupCode))
	testutil.AssertSuccess(t, enableResp)

	verifyClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	loginResp := verifyClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	token := loginResp.JSONData()["totp_token"].(string)
	failedResp := verifyClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"000000"}`, token))
	if failedResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected failed TOTP verify to be 403, got %d body=%s", failedResp.StatusCode, failedResp.Body)
	}

	loginResp = verifyClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	token = loginResp.JSONData()["totp_token"].(string)
	verifyCode, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate verify TOTP code: %v", err)
	}
	verifyResp := verifyClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"%s"}`, token, verifyCode))
	testutil.AssertSuccess(t, verifyResp)

	regenerateResp := suite.Client.POST("/api/v1/me/totp/backup-codes/regenerate", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	testutil.AssertSuccess(t, regenerateResp)
	backupCodes := regenerateResp.JSONData()["backup_codes"].([]any)
	backupCode := backupCodes[0].(string)

	backupClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	loginResp = backupClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	testutil.AssertSuccess(t, loginResp)
	token = loginResp.JSONData()["totp_token"].(string)
	backupResp := backupClient.POST("/api/v1/auth/verify-totp", fmt.Sprintf(`{"totp_token":"%s","code":"%s"}`, token, backupCode))
	testutil.AssertSuccess(t, backupResp)

	disableResp := suite.Client.POST("/api/v1/me/totp/disable", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	testutil.AssertSuccess(t, disableResp)

	for _, action := range []string{
		"auth.totp.setup_initiated",
		"auth.totp.enabled",
		"auth.totp.failed",
		"auth.totp.verified",
		"auth.totp.backup_codes_regenerated",
		"auth.totp.backup_code_used",
		"auth.totp.disabled",
	} {
		count := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE user_id = $1 AND action = $2", userID, action)
		if count < 1 {
			t.Fatalf("expected audit action %s to be recorded", action)
		}
	}
}

func TestTOTP_MissingFeatureFlagDisablesTOTP(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	defer setTOTPFeatureFlag(t, true)

	email := "totp-missing-flag@example.com"
	testutil.RegisterUser(t, suite.Client, email, "TOTP Missing Flag")
	if err := service.Config().Delete(context.Background(), "auth.totp.enabled"); err != nil {
		t.Fatalf("delete auth.totp.enabled: %v", err)
	}

	resp := suite.Client.POST("/api/v1/me/totp/setup", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected missing TOTP feature flag to disable setup with 403, got %d body=%s", resp.StatusCode, resp.Body)
	}
}

func createUserSecurityReadAPIKey(t *testing.T) string {
	t.Helper()
	resp := suite.Client.POST("/api/v1/tenants", `{"name":"TOTP API Key Org","slug":"totp-api-key-org"}`)
	testutil.AssertSuccess(t, resp)
	tenantID := testutil.ParseDataString(t, resp, "tenant.id")

	resp = suite.Client.DoWithHeaders(http.MethodPost, "/api/v1/tenants/"+tenantID+"/api-keys",
		`{"name":"TOTP Read Key","scopes":["user:security:read","tenant:read"]}`,
		map[string]string{"X-Tenant-ID": tenantID})
	testutil.AssertSuccess(t, resp)
	return resp.JSONData()["raw_key"].(string)
}

func setupAndEnableTOTP(t *testing.T, email string) (string, string, []string) {
	t.Helper()
	user, _ := testutil.RegisterUser(t, suite.Client, email, "TOTP User")
	userID := user["id"].(string)
	setupResp := suite.Client.POST("/api/v1/me/totp/setup", fmt.Sprintf(`{"password":"%s"}`, testutil.TestPassword()))
	testutil.AssertSuccess(t, setupResp)
	secret := testutil.ParseDataString(t, setupResp, "secret")
	setupCode, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate setup TOTP code: %v", err)
	}
	enableResp := suite.Client.POST("/api/v1/me/totp/enable", fmt.Sprintf(`{"code":"%s"}`, setupCode))
	testutil.AssertSuccess(t, enableResp)
	items := enableResp.JSONData()["backup_codes"].([]any)
	codes := make([]string, len(items))
	for i, item := range items {
		codes[i] = item.(string)
	}
	return userID, secret, codes
}

func runConcurrentPosts(clients []*testutil.TestClient, path, body string) []int {
	bodies := make([]string, len(clients))
	for i := range bodies {
		bodies[i] = body
	}
	return runConcurrentClientPosts(clients, path, bodies)
}

func runConcurrentClientPosts(clients []*testutil.TestClient, path string, bodies []string) []int {
	var wg sync.WaitGroup
	statuses := make([]int, len(clients))
	start := make(chan struct{})
	for i, client := range clients {
		i, client := i, client
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			resp := client.POST(path, bodies[i])
			statuses[i] = resp.StatusCode
		}()
	}
	close(start)
	wg.Wait()
	return statuses
}

func assertExactlyOneStatus(t *testing.T, statuses []int, expected int) {
	t.Helper()
	count := 0
	for _, status := range statuses {
		if status == expected {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one status %d, got statuses=%v", expected, statuses)
	}
}

func scalarString(t *testing.T, sql string, args ...any) string {
	t.Helper()
	value, err := g.DB().GetValue(context.Background(), sql, args...)
	if err != nil {
		t.Fatalf("query scalar string: %v", err)
	}
	return value.String()
}

func scalarInt(t *testing.T, sql string, args ...any) int {
	t.Helper()
	value, err := g.DB().GetValue(context.Background(), sql, args...)
	if err != nil {
		t.Fatalf("query scalar int: %v", err)
	}
	return value.Int()
}

func setTOTPFeatureFlag(t *testing.T, enabled bool) {
	t.Helper()
	setBoolConfig(t, "auth.totp.enabled", enabled, "TOTP 2FA global feature flag")
}

func setBoolConfig(t *testing.T, key string, enabled bool, description string) {
	t.Helper()
	if err := service.Config().Set(context.Background(), &service.ConfigSetParams{
		Key:         key,
		Value:       fmt.Sprintf("%t", enabled),
		ValueType:   "bool",
		Description: description,
	}); err != nil {
		t.Fatalf("set %s: %v", key, err)
	}
}
