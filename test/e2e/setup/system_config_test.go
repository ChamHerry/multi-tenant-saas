//go:build e2e

package setup_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/internal/testutil"
)

func TestSystemConfigAdmin_CRUDValidationSecretAuditCache(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	ctx := context.Background()
	boolKey := "e2e.systemConfig.bool"
	secretKey := "e2e.systemConfig.secret"
	jsonKey := "e2e.systemConfig.json"
	cleanupSystemConfigKeys(t, boolKey, secretKey, jsonKey)
	defer cleanupSystemConfigKeys(t, boolKey, secretKey, jsonKey)

	testutil.RegisterUser(t, suite.Client, "sysconfig-admin@example.com", "System Config Admin")

	seedSecret := suite.Client.GET("/api/v1/admin/system-config/auth.session.secret")
	testutil.AssertSuccess(t, seedSecret)
	rawSessionSecret := scalarString(t, "SELECT value FROM system_config WHERE key=$1", "auth.session.secret")
	if rawSessionSecret == "" {
		t.Fatal("expected seeded auth.session.secret value")
	}
	if strings.Contains(seedSecret.Body, rawSessionSecret) {
		t.Fatalf("seeded sensitive key leaked in API response: %s", seedSecret.Body)
	}
	seedSecretConfig := seedSecret.JSONData()["config"].(map[string]any)
	if seedSecretConfig["value"] != "" || seedSecretConfig["value_type"] != "secret" || seedSecretConfig["is_secret"] != true || seedSecretConfig["masked_value"] != "********" {
		t.Fatalf("expected seeded sensitive key to be masked as secret, got %v", seedSecretConfig)
	}
	stringSecretWrite := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/auth.session.secret", `{"value":"unsafe","value_provided":true,"value_type":"string","description":"bad"}`)
	testutil.AssertStatus(t, stringSecretWrite, http.StatusBadRequest)
	weakSessionSecret := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/auth.session.secret", `{"value":"change-me","value_provided":true,"value_type":"secret","description":"bad"}`)
	testutil.AssertStatus(t, weakSessionSecret, http.StatusBadRequest)
	removedDevHeader := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/auth.devHeader.enabled", `{"value":"true","value_provided":true,"value_type":"bool","description":"removed"}`)
	testutil.AssertStatus(t, removedDevHeader, http.StatusBadRequest)

	create := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/"+boolKey, `{"value":"true","value_provided":true,"value_type":"bool","description":"Integration bool flag"}`)
	testutil.AssertSuccess(t, create)
	config := create.JSONData()["config"].(map[string]any)
	if config["key"] != boolKey || config["value"] != "true" || config["value_type"] != "bool" || config["category"] != "e2e" {
		t.Fatalf("unexpected bool config response: %v", config)
	}
	if !service.Config().GetBool(ctx, boolKey, false) {
		t.Fatal("runtime Config().GetBool did not read the updated bool value")
	}
	firstUpdatedAt := scalarString(t, "SELECT updated_at::text FROM system_config WHERE key=$1", boolKey)
	time.Sleep(10 * time.Millisecond)

	update := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/"+boolKey, `{"value":"false","value_provided":true,"value_type":"bool","description":"Integration bool flag updated"}`)
	testutil.AssertSuccess(t, update)
	if service.Config().GetBool(ctx, boolKey, true) {
		t.Fatal("runtime Config().GetBool did not observe cache invalidation after update")
	}
	secondUpdatedAt := scalarString(t, "SELECT updated_at::text FROM system_config WHERE key=$1", boolKey)
	if firstUpdatedAt == secondUpdatedAt {
		t.Fatalf("expected updated_at to change, got %q", secondUpdatedAt)
	}

	list := suite.Client.GET("/api/v1/admin/system-config?query=e2e.systemConfig&category=e2e")
	testutil.AssertSuccess(t, list)
	items := list.JSONData()["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected e2e config in list response: %s", list.Body)
	}

	invalidBool := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/"+boolKey, `{"value":"enabled","value_provided":true,"value_type":"bool","description":"bad"}`)
	testutil.AssertStatus(t, invalidBool, http.StatusBadRequest)
	invalidNumber := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/e2e.systemConfig.number", `{"value":"not-number","value_provided":true,"value_type":"number","description":"bad"}`)
	testutil.AssertStatus(t, invalidNumber, http.StatusBadRequest)
	invalidJSON := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/"+jsonKey, `{"value":"{bad}","value_provided":true,"value_type":"json","description":"bad"}`)
	testutil.AssertStatus(t, invalidJSON, http.StatusBadRequest)

	secretPlaintext := "super-secret-never-leak"
	secretResp := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/"+secretKey, fmt.Sprintf(`{"value":"%s","value_provided":true,"value_type":"secret","description":"Integration secret"}`, secretPlaintext))
	testutil.AssertSuccess(t, secretResp)
	if strings.Contains(secretResp.Body, secretPlaintext) {
		t.Fatalf("secret plaintext leaked in API response: %s", secretResp.Body)
	}
	secretConfig := secretResp.JSONData()["config"].(map[string]any)
	if secretConfig["value"] != "" || secretConfig["masked_value"] != "********" || secretConfig["has_value"] != true || secretConfig["is_secret"] != true {
		t.Fatalf("unexpected secret response: %v", secretConfig)
	}
	rawSecret := scalarString(t, "SELECT value FROM system_config WHERE key=$1", secretKey)
	if rawSecret == secretPlaintext || rawSecret == "" {
		t.Fatalf("expected encrypted raw secret, got %q", rawSecret)
	}
	if got := service.Config().GetString(ctx, secretKey, ""); got != secretPlaintext {
		t.Fatalf("runtime Config().GetString secret=%q want plaintext", got)
	}

	preserveSecret := suite.Client.Do(http.MethodPut, "/api/v1/admin/system-config/"+secretKey, `{"value":"","value_provided":false,"value_type":"secret","description":"Description only"}`)
	testutil.AssertSuccess(t, preserveSecret)
	if got := service.Config().GetString(ctx, secretKey, ""); got != secretPlaintext {
		t.Fatalf("secret metadata update changed secret value to %q", got)
	}

	audit := suite.Client.GET("/api/v1/admin/audit-logs?action=system_config.updated&resource_type=system_config&limit=20")
	testutil.AssertSuccess(t, audit)
	if strings.Contains(audit.Body, secretPlaintext) {
		t.Fatalf("secret plaintext leaked in audit response: %s", audit.Body)
	}
	logs := audit.JSONData()["logs"].([]any)
	foundSecretAudit := false
	for _, raw := range logs {
		entry := raw.(map[string]any)
		if entry["resource_id"] == secretKey {
			foundSecretAudit = true
		}
	}
	if !foundSecretAudit {
		t.Fatalf("expected audit log for %s, got %s", secretKey, audit.Body)
	}

	deleteProtected := suite.Client.DELETE("/api/v1/admin/system-config/auth.session.secret")
	testutil.AssertStatus(t, deleteProtected, http.StatusBadRequest)
	deleteResp := suite.Client.DELETE("/api/v1/admin/system-config/" + boolKey)
	testutil.AssertSuccess(t, deleteResp)
	if got := service.Config().GetString(ctx, boolKey, "missing"); got != "missing" {
		t.Fatalf("deleted key still readable from runtime cache: %q", got)
	}
}

