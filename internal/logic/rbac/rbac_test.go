package rbac

import (
	"context"
	"testing"

	"repomind-temp/internal/service"
)

func TestCanByRole(t *testing.T) {
	s := &sRBAC{}
	ctx := context.Background()
	cases := []struct {
		role string
		perm service.Permission
		want bool
	}{
		{"owner", service.PermissionTenantManage, true},
		{"admin", service.PermissionMemberManage, true},
		{"admin", service.PermissionTenantManage, false},
		{"member", service.PermissionMemberManage, false},
		{"viewer", service.PermissionTenantRead, true},
		{"viewer", service.PermissionMemberRead, false},
	}
	for _, tt := range cases {
		got := s.Can(ctx, &service.TenantContext{Role: tt.role}, tt.perm)
		if got != tt.want {
			t.Fatalf("Can(%s,%s)=%v want %v", tt.role, tt.perm, got, tt.want)
		}
	}
}

func TestAPIKeyScopesIntersectRole(t *testing.T) {
	s := &sRBAC{}
	ctx := context.Background()
	tc := &service.TenantContext{Role: "owner", AuthType: "api_key", Scopes: []string{string(service.PermissionTenantRead)}}
	if !s.Can(ctx, tc, service.PermissionTenantRead) {
		t.Fatal("api key with tenant:read scope should allow tenant:read")
	}
	if s.Can(ctx, tc, service.PermissionTenantManage) {
		t.Fatal("api key without tenant:manage scope should not allow tenant:manage")
	}
}

func TestPermissionsForRoleReturnsCopyInStableOrder(t *testing.T) {
	s := &sRBAC{}
	permissions := s.PermissionsForRole("viewer")
	want := []service.Permission{service.PermissionTenantRead}
	if len(permissions) != len(want) {
		t.Fatalf("len=%d want %d", len(permissions), len(want))
	}
	for i := range want {
		if permissions[i] != want[i] {
			t.Fatalf("permissions[%d]=%s want %s", i, permissions[i], want[i])
		}
	}
	permissions[0] = service.PermissionTenantManage
	if got := s.PermissionsForRole("viewer")[0]; got != service.PermissionTenantRead {
		t.Fatalf("PermissionsForRole should return a copy, got first permission %s", got)
	}
}

func TestPermissionsForContextIntersectsAPIKeyScopes(t *testing.T) {
	s := &sRBAC{}
	ctx := context.Background()
	tc := &service.TenantContext{Role: "owner", AuthType: "api_key", Scopes: []string{string(service.PermissionTenantRead)}}
	permissions := s.PermissionsForContext(ctx, tc)
	if len(permissions) != 1 || permissions[0] != service.PermissionTenantRead {
		t.Fatalf("PermissionsForContext=%v want [tenant:read]", permissions)
	}
}

func TestKnownPermission(t *testing.T) {
	for _, scope := range []string{"*", "tenant:read", "member:manage", "api_key:manage"} {
		if !KnownPermission(scope) {
			t.Fatalf("KnownPermission(%q)=false", scope)
		}
	}
	if KnownPermission("unknown") {
		t.Fatal("KnownPermission(unknown)=true")
	}
}
