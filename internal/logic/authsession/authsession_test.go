package authsession

import (
	"context"
	"strconv"
	"testing"
	"time"

	"multi-tenant-saas/internal/service"
)

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

func TestCookieAttrsSecureDefault(t *testing.T) {
	service.RegisterConfig(testConfig{})

	attrs := cookieAttrs(context.Background())
	if !attrs.Secure {
		t.Fatal("cookieAttrs should default Secure=true")
	}

	service.RegisterConfig(testConfig{"auth.session.cookie.secure": "false"})
	attrs = cookieAttrs(context.Background())
	if attrs.Secure {
		t.Fatal("explicit auth.session.cookie.secure=false should override Secure default")
	}
}

type testConfig map[string]string

func (c testConfig) GetString(_ context.Context, key string, defaultVal string) string {
	if value, ok := c[key]; ok {
		return value
	}
	return defaultVal
}

func (c testConfig) GetInt(_ context.Context, key string, defaultVal int) int {
	if value, ok := c[key]; ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func (c testConfig) GetFloat(_ context.Context, key string, defaultVal float64) float64 {
	if value, ok := c[key]; ok {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func (c testConfig) GetBool(_ context.Context, key string, defaultVal bool) bool {
	if value, ok := c[key]; ok {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func (c testConfig) GetDuration(_ context.Context, key string, defaultVal time.Duration) time.Duration {
	if value, ok := c[key]; ok {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func (c testConfig) GetStrings(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := c[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (c testConfig) Set(_ context.Context, params *service.ConfigSetParams) error {
	c[params.Key] = params.Value
	return nil
}

func (c testConfig) Delete(_ context.Context, key string) error {
	delete(c, key)
	return nil
}
