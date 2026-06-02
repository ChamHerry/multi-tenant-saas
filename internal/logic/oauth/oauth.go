package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	goauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

const (
	providerGitHub = "github"
	providerGoogle = "google"

	defaultStateTTL     = 10 * time.Minute
	defaultChallengeTTL = 5 * time.Minute

	githubDefaultUserURL   = "https://api.github.com/user"
	githubDefaultEmailsURL = "https://api.github.com/user/emails"
	googleDefaultUserInfo  = "https://openidconnect.googleapis.com/v1/userinfo"
)

type sOAuth struct{}

type providerDefinition struct {
	Provider string
	Name     string
	Endpoint goauth2.Endpoint
	Scopes   []string
}

type providerConfig struct {
	providerDefinition
	Enabled      bool
	ClientID     string
	ClientSecret string
	RedirectURL  string
	UserURL      string
	EmailsURL    string
	UserInfoURL  string
}

type identityRecord struct {
	ID     string
	UserID string
	Email  string
}

type challengeMetadata struct {
	Email         string         `json:"email"`
	EmailVerified bool           `json:"email_verified"`
	DisplayName   string         `json:"display_name"`
	AvatarURL     string         `json:"avatar_url"`
	RawProfile    map[string]any `json:"raw_profile"`
}

func init() {
	service.RegisterOAuth(&sOAuth{})
}

func (s *sOAuth) GetEnabledProviders(ctx context.Context) ([]service.OAuthProviderItem, error) {
	providers := make([]service.OAuthProviderItem, 0, len(providerOrder()))
	for _, def := range providerOrder() {
		cfg := s.loadProviderConfig(ctx, def.Provider)
		providers = append(providers, service.OAuthProviderItem{
			Provider: def.Provider,
			Name:     def.Name,
			Enabled:  cfg.Enabled && cfg.ClientID != "" && cfg.ClientSecret != "",
		})
	}
	return providers, nil
}

func (s *sOAuth) CreateState(ctx context.Context, provider, redirectURI string) (*service.OAuthStartResult, error) {
	cfg, err := s.requireProviderConfig(ctx, provider)
	if err != nil {
		return nil, err
	}
	redirectPath, err := normalizeRedirectPath(redirectURI)
	if err != nil {
		return nil, err
	}
	state, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	if cfg.RedirectURL == "" {
		cfg.RedirectURL = callbackURLFromRequest(ctx, provider)
	}
	if cfg.RedirectURL == "" {
		return nil, gerror.NewCode(gcode.CodeMissingConfiguration, "OAuth redirect URI is required")
	}
	metadata, _ := json.Marshal(map[string]any{"request_ip": service.BizCtx().GetClientIP(ctx)})
	expiresAt := time.Now().UTC().Add(stateTTL(ctx))
	_, err = g.DB().Exec(ctx, `
		INSERT INTO oauth_states ("state", provider, redirect_uri, "metadata", expires_at, created_at)
		VALUES ($1, $2, $3, CAST($4 AS jsonb), $5, now())`,
		state, provider, redirectPath, string(metadata), expiresAt)
	if err != nil {
		return nil, gerror.Wrap(err, "create OAuth state")
	}
	_, _ = s.CleanupExpiredStates(ctx)
	oauthCfg := cfg.oauth2Config()
	return &service.OAuthStartResult{
		State:   state,
		AuthURL: oauthCfg.AuthCodeURL(state),
	}, nil
}

