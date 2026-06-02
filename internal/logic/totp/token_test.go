package totp

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateAndValidateTokenWithSecret(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Unix(1000, 0).UTC()
	token := generateTokenWithSecret("user-1", now, time.Minute, secret)
	userID, err := validateTokenWithSecret(token, now.Add(30*time.Second), secret)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if userID != "user-1" {
		t.Fatalf("userID=%q", userID)
	}
}

func TestValidateTokenExpired(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Unix(1000, 0).UTC()
	token := generateTokenWithSecret("user-1", now, time.Minute, secret)
	if _, err := validateTokenWithSecret(token, now.Add(2*time.Minute), secret); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestValidateTokenTampered(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Unix(1000, 0).UTC()
	token := generateTokenWithSecret("user-1", now, time.Minute, secret)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token format: %s", token)
	}
	parts[0] = parts[0] + "x"
	if _, err := validateTokenWithSecret(strings.Join(parts, "."), now, secret); err == nil {
		t.Fatal("expected tampered token error")
	}
}

func TestTokenSecretDomainSeparation(t *testing.T) {
	userBytes := []byte("user-1")
	expiresBytes := []byte("12345678")
	first := signTOTPToken(userBytes, expiresBytes, []byte("secret-a"))
	second := signTOTPToken(userBytes, expiresBytes, []byte("secret-b"))
	if sameBytes(first, second) {
		t.Fatal("different secrets should produce different signatures")
	}
}
