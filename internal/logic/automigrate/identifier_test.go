package automigrate

import "testing"

func TestValidateIdentifier(t *testing.T) {
	valid := []string{"public", "tenant_a1b2c3d4", "graph_tenant_a1b2c3d4_code_graph"}
	for _, input := range valid {
		if err := ValidateIdentifier(input); err != nil {
			t.Fatalf("ValidateIdentifier(%q) unexpected error: %v", input, err)
		}
	}
	invalid := []string{"Tenant_A", "tenant-a", "tenant.x", "1tenant", "tenant_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	for _, input := range invalid {
		if err := ValidateIdentifier(input); err == nil {
			t.Fatalf("ValidateIdentifier(%q) expected error", input)
		}
	}
}

func TestIsPostgresType(t *testing.T) {
	for _, input := range []string{"pgsql", "postgres", "postgresql", "PostgreSQL"} {
		if !isPostgresType(input) {
			t.Fatalf("isPostgresType(%q)=false", input)
		}
	}
	if isPostgresType("mysql") {
		t.Fatal("isPostgresType(mysql)=true")
	}
}
