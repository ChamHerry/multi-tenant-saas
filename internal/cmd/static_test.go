package cmd

import (
	"net/http"
	"testing"
)

func TestShouldServeSPAFallback(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{name: "root", method: http.MethodGet, path: "/", want: true},
		{name: "spa route", method: http.MethodGet, path: "/tenants/abc", want: true},
		{name: "head spa route", method: http.MethodHead, path: "/members", want: true},
		{name: "api exact", method: http.MethodGet, path: "/api", want: false},
		{name: "api nested", method: http.MethodGet, path: "/api/v1/me", want: false},
		{name: "health", method: http.MethodGet, path: "/healthz", want: false},
		{name: "ready", method: http.MethodGet, path: "/readyz", want: false},
		{name: "swagger", method: http.MethodGet, path: "/swagger", want: false},
		{name: "api json", method: http.MethodGet, path: "/api.json", want: false},
		{name: "asset miss", method: http.MethodGet, path: "/assets/missing.js", want: false},
		{name: "post", method: http.MethodPost, path: "/tenants", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldServeSPAFallback(tt.method, tt.path); got != tt.want {
				t.Fatalf("shouldServeSPAFallback(%q, %q) = %v, want %v", tt.method, tt.path, got, tt.want)
			}
		})
	}
}
