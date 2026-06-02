//go:build integration

package integration_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/internal/testutil"
)

func TestEmailVerification_RegisterVerifyResendExpiredReplayAndAudit(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "email-verify-chain@example.com"
	user, _ := testutil.RegisterUser(t, suite.Client, email, "Email Verify Chain")
	userID := user["id"].(string)

	waitForCountAtLeast(t, "SELECT COUNT(*) FROM email_verification_tokens WHERE user_id = $1", 1, userID)
	assertEmailVerified(t, email, false)

	// Immediate resend is rate-limited by the just-created registration token.
	resp := suite.Client.POST("/api/v1/auth/resend-verification", fmt.Sprintf(`{"email":"%s"}`, email))
	testutil.AssertStatus(t, resp, http.StatusTooManyRequests)

	// Once the cooldown window is older, resend creates a fresh token and remains enumeration-safe.
	mustExec(t, "UPDATE email_verification_tokens SET created_at = NOW() - INTERVAL '2 minutes' WHERE email = $1", email)
	resp = suite.Client.POST("/api/v1/auth/resend-verification", fmt.Sprintf(`{"email":"%s"}`, email))
	testutil.AssertSuccess(t, resp)
	waitForCountAtLeast(t, "SELECT COUNT(*) FROM email_verification_tokens WHERE user_id = $1", 2, userID)

	resp = suite.Client.POST("/api/v1/auth/resend-verification", `{"email":"missing-email-verify@example.com"}`)
	testutil.AssertSuccess(t, resp)

	validToken := strings.Repeat("a", 43)
	insertEmailVerificationToken(t, userID, email, validToken, time.Now().Add(time.Hour))
	resp = suite.Client.POST("/api/v1/auth/verify-email", fmt.Sprintf(`{"token":"%s"}`, validToken))
	testutil.AssertSuccess(t, resp)
	assertEmailVerified(t, email, true)
	if used := scalarInt(t, "SELECT COUNT(*) FROM email_verification_tokens WHERE token_hash = $1 AND used_at IS NOT NULL", emailVerificationHash(t, validToken)); used != 1 {
		t.Fatalf("expected verification token to be marked used once, got %d", used)
	}
	if audits := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE user_id = $1 AND action = 'auth.email.verified'", userID); audits != 1 {
		t.Fatalf("expected one auth.email.verified audit, got %d", audits)
	}

	resp = suite.Client.POST("/api/v1/auth/verify-email", fmt.Sprintf(`{"token":"%s"}`, validToken))
	testutil.AssertStatus(t, resp, http.StatusNotFound)

	resp = suite.Client.POST("/api/v1/auth/verify-email", `{"token":"not-a-real-token"}`)
	testutil.AssertStatus(t, resp, http.StatusNotFound)

	expiredEmail := "email-expired-chain@example.com"
	expiredUser := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), expiredEmail, "Email Expired Chain")
	expiredUserID := expiredUser["id"].(string)
	expiredToken := strings.Repeat("b", 43)
	insertEmailVerificationToken(t, expiredUserID, expiredEmail, expiredToken, time.Now().Add(-time.Minute))
	resp = suite.Client.POST("/api/v1/auth/verify-email", fmt.Sprintf(`{"token":"%s"}`, expiredToken))
	testutil.AssertStatus(t, resp, http.StatusGone)
	assertEmailVerified(t, expiredEmail, false)
}

