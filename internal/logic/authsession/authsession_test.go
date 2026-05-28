package authsession

import "testing"

func TestParseSessionCookie(t *testing.T) {
	sessionID := "00000000-0000-0000-0000-000000000001"
	gotID, gotSecret, err := parseSessionCookie(sessionID + ".secret")
	if err != nil {
		t.Fatalf("parseSessionCookie valid error = %v", err)
	}
	if gotID != sessionID || gotSecret != "secret" {
		t.Fatalf("parseSessionCookie = %q %q", gotID, gotSecret)
	}
	if _, _, err = parseSessionCookie("invalid"); err == nil {
		t.Fatal("parseSessionCookie invalid expected error")
	}
}

func TestHashTokenStable(t *testing.T) {
	first := hashToken("token", "secret")
	second := hashToken("token", "secret")
	if first == "" || first != second {
		t.Fatalf("hashToken not stable: %q %q", first, second)
	}
	if first == hashToken("token", "other") {
		t.Fatal("hashToken should include secret")
	}
}
