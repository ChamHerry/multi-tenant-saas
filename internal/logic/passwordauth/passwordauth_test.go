package passwordauth

import "testing"

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  User@Example.COM "); got != "user@example.com" {
		t.Fatalf("normalizeEmail=%q", got)
	}
}

func TestValidatePassword(t *testing.T) {
	if err := validatePassword(t.Context(), "short"); err == nil {
		t.Fatal("validatePassword short expected error")
	}
	if err := validatePassword(t.Context(), "0123456789abcde"); err != nil {
		t.Fatalf("validatePassword valid error = %v", err)
	}
	long := "0123456789abcde0123456789abcde0123456789abcde0123456789abcde0123456789abcde"
	if err := validatePassword(t.Context(), long); err == nil {
		t.Fatal("validatePassword >72 bytes expected error")
	}
}
