package platformadmin

import (
	"testing"

	"multi-tenant-saas/internal/service"
)

func TestPlatformRolePermissions(t *testing.T) {
	if len(rolePermissions["super_admin"]) <= len(rolePermissions["support"]) {
		t.Fatal("super_admin should have more permissions than support")
	}
	if _, ok := rolePermissions["auditor"]; !ok {
		t.Fatal("auditor role should be defined")
	}
	if _, ok := rolePermissions["billing_admin"]; ok {
		t.Fatal("billing_admin role should not be defined")
	}
}

func TestPlatformPermissionsForRoleReturnsCopy(t *testing.T) {
	s := &sPlatformAdmin{}
	permissions := s.PermissionsForRole("support")
	want := []service.PlatformPermission{
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionUserRead,
		service.PlatformPermissionAuditRead,
		service.PlatformPermissionConfigRead,
	}
	if len(permissions) != len(want) {
		t.Fatalf("len=%d want %d", len(permissions), len(want))
	}
	for i := range want {
		if permissions[i] != want[i] {
			t.Fatalf("permissions[%d]=%s want %s", i, permissions[i], want[i])
		}
	}
	permissions[0] = service.PlatformPermissionAdminManage
	if got := s.PermissionsForRole("support")[0]; got != service.PlatformPermissionTenantRead {
		t.Fatalf("PermissionsForRole should return a copy, got first permission %s", got)
	}
	if !containsPermission(s.PermissionsForRole("super_admin"), service.PlatformPermissionConfigManage) {
		t.Fatal("super_admin should be able to manage system config")
	}
	if containsPermission(s.PermissionsForRole("support"), service.PlatformPermissionConfigManage) {
		t.Fatal("support should not be able to manage system config")
	}
	if !containsPermission(s.PermissionsForRole("auditor"), service.PlatformPermissionConfigRead) {
		t.Fatal("auditor should be able to read system config")
	}
	if got := s.PermissionsForRole("unknown"); len(got) != 0 {
		t.Fatalf("unknown role permissions=%v want empty", got)
	}
}

func containsPermission(items []service.PlatformPermission, want service.PlatformPermission) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
