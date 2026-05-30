package membership

import (
	"testing"

	"multi-tenant-saas/internal/service"
)

func TestDefaultTenantNameAndSlug(t *testing.T) {
	user := &service.User{ID: "123e4567-e89b-12d3-a456-426614174000", Email: "admin@example.com", DisplayName: "Admin"}
	if got := defaultTenantName(user); got != "admin@example.com 的组织" {
		t.Fatalf("defaultTenantName(email)=%q", got)
	}
	if got := defaultTenantSlug(user); got != "admin-example-com-123e4567" {
		t.Fatalf("defaultTenantSlug(email)=%q", got)
	}

	user.Email = ""
	if got := defaultTenantName(user); got != "Admin 的组织" {
		t.Fatalf("defaultTenantName(display)=%q", got)
	}
	if got := defaultTenantSlug(user); got != "admin-123e4567" {
		t.Fatalf("defaultTenantSlug(display)=%q", got)
	}
}