func TestSystemConfigAdmin_Permissions(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)

	key := "e2e.systemConfig.permission"
	cleanupSystemConfigKeys(t, key)
	defer cleanupSystemConfigKeys(t, key)

	testutil.RegisterUser(t, suite.Client, "sysconfig-owner@example.com", "System Config Owner")
	supportUser := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "sysconfig-support@example.com", "System Config Support")
	auditorUser := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "sysconfig-auditor@example.com", "System Config Auditor")
	nonAdminUser := testutil.RegisterAdditionalUser(t, suite.Client.BaseURL(), "sysconfig-user@example.com", "System Config User")

	supportID := supportUser["id"].(string)
	auditorID := auditorUser["id"].(string)
	testutil.AssertSuccess(t, suite.Client.POST("/api/v1/admin/platform-admins", fmt.Sprintf(`{"user_id":"%s","role":"support"}`, supportID)))
	testutil.AssertSuccess(t, suite.Client.POST("/api/v1/admin/platform-admins", fmt.Sprintf(`{"user_id":"%s","role":"auditor"}`, auditorID)))

	supportClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, supportClient, "sysconfig-support@example.com", testutil.TestPassword())
	testutil.AssertSuccess(t, supportClient.GET("/api/v1/admin/system-config?limit=1"))
	supportWrite := supportClient.Do(http.MethodPut, "/api/v1/admin/system-config/"+key, `{"value":"true","value_provided":true,"value_type":"bool","description":"forbidden"}`)
	testutil.AssertStatus(t, supportWrite, http.StatusForbidden)

	auditorClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	testutil.LoginUser(t, auditorClient, "sysconfig-auditor@example.com", testutil.TestPassword())
	testutil.AssertSuccess(t, auditorClient.GET("/api/v1/admin/system-config?limit=1"))
	auditorDelete := auditorClient.DELETE("/api/v1/admin/system-config/auth.password.enabled")
	testutil.AssertStatus(t, auditorDelete, http.StatusForbidden)

	nonAdminClient := testutil.NewTestClient(t, suite.Client.BaseURL())
	_ = nonAdminUser
	testutil.LoginUser(t, nonAdminClient, "sysconfig-user@example.com", testutil.TestPassword())
	nonAdminRead := nonAdminClient.GET("/api/v1/admin/system-config?limit=1")
	if nonAdminRead.StatusCode != http.StatusForbidden && nonAdminRead.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected non-admin read forbidden, got %d body %s", nonAdminRead.StatusCode, nonAdminRead.Body)
	}
}

func cleanupSystemConfigKeys(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if _, err := g.DB().Exec(context.Background(), "DELETE FROM system_config WHERE key=$1", key); err != nil {
			t.Fatalf("cleanup system config %s: %v", key, err)
		}
		_ = service.Config().Delete(context.Background(), key)
	}
}
