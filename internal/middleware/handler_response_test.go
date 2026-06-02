package middleware

import (
	"net/http"
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"
)

func TestStatusFromCode(t *testing.T) {
	cases := []struct {
		code gcode.Code
		want int
	}{
		{gcode.CodeOK, http.StatusOK},
		{gcode.CodeValidationFailed, http.StatusBadRequest},
		{gcode.CodeNotAuthorized, http.StatusForbidden},
		{gcode.CodeNotFound, http.StatusNotFound},
		{gcode.CodeDbOperationError, http.StatusInternalServerError},
	}
	for _, tt := range cases {
		if got := statusFromCode(tt.code); got != tt.want {
			t.Fatalf("statusFromCode(%v)=%d want %d", tt.code, got, tt.want)
		}
	}
}

func TestStatusFromCodeConflict(t *testing.T) {
	if got := statusFromCode(gcode.New(409001, "Conflict", nil)); got != http.StatusConflict {
		t.Fatalf("statusFromCode(conflict)=%d", got)
	}
}

func TestStatusFromCodeHTTPStatusCustomCodes(t *testing.T) {
	cases := []struct {
		code gcode.Code
		want int
	}{
		{gcode.New(http.StatusNotFound, "TokenNotFound", nil), http.StatusNotFound},
		{gcode.New(http.StatusGone, "TokenExpired", nil), http.StatusGone},
		{gcode.New(http.StatusTooManyRequests, "RateLimited", nil), http.StatusTooManyRequests},
	}
	for _, tt := range cases {
		if got := statusFromCode(tt.code); got != tt.want {
			t.Fatalf("statusFromCode(%d)=%d want %d", tt.code.Code(), got, tt.want)
		}
	}
}
