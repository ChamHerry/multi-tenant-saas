//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	ptotp "github.com/pquerna/otp/totp"

	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/internal/testutil"
)

type fakeOAuthProfile struct {
	ID       string
	Email    string
	Name     string
	Avatar   string
	Verified bool
}

type fakeOAuthProvider struct {
	server  *httptest.Server
	mu      sync.Mutex
	profile fakeOAuthProfile
	codes   map[string]fakeOAuthProfile
	next    int
}

func newFakeOAuthProvider(t *testing.T) *fakeOAuthProvider {
	t.Helper()
	fake := &fakeOAuthProvider{codes: map[string]fakeOAuthProfile{}}
	router := http.NewServeMux()
	router.HandleFunc("/authorize", fake.authorize)
	router.HandleFunc("/token", fake.token)
	router.HandleFunc("/github/user", fake.githubUser)
	router.HandleFunc("/github/emails", fake.githubEmails)
	router.HandleFunc("/google/userinfo", fake.googleUserInfo)
	fake.server = httptest.NewServer(router)
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *fakeOAuthProvider) setProfile(profile fakeOAuthProfile) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.profile = profile
}

func (f *fakeOAuthProvider) authorize(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	if redirectURI == "" || state == "" {
		http.Error(w, "missing redirect_uri/state", http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	f.next++
	code := "code-" + strconv.Itoa(f.next)
	f.codes[code] = f.profile
	f.mu.Unlock()
	location, _ := url.Parse(redirectURI)
	query := location.Query()
	query.Set("code", code)
	query.Set("state", state)
	location.RawQuery = query.Encode()
	http.Redirect(w, r, location.String(), http.StatusFound)
}

func (f *fakeOAuthProvider) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	code := r.Form.Get("code")
	f.mu.Lock()
	_, ok := f.codes[code]
	f.mu.Unlock()
	if !ok {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"access_token":"token-%s","token_type":"Bearer","expires_in":3600}`, code)
}

func (f *fakeOAuthProvider) profileForRequest(r *http.Request) fakeOAuthProfile {
	token := r.Header.Get("Authorization")
	code := ""
	if len(token) > len("Bearer token-") {
		code = token[len("Bearer token-"):]
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if profile, ok := f.codes[code]; ok {
		return profile
	}
	return f.profile
}

func (f *fakeOAuthProvider) githubUser(w http.ResponseWriter, r *http.Request) {
	profile := f.profileForRequest(r)
	id, _ := strconv.ParseInt(profile.ID, 10, 64)
	writeJSON(w, map[string]any{"id": id, "login": profile.Name, "name": profile.Name, "avatar_url": profile.Avatar})
}

func (f *fakeOAuthProvider) githubEmails(w http.ResponseWriter, r *http.Request) {
	profile := f.profileForRequest(r)
	writeJSON(w, []map[string]any{{"email": profile.Email, "primary": true, "verified": profile.Verified}})
}

func (f *fakeOAuthProvider) googleUserInfo(w http.ResponseWriter, r *http.Request) {
	profile := f.profileForRequest(r)
	writeJSON(w, map[string]any{"sub": profile.ID, "email": profile.Email, "email_verified": profile.Verified, "name": profile.Name, "picture": profile.Avatar})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func TestOAuth_ProviderListAndDisabledStart(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	setOAuthProviderConfig(t, "github", nil, false)
	setOAuthProviderConfig(t, "google", nil, false)

	resp := suite.Client.GET("/api/v1/auth/oauth/providers")
	testutil.AssertSuccess(t, resp)
	providers := testutil.ParseResponseItems(t, resp, "providers")
	if len(providers) != 2 {
		t.Fatalf("expected 2 provider entries, got %v", resp.Body)
	}
	for _, item := range providers {
		provider := item.(map[string]any)
		if provider["enabled"] == true {
			t.Fatalf("expected disabled provider entry, got %v", provider)
		}
	}

	client := newNoRedirectHTTPClient(t)
	disabled := rawRequest(t, client, http.MethodGet, suite.Client.BaseURL()+"/api/v1/auth/oauth/github?redirect=/tenants", "")
	defer disabled.Body.Close()
	if disabled.StatusCode != http.StatusFound {
		t.Fatalf("expected disabled provider start to redirect to login, got %d", disabled.StatusCode)
	}
	if loc := disabled.Header.Get("Location"); !containsAll(loc, "/login", "oauth_error=oauth_provider_disabled") {
		t.Fatalf("expected oauth_provider_disabled login redirect, got %q", loc)
	}
}

func TestOAuth_GitHubNewUserFlowAndStateReplay(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	fake := newFakeOAuthProvider(t)
	fake.setProfile(fakeOAuthProfile{ID: "1001", Email: "oauth-new@example.com", Name: "OAuth New", Avatar: "https://example.com/a.png", Verified: true})
	setOAuthProviderConfig(t, "github", fake, true)

	client := newNoRedirectHTTPClient(t)
	finalResp, callbackURL := runOAuthBrowserFlow(t, client, "github", "/tenants")
	defer finalResp.Body.Close()
	if finalResp.StatusCode != http.StatusFound {
		t.Fatalf("expected final callback redirect, got %d", finalResp.StatusCode)
	}
	if loc := finalResp.Header.Get("Location"); !containsAll(loc, "/tenants") {
		t.Fatalf("expected final redirect to /tenants, got %q", loc)
	}
	assertRawSessionOK(t, client, "oauth-new@example.com")
	if count := scalarInt(t, "SELECT COUNT(*) FROM user_identities WHERE provider = 'github' AND auth_id = '1001' AND email = $1", "oauth-new@example.com"); count != 1 {
		t.Fatalf("expected github identity to be inserted, got %d", count)
	}
	if count := scalarInt(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'auth.login.oauth.success'"); count != 1 {
		t.Fatalf("expected one oauth success audit, got %d", count)
	}

	replay := rawRequest(t, client, http.MethodGet, callbackURL, "")
	defer replay.Body.Close()
	if replay.StatusCode != http.StatusFound {
		t.Fatalf("expected replay callback to redirect with error, got %d", replay.StatusCode)
	}
	if loc := replay.Header.Get("Location"); !containsAll(loc, "/login", "oauth_error=oauth_state_invalid") {
		t.Fatalf("expected state_invalid replay redirect, got %q", loc)
	}
}

func TestOAuth_GoogleExistingTOTPRequiresChallenge(t *testing.T) {
	suite.SetupTest(t)
	defer suite.TeardownTest(t)
	fake := newFakeOAuthProvider(t)
	email := "oauth-totp@example.com"
	userID, secret, _ := setupAndEnableTOTP(t, email)
	fake.setProfile(fakeOAuthProfile{ID: "google-sub-2002", Email: email, Name: "OAuth TOTP", Avatar: "https://example.com/g.png", Verified: true})
	setOAuthProviderConfig(t, "google", fake, true)

	client := newNoRedirectHTTPClient(t)
	callbackResp, _ := runOAuthBrowserFlow(t, client, "google", "/security")
	defer callbackResp.Body.Close()
	if callbackResp.StatusCode != http.StatusFound {
		t.Fatalf("expected TOTP callback redirect, got %d", callbackResp.StatusCode)
	}
	loc := callbackResp.Header.Get("Location")
	if !containsAll(loc, "/verify-totp", "challenge_token=", "from=%2Fsecurity") {
		t.Fatalf("expected verify-totp redirect with challenge, got %q", loc)
	}
	challenge := queryParam(t, loc, "challenge_token")
	if challenge == "" {
		t.Fatal("expected challenge token")
	}
	if session := rawRequest(t, client, http.MethodGet, suite.Client.BaseURL()+"/api/v1/auth/session", ""); session.StatusCode != http.StatusUnauthorized {
		session.Body.Close()
		t.Fatalf("expected no session before TOTP verify, got %d", session.StatusCode)
	} else {
		session.Body.Close()
	}

	invalid := rawRequest(t, client, http.MethodPost, suite.Client.BaseURL()+"/api/v1/auth/verify-totp", fmt.Sprintf(`{"challenge_token":"%s","code":"000000"}`, challenge))
	invalid.Body.Close()
	if invalid.StatusCode != http.StatusForbidden {
		t.Fatalf("expected invalid challenge TOTP to be 403, got %d", invalid.StatusCode)
	}

	code, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate TOTP code: %v", err)
	}
	verify := rawRequest(t, client, http.MethodPost, suite.Client.BaseURL()+"/api/v1/auth/verify-totp", fmt.Sprintf(`{"challenge_token":"%s","code":"%s"}`, challenge, code))
	defer verify.Body.Close()
	if verify.StatusCode != http.StatusOK {
		body := readBody(t, verify)
		t.Fatalf("expected challenge verify success, got %d body=%s", verify.StatusCode, body)
	}
	assertRawSessionOK(t, client, email)
	if count := scalarInt(t, "SELECT COUNT(*) FROM user_identities WHERE user_id = $1 AND provider = 'google' AND auth_id = 'google-sub-2002'", userID); count != 1 {
		t.Fatalf("expected google identity linked after TOTP success, got %d", count)
	}

	replayCode, err := ptotp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate replay TOTP code: %v", err)
	}
	replay := rawRequest(t, client, http.MethodPost, suite.Client.BaseURL()+"/api/v1/auth/verify-totp", fmt.Sprintf(`{"challenge_token":"%s","code":"%s"}`, challenge, replayCode))
	replay.Body.Close()
	if replay.StatusCode == http.StatusOK {
		t.Fatal("expected consumed challenge replay to fail")
	}
}

func setOAuthProviderConfig(t *testing.T, provider string, fake *fakeOAuthProvider, enabled bool) {
	t.Helper()
	setConfig(t, "web.baseUrl", suite.Client.BaseURL(), "string")
	setConfig(t, "oauth."+provider+".enabled", fmt.Sprintf("%t", enabled), "bool")
	setConfig(t, "oauth."+provider+".clientId", "test-client", "string")
	setConfig(t, "oauth."+provider+".clientSecret", "test-secret", "secret")
	setConfig(t, "oauth."+provider+".redirectUri", "", "string")
	if provider == "github" {
		setConfig(t, "oauth.github.scopes", "user:email", "string")
	} else {
		setConfig(t, "oauth.google.scopes", "openid email profile", "string")
	}
	if fake == nil {
		return
	}
	setConfig(t, "oauth."+provider+".authUrl", fake.server.URL+"/authorize", "string")
	setConfig(t, "oauth."+provider+".tokenUrl", fake.server.URL+"/token", "string")
	if provider == "github" {
		setConfig(t, "oauth.github.userUrl", fake.server.URL+"/github/user", "string")
		setConfig(t, "oauth.github.emailsUrl", fake.server.URL+"/github/emails", "string")
	} else {
		setConfig(t, "oauth.google.userInfoUrl", fake.server.URL+"/google/userinfo", "string")
	}
}

func setConfig(t *testing.T, key, value, valueType string) {
	t.Helper()
	if err := service.Config().Set(context.Background(), &service.ConfigSetParams{Key: key, Value: value, ValueType: valueType, Description: "test config"}); err != nil {
		t.Fatalf("set config %s: %v", key, err)
	}
}

func newNoRedirectHTTPClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}
	return &http.Client{Jar: jar, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
}

func runOAuthBrowserFlow(t *testing.T, client *http.Client, provider, redirectPath string) (*http.Response, string) {
	t.Helper()
	start := rawRequest(t, client, http.MethodGet, suite.Client.BaseURL()+"/api/v1/auth/oauth/"+provider+"?redirect="+url.QueryEscape(redirectPath), "")
	start.Body.Close()
	if start.StatusCode != http.StatusFound {
		t.Fatalf("start OAuth: expected 302, got %d", start.StatusCode)
	}
	authURL := start.Header.Get("Location")
	if authURL == "" {
		t.Fatal("start OAuth missing Location")
	}
	auth := rawRequest(t, client, http.MethodGet, authURL, "")
	if auth.StatusCode != http.StatusFound {
		body := readBody(t, auth)
		auth.Body.Close()
		t.Fatalf("fake authorize: expected 302, got %d url=%s body=%s", auth.StatusCode, authURL, body)
	}
	auth.Body.Close()
	callbackURL := auth.Header.Get("Location")
	if callbackURL == "" {
		t.Fatal("fake authorize missing callback Location")
	}
	callback := rawRequest(t, client, http.MethodGet, callbackURL, "")
	return callback, callbackURL
}

func rawRequest(t *testing.T, client *http.Client, method, rawURL, body string) *http.Response {
	t.Helper()
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		t.Fatalf("create raw request %s %s: %v", method, rawURL, err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("raw request %s %s: %v", method, rawURL, err)
	}
	return resp
}

func assertRawSessionOK(t *testing.T, client *http.Client, email string) {
	t.Helper()
	resp := rawRequest(t, client, http.MethodGet, suite.Client.BaseURL()+"/api/v1/auth/session", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body := readBody(t, resp)
		t.Fatalf("expected session OK, got %d body=%s", resp.StatusCode, body)
	}
	body := readBody(t, resp)
	var envelope map[string]any
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("parse session JSON: %v body=%s", err, body)
	}
	data := envelope["data"].(map[string]any)
	user := data["user"].(map[string]any)
	if user["email"] != email {
		t.Fatalf("expected session email %s, got %v", email, user["email"])
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	return buf.String()
}

func queryParam(t *testing.T, rawURL, key string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse URL %q: %v", rawURL, err)
	}
	return parsed.Query().Get(key)
}

func containsAll(value string, needles ...string) bool {
	for _, needle := range needles {
		if !stringsContains(value, needle) {
			return false
		}
	}
	return true
}

func stringsContains(value, needle string) bool {
	return len(needle) == 0 || bytes.Contains([]byte(value), []byte(needle))
}
