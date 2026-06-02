package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	authapi "multi-tenant-saas/api/auth"
	"multi-tenant-saas/api/auth/v1"
	"multi-tenant-saas/internal/service"
)

type OAuthControllerV1 struct{}

func NewOAuthV1() authapi.IAuthOAuthV1 {
	return &OAuthControllerV1{}
}

func (c *OAuthControllerV1) OAuthProviders(ctx context.Context, req *v1.OAuthProvidersReq) (res *v1.OAuthProvidersRes, err error) {
	items, err := service.OAuth().GetEnabledProviders(ctx)
	if err != nil {
		return nil, err
	}
	providers := make([]v1.OAuthProviderInfo, 0, len(items))
	for _, item := range items {
		providers = append(providers, v1.OAuthProviderInfo{Provider: item.Provider, Name: item.Name, Enabled: item.Enabled})
	}
	return &v1.OAuthProvidersRes{Providers: providers}, nil
}

func (c *OAuthControllerV1) OAuthLogin(ctx context.Context, req *v1.OAuthLoginReq) (res *v1.OAuthLoginRes, err error) {
	provider := requestProvider(ctx, req.Provider)
	start, err := service.OAuth().CreateState(ctx, provider, req.Redirect)
	if err != nil {
		redirectOAuthError(ctx, oauthErrorCode(err), "")
		return &v1.OAuthLoginRes{}, nil
	}
	redirectTo(ctx, start.AuthURL)
	return &v1.OAuthLoginRes{}, nil
}

func (c *OAuthControllerV1) OAuthCallback(ctx context.Context, req *v1.OAuthCallbackReq) (res *v1.OAuthCallbackRes, err error) {
	provider := requestProvider(ctx, req.Provider)
	result, err := service.OAuth().HandleCallback(ctx, service.OAuthCallbackInput{
		Provider:         provider,
		Code:             req.Code,
		State:            req.State,
		Error:            req.Error,
		ErrorDescription: req.ErrorDescription,
		IP:               service.BizCtx().GetClientIP(ctx),
		UserAgent:        service.BizCtx().GetUserAgent(ctx),
	})
	if err != nil {
		redirectOAuthError(ctx, oauthErrorCode(err), "")
		return &v1.OAuthCallbackRes{}, nil
	}
	if result.Requires2FA {
		redirectTo(ctx, frontendURL(ctx, "/verify-totp", map[string]string{
			"challenge_token": result.ChallengeToken,
			"from":            result.RedirectURI,
			"email":           result.User.Email,
		}))
		return &v1.OAuthCallbackRes{}, nil
	}
	setAuthCookies(ghttp.RequestFromCtx(ctx), result.Cookies)
	redirectTo(ctx, frontendURL(ctx, result.RedirectURI, nil))
	return &v1.OAuthCallbackRes{}, nil
}

func requestProvider(ctx context.Context, provider string) string {
	if trimmed := strings.TrimSpace(provider); trimmed != "" {
		return trimmed
	}
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		return strings.TrimSpace(r.GetRouter("provider").String())
	}
	return ""
}

func redirectTo(ctx context.Context, location string) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return
	}
	r.Response.Status = http.StatusFound
	r.Response.Header().Set("Location", location)
	// Write a tiny body so HandlerResponse does not wrap redirects in JSON.
	r.Response.Write("Found")
}

func redirectOAuthError(ctx context.Context, code, description string) {
	params := map[string]string{"oauth_error": code}
	if description != "" {
		params["oauth_error_description"] = description
	}
	redirectTo(ctx, frontendURL(ctx, "/login", params))
}

func frontendURL(ctx context.Context, path string, params map[string]string) string {
	normalized := normalizeFrontendPath(path)
	parsed, _ := url.Parse(normalized)
	query := parsed.Query()
	for key, value := range params {
		if strings.TrimSpace(value) != "" {
			query.Set(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	base := strings.TrimRight(service.Config().GetString(ctx, "web.baseUrl", ""), "/")
	if base == "" {
		return parsed.String()
	}
	return base + parsed.String()
}

func normalizeFrontendPath(path string) string {
	candidate := strings.TrimSpace(path)
	if candidate == "" || strings.HasPrefix(candidate, "//") {
		return "/"
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Path == "" || !strings.HasPrefix(parsed.Path, "/") {
		return "/"
	}
	if parsed.RawQuery != "" {
		return parsed.Path + "?" + parsed.RawQuery
	}
	return parsed.Path
}

func oauthErrorCode(err error) string {
	text := strings.ToLower(err.Error())
	for _, code := range []string{
		"oauth_denied",
		"oauth_state_invalid",
		"oauth_state_missing",
		"oauth_redirect_invalid",
		"oauth_code_missing",
		"oauth_provider_disabled",
		"oauth_provider_invalid",
		"oauth_verified_email_missing",
		"oauth_email_missing",
		"oauth_auth_id_missing",
		"oauth_token_exchange_failed",
		"oauth_github_profile_failed",
		"oauth_github_email_failed",
		"oauth_google_profile_failed",
	} {
		if strings.Contains(text, code) {
			return code
		}
	}
	return "oauth_failed"
}