func (s *sOAuth) HandleCallback(ctx context.Context, in service.OAuthCallbackInput) (*service.OAuthLoginResult, error) {
	if in.Error != "" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "oauth_denied")
	}
	if strings.TrimSpace(in.Code) == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "oauth_code_missing")
	}
	cfg, err := s.requireProviderConfig(ctx, in.Provider)
	if err != nil {
		return nil, err
	}
	stateRedirect, err := consumeState(ctx, in.Provider, in.State)
	if err != nil {
		return nil, err
	}
	if cfg.RedirectURL == "" {
		cfg.RedirectURL = callbackURLFromRequest(ctx, in.Provider)
	}
	token, err := cfg.oauth2Config().Exchange(ctx, in.Code)
	if err != nil {
		return nil, gerror.Wrap(err, "oauth_token_exchange_failed")
	}
	userInfo, err := s.fetchUserInfo(ctx, cfg, token)
	if err != nil {
		return nil, err
	}
	user, _, created, err := s.resolveOAuthUser(ctx, userInfo)
	if err != nil {
		return nil, err
	}
	// Newly created OAuth users cannot have TOTP configured yet. Existing users must be gated.
	if !created {
		totpEnabled, err := service.TOTP().IsEnabled(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		if totpEnabled {
			challenge, err := createChallenge(ctx, user, userInfo, stateRedirect)
			if err != nil {
				return nil, err
			}
			_ = service.Audit().Write(ctx, service.AuditLogInput{
				UserID:       user.ID,
				Action:       "auth.login.2fa_required",
				ResourceType: "auth",
				IP:           normalizeIP(in.IP),
				UserAgent:    in.UserAgent,
				Metadata:     map[string]any{"provider": in.Provider, "email": userInfo.Email},
			})
			return &service.OAuthLoginResult{User: user, Requires2FA: true, ChallengeToken: challenge, RedirectURI: stateRedirect}, nil
		}
	}

	session, cookies, err := service.AuthSessionService().Create(ctx, user.ID, in.UserAgent, in.IP)
	if err != nil {
		return nil, err
	}
	if err = s.finalizeOAuthLogin(ctx, user.ID, userInfo, session.ID, in.IP, in.UserAgent); err != nil {
		return nil, err
	}
	user, _ = service.UserService().GetUser(ctx, user.ID)
	return &service.OAuthLoginResult{User: user, Session: session, Cookies: cookies, RedirectURI: stateRedirect}, nil
}

func (s *sOAuth) VerifyChallenge(ctx context.Context, challengeToken, code, userAgent, ip string) (*service.OAuthLoginResult, error) {
	challenge, err := getActiveChallenge(ctx, challengeToken)
	if err != nil {
		return nil, err
	}
	result, err := service.TOTP().ValidateTOTP(ctx, challenge.UserID, code)
	if err != nil {
		return nil, err
	}
	if result == nil || !result.Valid {
		_ = service.Audit().Write(ctx, service.AuditLogInput{
			UserID:       challenge.UserID,
			Action:       "auth.totp.failed",
			ResourceType: "auth",
			Metadata:     map[string]any{"provider": challenge.Provider},
		})
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "Invalid TOTP code")
	}
	if err = consumeChallenge(ctx, challengeToken); err != nil {
		return nil, err
	}
	userInfo := &service.OAuthUserInfo{
		Provider:      challenge.Provider,
		AuthID:        challenge.AuthID,
		Email:         challenge.Email,
		EmailVerified: challenge.EmailVerified,
		DisplayName:   challenge.DisplayName,
		AvatarURL:     challenge.AvatarURL,
		RawProfile:    challenge.RawProfile,
	}
	session, cookies, err := service.AuthSessionService().Create(ctx, challenge.UserID, userAgent, ip)
	if err != nil {
		return nil, err
	}
	if err = s.finalizeOAuthLogin(ctx, challenge.UserID, userInfo, session.ID, ip, userAgent); err != nil {
		return nil, err
	}
	user, err := service.UserService().GetUser(ctx, challenge.UserID)
	if err != nil {
		return nil, err
	}
	action := "auth.totp.verified"
	if result.UsedBackupCode {
		action = "auth.totp.backup_code_used"
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       challenge.UserID,
		Action:       action,
		ResourceType: "auth_session",
		ResourceID:   session.ID,
		Metadata: map[string]any{
			"provider":                  challenge.Provider,
			"backup_code_used":          result.UsedBackupCode,
			"backup_codes_remaining":    result.BackupCodesRemaining,
			"oauth_challenge_completed": true,
		},
	})
	return &service.OAuthLoginResult{User: user, Session: session, Cookies: cookies, RedirectURI: challenge.RedirectURI}, nil
}

func (s *sOAuth) CleanupExpiredStates(ctx context.Context) (int64, error) {
	result, err := g.DB().Exec(ctx, `DELETE FROM oauth_states WHERE expires_at < now() OR (used_at IS NOT NULL AND used_at < now() - INTERVAL '1 hour')`)
	if err != nil {
		return 0, gerror.Wrap(err, "cleanup OAuth states")
	}
	_, _ = g.DB().Exec(ctx, `DELETE FROM auth_login_challenges WHERE expires_at < now() OR (consumed_at IS NOT NULL AND consumed_at < now() - INTERVAL '1 hour')`)
	rows, _ := result.RowsAffected()
	return rows, nil
}

