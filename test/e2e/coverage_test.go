//go:build e2e

package e2e_test

import "testing"

type coverageMapping struct {
	LegacyScenario string
	GoTarget       string
	Status         string
}

func TestLegacyToGoCoverageMatrix(t *testing.T) {
	mappings := []coverageMapping{
		{"setup_wizard", "test/e2e/setup/setup_wizard_test.go", "done"},
		{"session_management", "test/e2e/auth/session_test.go", "done"},
		{"password_reset", "test/e2e/auth/security_chain_test.go", "done"},
		{"email_verification", "test/e2e/auth/security_chain_test.go", "done"},
		{"account_lockout", "test/e2e/auth/security_chain_test.go", "done"},
		{"totp_2fa", "test/e2e/auth/totp_test.go", "done"},
		{"oauth_login", "test/e2e/auth/oauth_test.go", "done"},
		{"system_config", "test/e2e/setup/system_config_test.go", "done"},
		{"first_user_platform_admin_bootstrap", "test/e2e/setup/first_user_bootstrap_test.go", "done"},
		{"personal_api_keys", "test/e2e/tenant/personal_api_keys_test.go", "done"},
		{"menu_permissions", "test/e2e/tenant/menu_permissions_test.go", "done"},
		{"multitenancy_http", "test/e2e/tenant/multitenancy_test.go", "done"},
		{"saas_multitenancy_management", "test/e2e/tenant/saas_management_test.go", "done"},
		{"default_tenant_lazy_runtime", "test/e2e/tenant/default_tenant_test.go", "covered-by-tenant"},
		{"frontend_spa", "test/e2e/frontend/spa_smoke_test.go", "done"},
		{"full_flow", "test/e2e/auth+setup+tenant+frontend", "split-by-domain"},
		{"test_multitenancy", "test/e2e/tenant/multitenancy_test.go", "done"},
	}
	seen := map[string]bool{}
	for _, mapping := range mappings {
		if mapping.LegacyScenario == "" || mapping.GoTarget == "" || mapping.Status == "" {
			t.Fatalf("incomplete mapping: %+v", mapping)
		}
		if seen[mapping.LegacyScenario] {
			t.Fatalf("duplicate legacy mapping for %s", mapping.LegacyScenario)
		}
		seen[mapping.LegacyScenario] = true
	}
	if len(mappings) != 17 {
		t.Fatalf("expected 17 legacy mappings, got %d", len(mappings))
	}
}
