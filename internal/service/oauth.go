package service

import (
	"context"
	"time"
)

type OAuthProviderItem struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
}

type OAuthUserInfo struct {
	Provider      string
	AuthID        string
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
	RawProfile    map[string]any
}

type OAuthCallbackInput struct {
	Provider         string
	Code             string
	State            string
	Error            string
	ErrorDescription string
	IP               string
	UserAgent        string
}

type OAuthLoginResult struct {
	User           *User               `json:"user"`
	Session        *AuthSession        `json:"session,omitempty"`
	Cookies        *AuthSessionCookies `json:"-"`
	Requires2FA    bool                `json:"requires_2fa,omitempty"`
	ChallengeToken string              `json:"challenge_token,omitempty"`
	RedirectURI    string              `json:"redirect_uri,omitempty"`
}

type OAuthStartResult struct {
	State   string
	AuthURL string
}

type OAuthChallenge struct {
	UserID        string
	Provider      string
	AuthID        string
	LoginKey      string
	RedirectURI   string
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
	RawProfile    map[string]any
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

type IOAuth interface {
	GetEnabledProviders(ctx context.Context) ([]OAuthProviderItem, error)
	CreateState(ctx context.Context, provider, redirectURI string) (*OAuthStartResult, error)
	HandleCallback(ctx context.Context, in OAuthCallbackInput) (*OAuthLoginResult, error)
	VerifyChallenge(ctx context.Context, challengeToken, code, userAgent, ip string) (*OAuthLoginResult, error)
	CleanupExpiredStates(ctx context.Context) (int64, error)
}

var localOAuth IOAuth

func OAuth() IOAuth {
	if localOAuth == nil {
		panic("implement not found for interface IOAuth")
	}
	return localOAuth
}

func RegisterOAuth(i IOAuth) {
	localOAuth = i
}
