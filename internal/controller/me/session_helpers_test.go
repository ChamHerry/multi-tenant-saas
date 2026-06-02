package me

import "testing"

func TestMaskSessionIP(t *testing.T) {
	cases := map[string]string{
		"":            "unknown",
		"not-an-ip":   "unknown",
		"192.168.1.5": "192.168.*.*",
		"10.1.2.3":    "10.1.*.*",
		"2001:db8::1": "2001:db8:0:0::*",
	}
	for input, want := range cases {
		if got := maskSessionIP(input); got != want {
			t.Fatalf("maskSessionIP(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseSessionDevice(t *testing.T) {
	chrome := parseSessionDevice("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	if chrome.Browser != "Chrome" || chrome.BrowserVersion != "125.0.0.0" || chrome.OS != "macOS" || chrome.IsMobile {
		t.Fatalf("unexpected chrome device = %+v", chrome)
	}

	safari := parseSessionDevice("Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1")
	if safari.Browser != "Safari" || safari.BrowserVersion != "17.0" || safari.OS != "iOS" || !safari.IsMobile {
		t.Fatalf("unexpected safari device = %+v", safari)
	}

	unknown := parseSessionDevice("")
	if unknown.Browser != "Unknown" || unknown.OS != "Unknown" {
		t.Fatalf("unexpected unknown device = %+v", unknown)
	}
}
