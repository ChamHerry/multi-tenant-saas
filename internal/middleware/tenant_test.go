package middleware

import "testing"

func TestTenantFromPath(t *testing.T) {
	cases := map[string]string{
		"/api/v1/tenants/acme":            "acme",
		"/api/v1/tenants/acme/members/u1": "acme",
		"/api/v1/api-keys":                "",
		"/api/v1/tenants/123/members/u1":  "123",
	}
	for input, want := range cases {
		if got := tenantFromPath(input); got != want {
			t.Fatalf("tenantFromPath(%q)=%q want %q", input, got, want)
		}
	}
}

func TestTenantOptionalRouteIncludesAccessSnapshot(t *testing.T) {
	if !tenantOptionalPath("/api/v1/me/access", "GET") {
		t.Fatal("/api/v1/me/access should not require tenant selector")
	}
}

func TestTenantOptionalRouteDoesNotExposeOrganizationMaterialize(t *testing.T) {
	if tenantOptionalPath("/api/v1/me/onboarding", "GET") {
		t.Fatal("/api/v1/me/onboarding is not part of the tenant-only model")
	}
	if tenantOptionalPath("/api/v1/organizations/123e4567-e89b-12d3-a456-426614174000/materialize", "POST") {
		t.Fatal("organization materialize endpoint should not be exposed")
	}
}