func (s *sOAuth) resolveOAuthUser(ctx context.Context, info *service.OAuthUserInfo) (*service.User, *identityRecord, bool, error) {
	identity, err := findIdentityByProviderAuthID(ctx, info.Provider, info.AuthID)
	if err != nil {
		return nil, nil, false, err
	}
	if identity != nil {
		user, err := service.UserService().GetUser(ctx, identity.UserID)
		return user, identity, false, err
	}
	if info.Email == "" {
		return nil, nil, false, gerror.NewCode(gcode.CodeInvalidParameter, "oauth_email_missing")
	}
	passwordIdentity, err := service.UserService().GetIdentityByEmail(ctx, "password", info.Email)
	if err != nil {
		return nil, nil, false, err
	}
	if passwordIdentity != nil {
		user, err := service.UserService().GetUser(ctx, passwordIdentity.UserID)
		return user, &identityRecord{UserID: passwordIdentity.UserID, Email: passwordIdentity.Email}, false, err
	}
	user, err := service.UserService().EnsureUserByIdentity(ctx, service.EnsureUserByIdentityInput{
		Provider:      info.Provider,
		AuthID:        info.AuthID,
		Email:         info.Email,
		EmailVerified: info.EmailVerified,
		DisplayName:   info.DisplayName,
		AvatarURL:     info.AvatarURL,
		RawProfile:    info.RawProfile,
		Metadata:      map[string]any{"source": "oauth", "provider": info.Provider},
	})
	return user, nil, true, err
}

func (s *sOAuth) finalizeOAuthLogin(ctx context.Context, userID string, info *service.OAuthUserInfo, sessionID, ip, userAgent string) error {
	rawProfile, err := marshalJSON(defaultMap(info.RawProfile), "marshal OAuth raw profile")
	if err != nil {
		return err
	}
	identCols := dao.UserIdentities.Columns()
	identData := g.Map{
		identCols.RawProfile:  rawProfile,
		identCols.LastLoginAt: "now()",
		identCols.UpdatedAt:   "now()",
	}
	if info.Email != "" {
		identData[identCols.Email] = info.Email
	}
	if info.EmailVerified {
		identData[identCols.EmailVerified] = true
	}
	result, err := dao.UserIdentities.Ctx(ctx).
		Where(identCols.Provider, info.Provider).
		Where(identCols.AuthId, info.AuthID).
		Data(identData).
		Update()
	if err != nil {
		return gerror.Wrap(err, "touch OAuth identity login")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		identityID := uuid.GenerateV4()
		if _, err = dao.UserIdentities.Ctx(ctx).Data(g.Map{
			identCols.Id:            identityID,
			identCols.UserId:        userID,
			identCols.Provider:      info.Provider,
			identCols.AuthId:        info.AuthID,
			identCols.Email:         nullableString(info.Email),
			identCols.EmailVerified: info.EmailVerified,
			identCols.RawProfile:    rawProfile,
			identCols.LastLoginAt:   "now()",
			identCols.CreatedAt:     "now()",
			identCols.UpdatedAt:     "now()",
		}).Insert(); err != nil {
			return gerror.Wrap(err, "insert OAuth identity login")
		}
	}

	userCols := dao.Users.Columns()
	userData := g.Map{userCols.LastLoginAt: "now()", userCols.UpdatedAt: "now()"}
	if info.Email != "" {
		userData[userCols.Email] = info.Email
	}
	if info.DisplayName != "" {
		userData[userCols.DisplayName] = info.DisplayName
	}
	if info.AvatarURL != "" {
		userData[userCols.AvatarUrl] = info.AvatarURL
	}
	if _, err = dao.Users.Ctx(ctx).
		Where(userCols.Id, userID).
		Where("deleted_at IS NULL").
		Data(userData).
		Update(); err != nil {
		return gerror.Wrap(err, "touch OAuth user login")
	}

	return service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.login.oauth.success",
		ResourceType: "auth_session",
		ResourceID:   sessionID,
		IP:           normalizeIP(ip),
		UserAgent:    userAgent,
		Metadata: map[string]any{
			"provider": info.Provider,
			"auth_id":  info.AuthID,
			"email":    info.Email,
		},
	})
}

