package auth

import "testing"

func TestBearerToken(t *testing.T) {
	if got := bearerToken("Bearer abc"); got != "abc" {
		t.Fatalf("bearerToken=%q", got)
	}
	if got := bearerToken("bearer abc"); got != "abc" {
		t.Fatalf("bearerToken lower=%q", got)
	}
	if got := bearerToken("Basic abc"); got != "" {
		t.Fatalf("bearerToken basic=%q", got)
	}
}
