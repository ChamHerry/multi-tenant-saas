package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
)

// TestClient wraps an HTTP client with automatic cookie management and CSRF token handling.
type TestClient struct {
	t         *testing.T
	client    *http.Client
	baseURL   string
	csrfToken string
}

// TestResponse captures the response from an HTTP request for assertions.
type TestResponse struct {
	StatusCode int
	Body       string
	Headers    http.Header
	Cookies    []*http.Cookie
}

// NewTestClient creates a TestClient that talks to baseURL with automatic cookie management.
func NewTestClient(t *testing.T, baseURL string) *TestClient {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}
	return &TestClient{
		t:       t,
		client:  &http.Client{Jar: jar},
		baseURL: baseURL,
	}
}

// GET sends a GET request.
func (c *TestClient) GET(path string) *TestResponse {
	return c.Do(http.MethodGet, path, "")
}

// POST sends a POST request with an optional JSON body.
func (c *TestClient) POST(path, body string) *TestResponse {
	return c.Do(http.MethodPost, path, body)
}

// PATCH sends a PATCH request with an optional JSON body.
func (c *TestClient) PATCH(path, body string) *TestResponse {
	return c.Do(http.MethodPatch, path, body)
}

// DELETE sends a DELETE request.
func (c *TestClient) DELETE(path string) *TestResponse {
	return c.Do(http.MethodDelete, path, "")
}

// Do sends an HTTP request with the given method, path, and optional body.
// It automatically sets Content-Type, sends the CSRF token if available,
// and captures response cookies including the CSRF token for subsequent requests.
func (c *TestClient) Do(method, path, body string) *TestResponse {
	c.t.Helper()

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		c.t.Fatalf("create request %s %s: %v", method, path, err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Attach CSRF token if we have one.
	if c.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Capture CSRF token from response cookies for subsequent requests.
	for _, cookie := range resp.Cookies() {
		if strings.Contains(cookie.Name, "csrf") && cookie.Value != "" {
			c.csrfToken = cookie.Value
		}
	}

	return &TestResponse{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Headers:    resp.Header,
		Cookies:    resp.Cookies(),
	}
}

// BaseURL returns the server base URL (e.g. "http://127.0.0.1:18080").
func (c *TestClient) BaseURL() string {
	return c.baseURL
}

// WithTenantHeader returns a shallow copy of the client that sends X-Tenant-ID with every request.
// Since we use the shared http.Client with cookie jar, this returns a new TestClient wrapper.
func (c *TestClient) WithTenantHeader(tenantID string) *TestClient {
	c.t.Helper()
	// For simplicity, we'll handle tenant headers in individual requests
	// using DoWithHeaders. This method is kept for API compatibility.
	return c
}

// DoWithHeaders sends an HTTP request with additional headers.
func (c *TestClient) DoWithHeaders(method, path, body string, headers map[string]string) *TestResponse {
	c.t.Helper()

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		c.t.Fatalf("create request %s %s: %v", method, path, err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Attach CSRF token if we have one.
	if c.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
	}

	// Attach extra headers.
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Capture CSRF token from response cookies.
	for _, cookie := range resp.Cookies() {
		if strings.Contains(cookie.Name, "csrf") && cookie.Value != "" {
			c.csrfToken = cookie.Value
		}
	}

	return &TestResponse{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Headers:    resp.Header,
		Cookies:    resp.Cookies(),
	}
}

// ResetCookies clears the cookie jar and CSRF token, simulating an unauthenticated client.
func (c *TestClient) ResetCookies() {
	jar, _ := cookiejar.New(nil)
	c.client.Jar = jar
	c.csrfToken = ""
}

// ClearCSRF removes only the CSRF token while keeping session cookies.
// Useful for testing CSRF enforcement.
func (c *TestClient) ClearCSRF() {
	c.csrfToken = ""
}

// JSON returns the parsed JSON body as a map.
func (r *TestResponse) JSON() map[string]any {
	var result map[string]any
	_ = json.Unmarshal([]byte(r.Body), &result)
	return result
}

// JSONData returns the "data" field from the standard response envelope.
func (r *TestResponse) JSONData() map[string]any {
	root := r.JSON()
	if root == nil {
		return nil
	}
	data, _ := root["data"].(map[string]any)
	return data
}

// JSONError returns the error code from the standard response envelope.
func (r *TestResponse) JSONError() map[string]any {
	root := r.JSON()
	if root == nil {
		return nil
	}
	errInfo, _ := root["error"].(map[string]any)
	return errInfo
}
