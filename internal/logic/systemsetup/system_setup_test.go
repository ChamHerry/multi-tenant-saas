package systemsetup

import "testing"

func TestIsSafeSecret(t *testing.T) {
	if isSafeSecret("docker-dev-session-secret-change-me") {
		t.Fatal("change-me secret must not be considered safe")
	}
	if isSafeSecret("short") {
		t.Fatal("short secret must not be considered safe")
	}
	if !isSafeSecret("0123456789abcdef0123456789abcdef") {
		t.Fatal("32-byte non-placeholder secret should be safe")
	}
}

func TestValidateRuntimeValues(t *testing.T) {
	if err := validateRuntimeValues(runtimeValues{SessionSecret: "change-me", APIKeySecret: "0123456789abcdef0123456789abcdef"}); err == nil {
		t.Fatal("session change-me should fail")
	}
	if err := validateRuntimeValues(runtimeValues{SessionSecret: "0123456789abcdef0123456789abcdef", APIKeySecret: "change-me"}); err == nil {
		t.Fatal("api key change-me should fail")
	}
	if err := validateRuntimeValues(runtimeValues{SessionSecret: "short", APIKeySecret: "0123456789abcdef0123456789abcdef"}); err == nil {
		t.Fatal("short session secret should fail")
	}
	if err := validateRuntimeValues(runtimeValues{SessionSecret: "0123456789abcdef0123456789abcdef", APIKeySecret: "short"}); err == nil {
		t.Fatal("short api key secret should fail")
	}
	if err := validateRuntimeValues(runtimeValues{SessionSecret: "0123456789abcdef0123456789abcdef", APIKeySecret: "fedcba9876543210fedcba9876543210"}); err != nil {
		t.Fatalf("safe runtime secrets should pass: %v", err)
	}
}
