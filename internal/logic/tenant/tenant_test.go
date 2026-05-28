package tenant

import (
	"testing"

	"repomind-temp/internal/service"
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
