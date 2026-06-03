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
	if err := validateRuntimeValues(runtimeValues{Env: "prod", DevHeader: false, PasswordEnabled: true, SessionSecret: "change-me", APIKeySecret: "0123456789abcdef0123456789abcdef"}); err == nil {
		t.Fatal("prod session change-me should fail")
	}
	if err := validateRuntimeValues(runtimeValues{Env: "prod", DevHeader: false, PasswordEnabled: true, SessionSecret: "0123456789abcdef0123456789abcdef", APIKeySecret: "change-me"}); err == nil {
		t.Fatal("prod api key change-me should fail")
	}
	if err := validateRuntimeValues(runtimeValues{Env: "local", DevHeader: true, PasswordEnabled: true, SessionSecret: "change-me", APIKeySecret: "change-me"}); err != nil {
		t.Fatalf("local placeholders should be allowed by strict runtime validator: %v", err)
	}
	if err := validateRuntimeValues(runtimeValues{Env: "prod", DevHeader: true, PasswordEnabled: true, SessionSecret: "0123456789abcdef0123456789abcdef", APIKeySecret: "0123456789abcdef0123456789abcdef"}); err == nil {
		t.Fatal("dev header must fail outside local/test")
	}
}