func (s *sOAuth) fetchUserInfo(ctx context.Context, cfg providerConfig, token *goauth2.Token) (*service.OAuthUserInfo, error) {
	switch cfg.Provider {
	case providerGitHub:
		return fetchGitHubUser(ctx, cfg, token)
	case providerGoogle:
		return fetchGoogleUser(ctx, cfg, token)
	default:
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "unsupported OAuth provider")
	}
}

func fetchGitHubUser(ctx context.Context, cfg providerConfig, token *goauth2.Token) (*service.OAuthUserInfo, error) {
	client := cfg.oauth2Config().Client(ctx, token)
	var profile struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := getJSON(ctx, client, firstNonEmpty(cfg.UserURL, githubDefaultUserURL), &profile); err != nil {
		return nil, gerror.Wrap(err, "oauth_github_profile_failed")
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := getJSON(ctx, client, firstNonEmpty(cfg.EmailsURL, githubDefaultEmailsURL), &emails); err != nil {
		return nil, gerror.Wrap(err, "oauth_github_email_failed")
	}
	email := ""
	for _, item := range emails {
		if item.Verified && item.Primary && item.Email != "" {
			email = item.Email
			break
		}
	}
	if email == "" {
		for _, item := range emails {
			if item.Verified && item.Email != "" {
				email = item.Email
				break
			}
		}
	}
	if email == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "oauth_verified_email_missing")
	}
	displayName := strings.TrimSpace(profile.Name)
	if displayName == "" {
		displayName = profile.Login
	}
	raw := map[string]any{"profile": profile, "emails": emails}
	return &service.OAuthUserInfo{
		Provider:      providerGitHub,
		AuthID:        fmt.Sprintf("%d", profile.ID),
		Email:         strings.ToLower(email),
		EmailVerified: true,
		DisplayName:   displayName,
		AvatarURL:     profile.AvatarURL,
		RawProfile:    raw,
	}, nil
}

