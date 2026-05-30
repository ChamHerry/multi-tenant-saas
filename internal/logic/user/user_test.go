package user

import (
	"testing"

	"multi-tenant-saas/internal/service"
)

func TestValidateIdentityInput(t *testing.T) {
	if err := validateIdentityInput(service.EnsureUserByIdentityInput{Provider: "github", AuthID: "42"}); err != nil {
		t.Fatalf("validateIdentityInput(valid) error = %v", err)
	}
	if err := validateIdentityInput(service.EnsureUserByIdentityInput{Provider: "unknown", AuthID: "42"}); err == nil {
		t.Fatal("validateIdentityInput(invalid provider) expected error")
	}
	if err := validateIdentityInput(service.EnsureUserByIdentityInput{Provider: "github"}); err == nil {
		t.Fatal("validateIdentityInput(empty auth id) expected error")
	}
}
