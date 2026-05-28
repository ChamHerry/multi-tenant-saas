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

func TestIntersectScopesUsesKeyAndGrantIntersection(t *testing.T) {
	got := intersectScopes([]string{"tenant:read", "member:manage"}, []string{"tenant:read"})
	if len(got) != 1 || got[0] != "tenant:read" {
		t.Fatalf("intersectScopes=%v want [tenant:read]", got)
	}
	got = intersectScopes([]string{"*"}, []string{"tenant:read"})
	if len(got) != 1 || got[0] != "tenant:read" {
		t.Fatalf("intersectScopes star key=%v want [tenant:read]", got)
	}
	got = intersectScopes([]string{"tenant:read"}, nil)
	if len(got) != 1 || got[0] != "tenant:read" {
		t.Fatalf("intersectScopes empty grant=%v want key scopes", got)
	}
}

func TestTenantOptionalRouteIncludesPersonalAPIKeys(t *testing.T) {
	if !tenantOptionalPath("/api/v1/api-keys", "GET") || !tenantOptionalPath("/api/v1/api-keys", "POST") {
		t.Fatal("/api/v1/api-keys should not require tenant selector")
	}
	if !tenantOptionalPath("/api/v1/api-keys/123e4567-e89b-12d3-a456-426614174000", "DELETE") {
		t.Fatal("DELETE /api/v1/api-keys/{id} should not require tenant selector")
	}
	if tenantOptionalPath("/api/v1/me/api-keys", "GET") {
		t.Fatal("/api/v1/me/api-keys should not remain a personal api key route")
	}
}
