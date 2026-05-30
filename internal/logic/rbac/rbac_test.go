package rbac

import (
	"context"
	"testing"

	"multi-tenant-saas/internal/service"
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
	for _, scope := range []string{"*", "tenant:read", "member:manage"} {
		if !KnownPermission(scope) {
			t.Fatalf("KnownPermission(%q)=false", scope)
		}
	}
	if KnownPermission("api_key:manage") {
		t.Fatal("KnownPermission(api_key:manage)=true")
	}
	if KnownPermission("tenant:billing:read") || KnownPermission("tenant:billing:manage") {
		t.Fatal("KnownPermission should not include removed tenant billing scopes")
	}
	if KnownPermission("unknown") {
		t.Fatal("KnownPermission(unknown)=true")
	}
}

func TestAPIKeyUserScopes(t *testing.T) {
	s := &sRBAC{}
	ctx := service.WithAuthIdentity(context.Background(), &service.AuthIdentity{Type: "api_key", Scopes: []string{string(service.PermissionUserRead)}})
	if err := s.RequireAuthScope(ctx, service.PermissionUserRead); err != nil {
		t.Fatalf("RequireAuthScope(user:read) error = %v", err)
	}
	if err := s.RequireAuthScope(ctx, service.PermissionUserTenantRead); err == nil {
		t.Fatal("RequireAuthScope(user:tenant:read) expected error")
	}
	if !KnownPermission("user:read") || !KnownPermission("user:tenant:read") || !KnownPermission("api_key:self_manage") {
		t.Fatal("KnownPermission should include user-level api key scopes")
	}
}

func TestSessionBypassesAuthScope(t *testing.T) {
	s := &sRBAC{}
	ctx := service.WithAuthIdentity(context.Background(), &service.AuthIdentity{Type: "session"})
	if err := s.RequireAuthScope(ctx, service.PermissionUserRead); err != nil {
		t.Fatalf("session RequireAuthScope error = %v", err)
	}
}
