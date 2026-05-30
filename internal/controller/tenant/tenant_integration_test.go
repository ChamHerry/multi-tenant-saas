package tenant_test

import (
	"testing"

	"github.com/gogf/gf/v2/test/gtest"

	_ "multi-tenant-saas/internal/logic"
	"multi-tenant-saas/internal/service"
)

func TestTenantServiceRegistered(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.AssertNE(service.TenantProvision(), nil)
	})
}

func TestTenantAdminServiceRegistered(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.AssertNE(service.TenantAdmin(), nil)
	})
}

func TestTenantMembershipServiceRegistered(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.AssertNE(service.TenantMembershipService(), nil)
	})
}

func TestTenantLifecycleServiceRegistered(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.AssertNE(service.TenantLifecycle(), nil)
	})
}
