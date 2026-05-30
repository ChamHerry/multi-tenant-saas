package member_test

import (
	"testing"

	"github.com/gogf/gf/v2/test/gtest"

	_ "multi-tenant-saas/internal/logic"
	"multi-tenant-saas/internal/service"
)

func TestMemberRBACServiceRegistered(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.AssertNE(service.RBAC(), nil)
	})
}

func TestMembershipServiceRegistered(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.AssertNE(service.TenantMembershipService(), nil)
	})
}

func TestPermissionsForRole(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		perms := service.RBAC().PermissionsForRole("owner")
		t.AssertGT(len(perms), 0)
		viewerPerms := service.RBAC().PermissionsForRole("viewer")
		t.AssertGT(len(viewerPerms), 0)
		t.AssertLT(len(viewerPerms), len(perms))
	})
}