func TestPasswordReset_ForgotResetReplayExpiredSessionRevokeAndAudit(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	email := "password-reset-chain@example.com"
	user, _ := testutil.RegisterUser(t, suite.Client, email, "Password Reset Chain")
	userID := user["id"].(string)

	resp := suite.Client.POST("/api/v1/auth/forgot-password", `{"email":"missing-password-reset@example.com"}`)
	testutil.AssertSuccess(t, resp)
	if tokens := scalarInt(t, "SELECT COUNT(*) FROM password_reset_tokens"); tokens != 0 {
		t.Fatalf("expected missing-user forgot password to create no token, got %d", tokens)
	}
	if audits := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'auth.password.forgot' AND metadata::text LIKE '%email_not_found%'"); audits != 1 {
		t.Fatalf("expected one enumeration-safe forgot audit for missing email, got %d", audits)
	}

	resp = suite.Client.POST("/api/v1/auth/forgot-password", fmt.Sprintf(`{"email":"%s"}`, email))
	testutil.AssertSuccess(t, resp)
	if tokens := scalarInt(t, "SELECT COUNT(*) FROM password_reset_tokens WHERE user_id = $1 AND used_at IS NULL", userID); tokens != 1 {
		t.Fatalf("expected forgot password to create one unused token, got %d", tokens)
	}
	if audits := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE user_id = $1 AND action = 'auth.password.forgot'", userID); audits != 1 {
		t.Fatalf("expected one auth.password.forgot audit, got %d", audits)
	}

	secondSession := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, secondSession, email, testutil.TestPassword())
	testutil.AssertSuccess(t, suite.Client.GET("/api/v1/auth/session"))
	testutil.AssertSuccess(t, secondSession.GET("/api/v1/auth/session"))

	validToken := strings.Repeat("1", 64)
	newPassword := "NewResetPassword12345!"
	insertPasswordResetToken(t, userID, validToken, time.Now().Add(time.Hour))

	resetClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp = resetClient.POST("/api/v1/auth/reset-password", fmt.Sprintf(`{"token":"%s","new_password":"%s"}`, validToken, newPassword))
	testutil.AssertSuccess(t, resp)
	if used := scalarInt(t, "SELECT COUNT(*) FROM password_reset_tokens WHERE user_id = $1 AND used_at IS NOT NULL", userID); used < 2 {
		t.Fatalf("expected reset to mark current and prior unused tokens used, got %d", used)
	}
	if audits := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE user_id = $1 AND action = 'auth.password.reset'", userID); audits != 1 {
		t.Fatalf("expected one auth.password.reset audit, got %d", audits)
	}
	if revoked := scalarInt(t, "SELECT COUNT(*) FROM auth_sessions WHERE user_id = $1 AND revoked_at IS NOT NULL AND revoke_reason = 'password_reset'", userID); revoked < 2 {
		t.Fatalf("expected existing sessions to be revoked by password reset, got %d", revoked)
	}
	testutil.AssertStatus(t, suite.Client.GET("/api/v1/auth/session"), http.StatusUnauthorized)
	testutil.AssertStatus(t, secondSession.GET("/api/v1/auth/session"), http.StatusUnauthorized)

	oldPasswordClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp = oldPasswordClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, testutil.TestPassword()))
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected old password login to fail after reset, got %d body=%s", resp.StatusCode, resp.Body)
	}
	newPasswordClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	resp = newPasswordClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, newPassword))
	testutil.AssertSuccess(t, resp)

	resp = resetClient.POST("/api/v1/auth/reset-password", fmt.Sprintf(`{"token":"%s","new_password":"%s"}`, validToken, "ReplayResetPassword12345!"))
	testutil.AssertStatus(t, resp, http.StatusForbidden)
	if failed := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'auth.password.reset.failed' AND metadata::text LIKE '%invalid_token%'"); failed < 1 {
		t.Fatalf("expected replay to write invalid_token reset failure audit, got %d", failed)
	}

	expiredToken := strings.Repeat("2", 64)
	insertPasswordResetToken(t, userID, expiredToken, time.Now().Add(-time.Minute))
	resp = resetClient.POST("/api/v1/auth/reset-password", fmt.Sprintf(`{"token":"%s","new_password":"%s"}`, expiredToken, "ExpiredResetPassword12345!"))
	testutil.AssertStatus(t, resp, http.StatusForbidden)
	if failed := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'auth.password.reset.failed' AND metadata::text LIKE '%expired_token%'"); failed < 1 {
		t.Fatalf("expected expired token to write expired_token reset failure audit, got %d", failed)
	}
}