func fetchGoogleUser(ctx context.Context, cfg providerConfig, token *goauth2.Token) (*service.OAuthUserInfo, error) {
	client := cfg.oauth2Config().Client(ctx, token)
	var profile struct {
		Sub           string `json:"sub"`
		ID            string `json:"id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := getJSON(ctx, client, firstNonEmpty(cfg.UserInfoURL, googleDefaultUserInfo), &profile); err != nil {
		return nil, gerror.Wrap(err, "oauth_google_profile_failed")
	}
	authID := firstNonEmpty(profile.Sub, profile.ID)
	if authID == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "oauth_auth_id_missing")
	}
	verified := profile.EmailVerified || profile.VerifiedEmail
	if !verified || strings.TrimSpace(profile.Email) == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "oauth_verified_email_missing")
	}
	raw := map[string]any{"profile": profile}
	return &service.OAuthUserInfo{
		Provider:      providerGoogle,
		AuthID:        authID,
		Email:         strings.ToLower(strings.TrimSpace(profile.Email)),
		EmailVerified: true,
		DisplayName:   strings.TrimSpace(profile.Name),
		AvatarURL:     profile.Picture,
		RawProfile:    raw,
	}, nil
}

func getJSON(ctx context.Context, client *http.Client, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("GET %s returned %d: %s", endpoint, resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func findIdentityByProviderAuthID(ctx context.Context, provider, authID string) (*identityRecord, error) {
	identCols := dao.UserIdentities.Columns()
	record, err := dao.UserIdentities.Ctx(ctx).
		Fields(identCols.Id, identCols.UserId, identCols.Email).
		Where(identCols.Provider, provider).
		Where(identCols.AuthId, authID).
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select OAuth identity")
	}
	if record.IsEmpty() {
		return nil, nil
	}
	return &identityRecord{ID: record[identCols.Id].String(), UserID: record[identCols.UserId].String(), Email: record[identCols.Email].String()}, nil
}

func consumeState(ctx context.Context, provider, state string) (string, error) {
	if strings.TrimSpace(state) == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, "oauth_state_missing")
	}
	record, err := g.DB().GetOne(ctx, `
		UPDATE oauth_states
		SET used_at = now()
		WHERE "state" = $1 AND provider = $2 AND used_at IS NULL AND expires_at > now()
		RETURNING redirect_uri`, state, provider)
	if err != nil {
		return "", gerror.Wrap(err, "consume OAuth state")
	}
	if record.IsEmpty() {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "oauth_state_invalid")
	}
	redirectURI := record["redirect_uri"].String()
	if redirectURI == "" {
		redirectURI = "/"
	}
	return redirectURI, nil
}

func createChallenge(ctx context.Context, user *service.User, info *service.OAuthUserInfo, redirectURI string) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	metadata, err := json.Marshal(challengeMetadata{
		Email:         info.Email,
		EmailVerified: info.EmailVerified,
		DisplayName:   info.DisplayName,
		AvatarURL:     info.AvatarURL,
		RawProfile:    info.RawProfile,
	})
	if err != nil {
		return "", gerror.Wrap(err, "marshal OAuth challenge metadata")
	}
	expiresAt := time.Now().UTC().Add(challengeTTL(ctx))
	_, err = g.DB().Exec(ctx, `
		INSERT INTO auth_login_challenges (challenge_hash, user_id, provider, auth_id, login_key, redirect_uri, metadata, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CAST($7 AS jsonb), $8, now())`,
		hashToken(token), user.ID, info.Provider, info.AuthID, info.Email, redirectURI, string(metadata), expiresAt)
	if err != nil {
		return "", gerror.Wrap(err, "create OAuth login challenge")
	}
	return token, nil
}

func getActiveChallenge(ctx context.Context, token string) (*service.OAuthChallenge, error) {
	if strings.TrimSpace(token) == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "oauth_challenge_missing")
	}
	record, err := g.DB().GetOne(ctx, `
		SELECT user_id, provider, auth_id, login_key, redirect_uri, metadata, created_at, expires_at
		FROM auth_login_challenges
		WHERE challenge_hash = $1 AND consumed_at IS NULL AND expires_at > now()`, hashToken(token))
	if err != nil {
		return nil, gerror.Wrap(err, "select OAuth login challenge")
	}
	if record.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "oauth_challenge_invalid")
	}
	meta := challengeMetadata{RawProfile: map[string]any{}}
	if raw := record["metadata"].String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &meta)
	}
	return &service.OAuthChallenge{
		UserID:        record["user_id"].String(),
		Provider:      record["provider"].String(),
		AuthID:        record["auth_id"].String(),
		LoginKey:      record["login_key"].String(),
		RedirectURI:   record["redirect_uri"].String(),
		Email:         meta.Email,
		EmailVerified: meta.EmailVerified,
		DisplayName:   meta.DisplayName,
		AvatarURL:     meta.AvatarURL,
		RawProfile:    defaultMap(meta.RawProfile),
		CreatedAt:     record["created_at"].Time(),
		ExpiresAt:     record["expires_at"].Time(),
	}, nil
}

func consumeChallenge(ctx context.Context, token string) error {
	result, err := g.DB().Exec(ctx, `
		UPDATE auth_login_challenges
		SET consumed_at = now()
		WHERE challenge_hash = $1 AND consumed_at IS NULL AND expires_at > now()`, hashToken(token))
	if err != nil {
		return gerror.Wrap(err, "consume OAuth login challenge")
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return gerror.NewCode(gcode.CodeNotAuthorized, "oauth_challenge_invalid")
	}
	return nil
}

func (s *sOAuth) requireProviderConfig(ctx context.Context, provider string) (providerConfig, error) {
	cfg := s.loadProviderConfig(ctx, provider)
	if cfg.Provider == "" {
		return providerConfig{}, gerror.NewCode(gcode.CodeInvalidParameter, "oauth_provider_invalid")
	}
	if !cfg.Enabled || cfg.ClientID == "" || cfg.ClientSecret == "" {
		return providerConfig{}, gerror.NewCode(gcode.CodeNotAuthorized, "oauth_provider_disabled")
	}
	return cfg, nil
}

func (s *sOAuth) loadProviderConfig(ctx context.Context, provider string) providerConfig {
	def, ok := providerByName(provider)
	if !ok {
		return providerConfig{}
	}
	prefix := "oauth." + provider + "."
	cfg := providerConfig{providerDefinition: def}
	cfg.Enabled = service.Config().GetBool(ctx, prefix+"enabled", false)
	cfg.ClientID = strings.TrimSpace(service.Config().GetString(ctx, prefix+"clientId", ""))
	cfg.ClientSecret = strings.TrimSpace(service.Config().GetString(ctx, prefix+"clientSecret", ""))
	cfg.RedirectURL = strings.TrimSpace(service.Config().GetString(ctx, prefix+"redirectUri", ""))
	cfg.Scopes = splitScopes(service.Config().GetString(ctx, prefix+"scopes", strings.Join(def.Scopes, " ")))
	authURL := strings.TrimSpace(service.Config().GetString(ctx, prefix+"authUrl", ""))
	tokenURL := strings.TrimSpace(service.Config().GetString(ctx, prefix+"tokenUrl", ""))
	if authURL != "" {
		cfg.Endpoint.AuthURL = authURL
	}
	if tokenURL != "" {
		cfg.Endpoint.TokenURL = tokenURL
	}
	switch provider {
	case providerGitHub:
		cfg.UserURL = strings.TrimSpace(service.Config().GetString(ctx, prefix+"userUrl", ""))
		cfg.EmailsURL = strings.TrimSpace(service.Config().GetString(ctx, prefix+"emailsUrl", ""))
	case providerGoogle:
		cfg.UserInfoURL = strings.TrimSpace(service.Config().GetString(ctx, prefix+"userInfoUrl", ""))
	}
	return cfg
}

func (cfg providerConfig) oauth2Config() *goauth2.Config {
	return &goauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     cfg.Endpoint,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
	}
}

func providerOrder() []providerDefinition {
	return []providerDefinition{
		{Provider: providerGitHub, Name: "GitHub", Endpoint: github.Endpoint, Scopes: []string{"user:email"}},
		{Provider: providerGoogle, Name: "Google", Endpoint: google.Endpoint, Scopes: []string{"openid", "email", "profile"}},
	}
}

func providerByName(provider string) (providerDefinition, bool) {
	for _, def := range providerOrder() {
		if def.Provider == provider {
			return def, true
		}
	}
	return providerDefinition{}, false
}

func normalizeRedirectPath(raw string) (string, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return "/", nil
	}
	if strings.HasPrefix(candidate, "//") {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "oauth_redirect_invalid")
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "oauth_redirect_invalid")
	}
	if parsed.Path == "" || !strings.HasPrefix(parsed.Path, "/") {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "oauth_redirect_invalid")
	}
	if parsed.RawQuery != "" {
		return parsed.Path + "?" + parsed.RawQuery, nil
	}
	return parsed.Path, nil
}

func frontendRedirect(ctx context.Context, path string) string {
	normalized, err := normalizeRedirectPath(path)
	if err != nil {
		normalized = "/"
	}
	base := strings.TrimRight(service.Config().GetString(ctx, "web.baseUrl", ""), "/")
	if base == "" {
		return normalized
	}
	return base + normalized
}

func callbackURLFromRequest(ctx context.Context, provider string) string {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil || r.Host == "" {
		return ""
	}
	scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = "http"
		if r.TLS != nil {
			scheme = "https"
		}
	}
	return fmt.Sprintf("%s://%s/api/v1/auth/oauth/%s/callback", scheme, r.Host, provider)
}

func randomToken(byteLen int) (string, error) {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", gerror.Wrap(err, "generate random token")
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func splitScopes(raw string) []string {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return nil
	}
	return parts
}

func stateTTL(ctx context.Context) time.Duration {
	ttl := service.Config().GetDuration(ctx, "oauth.stateTTL", defaultStateTTL)
	if ttl <= 0 {
		return defaultStateTTL
	}
	return ttl
}

func challengeTTL(ctx context.Context) time.Duration {
	ttl := service.Config().GetDuration(ctx, "oauth.challengeTTL", defaultChallengeTTL)
	if ttl <= 0 {
		return defaultChallengeTTL
	}
	return ttl
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func normalizeIP(ip string) string {
	return strings.TrimSpace(ip)
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func defaultMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	return in
}

func marshalJSON(in map[string]any, message string) (string, error) {
	payload, err := json.Marshal(in)
	if err != nil {
		return "", gerror.Wrap(err, message)
	}
	return string(payload), nil
}

// keep gdb imported for generated model type compatibility in future changes and to ensure
// raw record assumptions are explicit at compile time.
var _ gdb.Record

var _ = frontendRedirect
