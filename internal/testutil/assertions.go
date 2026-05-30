package testutil

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

// AssertSuccess asserts that the response has a 2xx status and the envelope code is 0.
func AssertSuccess(t *testing.T, resp *TestResponse) {
	t.Helper()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("expected success, got status %d body %s", resp.StatusCode, resp.Body)
	}
	root := resp.JSON()
	if root == nil {
		t.Fatalf("expected JSON response, got %s", resp.Body)
	}
	if code, ok := root["code"].(float64); ok && code != 0 {
		if errInfo, ok := root["error"].(map[string]any); ok {
			t.Fatalf("expected code=0, got code=%.0f error=%v", code, errInfo)
		}
		t.Fatalf("expected code=0, got code=%.0f", code)
	}
}

// AssertStatus asserts the HTTP status code.
func AssertStatus(t *testing.T, resp *TestResponse, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		t.Fatalf("expected status %d, got %d body %s", expected, resp.StatusCode, resp.Body)
	}
}

// AssertErrorCode asserts the error code in the response envelope.
func AssertErrorCode(t *testing.T, resp *TestResponse, code string) {
	t.Helper()
	root := resp.JSON()
	if root == nil {
		t.Fatalf("expected JSON, got %s", resp.Body)
	}
	errInfo, _ := root["error"].(map[string]any)
	if errInfo == nil {
		t.Fatalf("expected error object, got %v", root)
	}
	gotCode, _ := errInfo["code"].(string)
	if gotCode != code {
		t.Fatalf("expected error code %q, got %q (full: %v)", code, gotCode, errInfo)
	}
}

// AssertCookieSet asserts that a cookie with the given name prefix exists in the response.
func AssertCookieSet(t *testing.T, resp *TestResponse, namePrefix string) {
	t.Helper()
	for _, c := range resp.Cookies {
		if strings.HasPrefix(c.Name, namePrefix) && c.Value != "" {
			return
		}
	}
	t.Fatalf("expected cookie with prefix %q to be set", namePrefix)
}

// AssertDBCount asserts the row count in a table.
func AssertDBCount(t *testing.T, table string, expected int) {
	t.Helper()
	count, err := g.DB().Model(table).Count()
	if err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if count != expected {
		t.Fatalf("expected %s count=%d, got %d", table, expected, count)
	}
}

// ParseDataField extracts a nested field from the response "data" object.
// E.g. ParseDataField(t, resp, "user.id") returns the user's ID.
func ParseDataField(t *testing.T, resp *TestResponse, path string) any {
	t.Helper()
	data := resp.JSONData()
	if data == nil {
		t.Fatalf("no data in response: %s", resp.Body)
	}
	parts := strings.Split(path, ".")
	current := any(data)
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("cannot navigate %q in response data: %s", path, resp.Body)
		}
		current = m[part]
	}
	return current
}

// ParseDataString extracts a string field from the response "data" object.
func ParseDataString(t *testing.T, resp *TestResponse, path string) string {
	t.Helper()
	v := ParseDataField(t, resp, path)
	s, ok := v.(string)
	if !ok {
		t.Fatalf("expected string at %q, got %T: %v", path, v, v)
	}
	return s
}

// ParseResponseUser extracts the user object from a register/login response.
func ParseResponseUser(t *testing.T, resp *TestResponse) map[string]any {
	t.Helper()
	data := resp.JSONData()
	if data == nil {
		t.Fatalf("no data in response: %s", resp.Body)
	}
	user, ok := data["user"].(map[string]any)
	if !ok {
		t.Fatalf("no user in response data: %s", resp.Body)
	}
	return user
}

// ParseResponseItems extracts the items array from a list response.
func ParseResponseItems(t *testing.T, resp *TestResponse, key string) []any {
	t.Helper()
	data := resp.JSONData()
	if data == nil {
		t.Fatalf("no data in response: %s", resp.Body)
	}
	items, ok := data[key].([]any)
	if !ok {
		t.Fatalf("no %s array in response data: %s", key, resp.Body)
	}
	return items
}

// toJSONString converts a value to a compact JSON string.
func toJSONString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