func TestAccountLockout_ConfigAuditUnlockAndSuccessCleanup(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	setNumberConfig(t, "auth.lockout.max_attempts", "3", "Lockout threshold for integration test")
	defer setNumberConfig(t, "auth.lockout.max_attempts", "5", "Default lockout threshold")
	setNumberConfig(t, "auth.lockout.window_minutes", "15", "Lockout window for integration test")
	defer setNumberConfig(t, "auth.lockout.window_minutes", "15", "Default lockout window")
	setNumberConfig(t, "auth.lockout.duration_minutes", "15", "Lockout duration for integration test")
	defer setNumberConfig(t, "auth.lockout.duration_minutes", "15", "Default lockout duration")

	adminUser, _ := testutil.RegisterUser(t, suite.Client, "lockout-admin-chain@example.com", "Lockout Admin Chain")
	if adminUser["id"] == "" {
		t.Fatal("expected admin user id")
	}
	targetEmail := "lockout-target-chain@example.com"
	targetUser := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), targetEmail, "Lockout Target Chain")
	targetID := targetUser["id"].(string)

	targetClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	for i := 0; i < 3; i++ {
		resp := targetClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"WrongPassword12345!"}`, targetEmail))
		if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
			t.Fatalf("attempt %d: expected 401/403 wrong-password failure, got %d body=%s", i+1, resp.StatusCode, resp.Body)
		}
	}
	lockedUntil := scalarString(t, "SELECT COALESCE(locked_until::text, '') FROM auth_login_attempts WHERE login_key = $1", targetEmail)
	if lockedUntil == "" {
		t.Fatal("expected locked_until to be set after configured threshold")
	}
	if audits := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'auth.account.locked' AND metadata::text LIKE $1", "%"+targetEmail+"%"); audits != 1 {
		t.Fatalf("expected one auth.account.locked audit, got %d", audits)
	}

	resp := targetClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, targetEmail, testutil.TestPassword()))
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected locked account to reject correct password, got %d body=%s", resp.StatusCode, resp.Body)
	}

	resp = suite.Client.POST(fmt.Sprintf("/api/v1/admin/users/%s/unlock", targetID), "")
	testutil.AssertSuccess(t, resp)
	if audits := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'auth.account.unlocked' AND resource_id = $1", targetID); audits != 1 {
		t.Fatalf("expected one auth.account.unlocked audit, got %d", audits)
	}

	resp = targetClient.POST("/api/v1/auth/login", fmt.Sprintf(`{"email":"%s","password":"%s"}`, targetEmail, testutil.TestPassword()))
	testutil.AssertSuccess(t, resp)
	failedCount := scalarInt(t, "SELECT failed_count FROM auth_login_attempts WHERE login_key = $1", targetEmail)
	lockedUntil = scalarString(t, "SELECT COALESCE(locked_until::text, '') FROM auth_login_attempts WHERE login_key = $1", targetEmail)
	lastSuccessAt := scalarString(t, "SELECT COALESCE(last_success_at::text, '') FROM auth_login_attempts WHERE login_key = $1", targetEmail)
	if failedCount != 0 || lockedUntil != "" || lastSuccessAt == "" {
		t.Fatalf("expected successful login to clear lockout state, got failed_count=%d locked_until=%q last_success_at=%q", failedCount, lockedUntil, lastSuccessAt)
	}
}

func insertEmailVerificationToken(t *testing.T, userID, email, token string, expiresAt time.Time) {
	t.Helper()
	mustExec(t,
		"INSERT INTO email_verification_tokens(user_id, email, token_hash, expires_at, created_at) VALUES($1, $2, $3, $4, NOW() - INTERVAL '2 minutes')",
		userID, strings.ToLower(strings.TrimSpace(email)), emailVerificationHash(t, token), expiresAt,
	)
}

func emailVerificationHash(t *testing.T, token string) string {
	t.Helper()
	secret := strings.TrimSpace(service.Config().GetString(context.Background(), "auth.session.secret", ""))
	if secret == "" {
		t.Fatal("auth.session.secret is required for email verification token hashing")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func insertPasswordResetToken(t *testing.T, userID, token string, expiresAt time.Time) {
	t.Helper()
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	mustExec(t,
		"INSERT INTO password_reset_tokens(user_id, token_hash, expires_at, requested_ip) VALUES($1, $2, $3, '127.0.0.1')",
		userID, hex.EncodeToString(sum[:]), expiresAt,
	)
}

func assertEmailVerified(t *testing.T, email string, expected bool) {
	t.Helper()
	got := scalarString(t, "SELECT COALESCE(email_verified::text, '') FROM user_identities WHERE provider = 'password' AND lower(email) = lower($1)", email)
	if got != fmt.Sprintf("%t", expected) {
		t.Fatalf("expected email_verified=%t for %s, got %q", expected, email, got)
	}
}

func waitForCountAtLeast(t *testing.T, sql string, min int, args ...any) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	var got int
	for {
		got = scalarInt(t, sql, args...)
		if got >= min {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for count >= %d; got %d for query %q", min, got, sql)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func mustExec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := g.DB().Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec failed: %v; sql=%s", err, sql)
	}
}

func setNumberConfig(t *testing.T, key, value, description string) {
	t.Helper()
	if err := service.Config().Set(context.Background(), &service.ConfigSetParams{
		Key:         key,
		Value:       value,
		ValueType:   "number",
		Description: description,
	}); err != nil {
		t.Fatalf("set %s: %v", key, err)
	}
}
