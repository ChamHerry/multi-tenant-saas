package invitation

import (
	"testing"

	"repomind-temp/internal/service"
)

func TestValidateCreateInput(t *testing.T) {
	valid := service.CreateTenantInvitationInput{
		TenantID:        "123e4567-e89b-12d3-a456-426614174000",
		InvitedByUserID: "123e4567-e89b-12d3-a456-426614174001",
		InviteeEmail:    "User@Example.Test",
		Role:            "viewer",
	}
	if err := validateCreateInput(valid); err != nil {
		t.Fatalf("validateCreateInput(valid) error = %v", err)
	}
	valid.InviteeEmail = "bad email"
	if err := validateCreateInput(valid); err == nil {
		t.Fatal("validateCreateInput(invalid email) expected error")
	}
	valid.InviteeEmail = "user@example.test"
	valid.Role = "root"
	if err := validateCreateInput(valid); err == nil {
		t.Fatal("validateCreateInput(invalid role) expected error")
	}
}

func TestInvitationTokenHashStable(t *testing.T) {
	if hashToken("abc") != hashToken("abc") {
		t.Fatal("hashToken should be stable")
	}
	if hashToken("abc") == hashToken("def") {
		t.Fatal("hashToken should differ for different tokens")
	}
}
