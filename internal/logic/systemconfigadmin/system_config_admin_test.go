package systemconfigadmin

import "testing"

func TestNormalizeKey(t *testing.T) {
	cases := []struct {
		name string
		key  string
		ok   bool
	}{
		{name: "valid dotted", key: "auth.session.secret", ok: true},
		{name: "valid namespace", key: "system_config:test-key", ok: true},
		{name: "empty", key: " ", ok: false},
		{name: "space", key: "bad key", ok: false},
		{name: "slash", key: "bad/key", ok: false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeKey(tt.key)
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestValidateAndNormalizeValue(t *testing.T) {
	cases := []struct {
		name      string
		valueType string
		value     string
		want      string
		ok        bool
	}{
		{name: "string", valueType: "string", value: "  keep spacing  ", want: "  keep spacing  ", ok: true},
		{name: "number", valueType: "number", value: " 12.5 ", want: "12.5", ok: true},
		{name: "invalid number", valueType: "number", value: "abc", ok: false},
		{name: "bool true", valueType: "bool", value: "TRUE", want: "true", ok: true},
		{name: "invalid bool", valueType: "bool", value: "enabled", ok: false},
		{name: "json compact", valueType: "json", value: `{"b":2,"a":1}`, want: `{"a":1,"b":2}`, ok: true},
		{name: "invalid json", valueType: "json", value: `{bad}`, ok: false},
		{name: "secret", valueType: "secret", value: "raw-secret", want: "raw-secret", ok: true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateAndNormalizeValue(tt.valueType, tt.value)
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatal("expected error")
			}
			if tt.ok && got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestSensitiveKeyDetection(t *testing.T) {
	if !isSensitiveKey("auth.session.secret") {
		t.Fatal("auth.session.secret should be treated as sensitive")
	}
	if !isSensitiveKey("provider:secret") {
		t.Fatal("provider:secret should be treated as sensitive")
	}
	if isSensitiveKey("auth.session.absoluteTTL") {
		t.Fatal("non-secret config should not be treated as sensitive")
	}
}
