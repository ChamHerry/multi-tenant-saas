package service

import (
	"context"
	"net/http"
	"time"
)

type AuthSession struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	UserAgent     string     `json:"user_agent,omitempty"`
	IP            string     `json:"ip,omitempty"`
	LastUsedAt    time.Time  `json:"last_used_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
	IdleExpiresAt time.Time  `json:"idle_expires_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AuthSessionCookies struct {
	Session *http.Cookie
	CSRF    *http.Cookie
}

type IAuthSession interface {
	Create(ctx context.Context, userID, userAgent, ip string) (*AuthSession, *AuthSessionCookies, error)
	Authenticate(ctx context.Context, cookieValue string) (*AuthIdentity, error)
	Get(ctx context.Context, sessionID string) (*AuthSession, error)
	ValidateCSRF(ctx context.Context, sessionID, token string) error
	Revoke(ctx context.Context, sessionID, reason string) error
	RevokeUserSessions(ctx context.Context, userID, exceptSessionID, reason string) error
	SessionCookieName(ctx context.Context) string
	CSRFCookieName(ctx context.Context) string
	ClearCookies(ctx context.Context) *AuthSessionCookies
}

var localAuthSession IAuthSession

func AuthSessionService() IAuthSession {
	if localAuthSession == nil {
		panic("implement not found for interface IAuthSession")
	}
	return localAuthSession
}

func RegisterAuthSession(i IAuthSession) {
	localAuthSession = i
}
