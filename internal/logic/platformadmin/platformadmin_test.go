package platformadmin

import (
	"testing"

	"repomind-temp/internal/service"
)

func TestPlatformRolePermissions(t *testing.T) {
	if len(rolePermissions["super_admin"]) <= len(rolePermissions["support"]) {
		t.Fatal("super_admin should have more permissions than support")
	}
	if _, ok := rolePermissions["auditor"]; !ok {
		t.Fatal("auditor role should be defined")
	}
}

func TestPlatformPermissionsForRoleReturnsCopy(t *testing.T) {
	s := &sPlatformAdmin{}
	permissions := s.PermissionsForRole("support")
	want := []service.PlatformPermission{
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionUserRead,
		service.PlatformPermissionAuditRead,
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
	if got := s.PermissionsForRole("unknown"); len(got) != 0 {
		t.Fatalf("unknown role permissions=%v want empty", got)
	}
}
