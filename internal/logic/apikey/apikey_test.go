package apikey

import "testing"

func TestHashRawKeyUsesSecret(t *testing.T) {
	a := hashRawKey("rpm_demo_key", "secret-a")
	b := hashRawKey("rpm_demo_key", "secret-b")
	if a == b {
		t.Fatal("hashRawKey should depend on secret")
	}
	if a != hashRawKey("rpm_demo_key", "secret-a") {
		t.Fatal("hashRawKey should be deterministic")
	}
}

func TestParseTextArray(t *testing.T) {
	got := parseTextArray(`{tenant:read,member:read}`)
	if len(got) != 2 || got[0] != "tenant:read" || got[1] != "member:read" {
		t.Fatalf("parseTextArray unexpected: %#v", got)
	}
	if got := parseTextArray("{}"); len(got) != 0 {
		t.Fatalf("parseTextArray(empty)=%#v", got)
	}
}

func TestRawKeyPrefix(t *testing.T) {
	if got := rawKeyPrefix("rpm_12345678_abcdef"); got != "rpm_12345678" {
		t.Fatalf("rawKeyPrefix=%s", got)
	}
}

func TestTextArrayLiteralAndExpiresAt(t *testing.T) {
	if got := textArrayLiteral([]string{"tenant:read", "member:read"}); got != `{"tenant:read","member:read"}` {
		t.Fatalf("textArrayLiteral=%s", got)
	}
	if got := expiresAtString(nil); got != "" {
		t.Fatalf("expiresAtString(nil)=%q", got)
	}
}

func TestParseJSONArrayScopes(t *testing.T) {
	got := parseTextArray(`["tenant:read","member:read"]`)
	if len(got) != 2 || got[0] != "tenant:read" || got[1] != "member:read" {
		t.Fatalf("parseTextArray json unexpected: %#v", got)
	}
}
