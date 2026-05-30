package tenant

import (
	"strings"
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"multi-tenant-saas/internal/service"
)

func TestValidateCreateTenantInputRequiresOwnerOrSystem(t *testing.T) {
	err := validateCreateTenantInput(service.CreateTenantInput{
		Name: "Acme",
		Slug: "acme",
	})
	if err == nil {
		t.Fatal("validateCreateTenantInput() expected owner requirement error")
	}
	err = validateCreateTenantInput(service.CreateTenantInput{
		Name:            "Acme",
		Slug:            "acme",
		SystemOwnerless: true,
	})
	if err != nil {
		t.Fatalf("validateCreateTenantInput(system) error = %v", err)
	}
	err = validateCreateTenantInput(service.CreateTenantInput{
		Name:        "Acme",
		Slug:        "acme",
		OwnerUserID: "123e4567-e89b-12d3-a456-426614174000",
	})
	if err != nil {
		t.Fatalf("validateCreateTenantInput(owner) error = %v", err)
	}
}

func TestValidateCreateTenantInputSlugRules(t *testing.T) {
	validSlugs := []string{
		"abc",
		"a-b",
		strings.Repeat("a", 80),
	}
	for _, slug := range validSlugs {
		t.Run("valid_"+slug, func(t *testing.T) {
			err := validateCreateTenantInput(service.CreateTenantInput{
				Name:            "Acme",
				Slug:            slug,
				SystemOwnerless: true,
			})
			if err != nil {
				t.Fatalf("validateCreateTenantInput(%q) error = %v", slug, err)
			}
		})
	}

	invalidSlugs := []string{
		"",
		"a",
		"ab",
		"a-",
		"-a",
		"a_b",
		strings.Repeat("a", 81),
	}
	for _, slug := range invalidSlugs {
		t.Run("invalid_"+slug, func(t *testing.T) {
			err := validateCreateTenantInput(service.CreateTenantInput{
				Name:            "Acme",
				Slug:            slug,
				SystemOwnerless: true,
			})
			if err == nil {
				t.Fatalf("validateCreateTenantInput(%q) expected slug error", slug)
			}
			if code := gerror.Code(err); code != gcode.CodeInvalidParameter {
				t.Fatalf("validateCreateTenantInput(%q) code = %v, want %v", slug, code, gcode.CodeInvalidParameter)
			}
			if !strings.Contains(err.Error(), tenantSlugRuleMessage) {
				t.Fatalf("validateCreateTenantInput(%q) error = %q, want slug rule message", slug, err.Error())
			}
		})
	}
}
