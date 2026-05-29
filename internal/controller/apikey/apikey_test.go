package apikey

import (
	"testing"

	"repomind-temp/internal/service"
)

func TestTenantGrantScopes(t *testing.T) {
	allowed := []service.Permission{service.PermissionTenantRead, service.PermissionMemberRead}
	grantScopes, err := tenantGrantScopes([]string{"user:read", "tenant:read", "member:read"}, allowed)
	if err != nil {
		t.Fatalf("tenantGrantScopes returned error: %v", err)
	}
	if len(grantScopes) != 2 || grantScopes[0] != "tenant:read" || grantScopes[1] != "member:read" {
		t.Fatalf("tenantGrantScopes=%v want [tenant:read member:read]", grantScopes)
	}
}

func TestTenantGrantScopesRejectsEscalation(t *testing.T) {
	allowed := []service.Permission{service.PermissionTenantRead}
	if _, err := tenantGrantScopes([]string{"tenant:manage"}, allowed); err == nil {
		t.Fatal("tenantGrantScopes should reject scopes outside current tenant permissions")
	}
}

func TestTenantGrantScopesRejectsWildcard(t *testing.T) {
	allowed := []service.Permission{service.PermissionTenantRead}
	if _, err := tenantGrantScopes([]string{"*"}, allowed); err == nil {
		t.Fatal("tenantGrantScopes should reject wildcard scopes for tenant-scoped api keys")
	}
}
