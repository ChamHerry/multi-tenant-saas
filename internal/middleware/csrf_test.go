package middleware

import (
	"net/http"
	"testing"
)

func TestCSRFSafeMethod(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace, "get"} {
		if !csrfSafeMethod(method) {
			t.Fatalf("%s should be csrf-safe", method)
		}
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if csrfSafeMethod(method) {
			t.Fatalf("%s should not be csrf-safe", method)
		}
	}
}
