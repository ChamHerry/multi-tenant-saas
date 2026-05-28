package authsession

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
	"repomind-temp/utility/uuid"
)

const (
	defaultSessionCookieName = "repomind_session"
	defaultCSRFCookieName    = "repomind_csrf"
	defaultAbsoluteTTL       = 7 * 24 * time.Hour
	defaultIdleTTL           = 12 * time.Hour
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type sAuthSession struct{}

func init() {
	service.RegisterAuthSession(&sAuthSession{})
}

func (s *sAuthSession) Create(ctx context.Context, userID, userAgent, ip string) (*service.AuthSession, *service.AuthSessionCookies, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, nil, err
	}
	secret, err := sessionSecret(ctx)
	if err != nil {
		return nil, nil, err
	}
	sessionID := uuid.GenerateV4()
	if err = validateInternalID(sessionID); err != nil {
		return nil, nil, err
	}
	secretToken, err := randomToken()
	if err != nil {
		return nil, nil, err
	}
	csrfToken, err := randomToken()
	if err != nil {
		return nil, nil, err
	}
	absoluteTTL := durationConfig(ctx, "auth.session.absoluteTTL", defaultAbsoluteTTL)
	idleTTL := durationConfig(ctx, "auth.session.idleTTL", defaultIdleTTL)
	now := time.Now().UTC()
	expiresAt := now.Add(absoluteTTL)
	idleExpiresAt := now.Add(idleTTL)
	if idleExpiresAt.After(expiresAt) {
		idleExpiresAt = expiresAt
	}
	record, err := g.DB().GetOne(ctx, `
INSERT INTO public.auth_sessions(id, user_id, secret_hash, csrf_hash, user_agent, ip, last_used_at, expires_at, idle_expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, '')::inet, now(), ?, ?, now(), now())
RETURNING id, user_id, user_agent, ip, last_used_at, expires_at, idle_expires_at, revoked_at, created_at, updated_at`,
		sessionID, userID, hashToken(secretToken, secret), hashToken(csrfToken, secret), trimForDB(userAgent), normalizeIP(ip), expiresAt, idleExpiresAt)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "insert auth session")
	}
	session, err := mapSession(record)
	if err != nil {
		return nil, nil, err
	}
	cookies := s.buildCookies(ctx, sessionID+"."+secretToken, csrfToken, expiresAt)
	return session, cookies, nil
}

func (s *sAuthSession) Authenticate(ctx context.Context, cookieValue string) (*service.AuthIdentity, error) {
	sessionID, secretToken, err := parseSessionCookie(cookieValue)
	if err != nil {
		return nil, err
	}
	secret, err := sessionSecret(ctx)
	if err != nil {
		return nil, err
	}
	record, err := g.DB().GetOne(ctx, `
SELECT s.id, s.user_id
FROM public.auth_sessions s
JOIN public.users u ON u.id = s.user_id
WHERE s.id=?
  AND s.secret_hash=?
  AND s.revoked_at IS NULL
  AND s.expires_at > now()
  AND s.idle_expires_at > now()
  AND u.status='active'
  AND u.deleted_at IS NULL
LIMIT 1`, sessionID, hashToken(secretToken, secret))
	if err != nil {
		return nil, gerror.Wrap(err, "select auth session")
	}
	if record.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "session is invalid, expired, or revoked")
	}
	if err = s.touch(ctx, sessionID); err != nil {
		return nil, err
	}
	return &service.AuthIdentity{UserID: record["user_id"].String(), Type: "session", SessionID: sessionID}, nil
}

func (s *sAuthSession) Get(ctx context.Context, sessionID string) (*service.AuthSession, error) {
	if err := validateInternalID(sessionID); err != nil {
		return nil, err
	}
	record, err := g.DB().GetOne(ctx, `
SELECT id, user_id, user_agent, ip, last_used_at, expires_at, idle_expires_at, revoked_at, created_at, updated_at
FROM public.auth_sessions
WHERE id=?`, sessionID)
	if err != nil {
		return nil, gerror.Wrap(err, "select auth session")
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("auth session %s not found", sessionID)
	}
	return mapSession(record)
}

func (s *sAuthSession) ValidateCSRF(ctx context.Context, sessionID, token string) error {
	if err := validateInternalID(sessionID); err != nil {
		return err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return gerror.NewCode(gcode.CodeNotAuthorized, "csrf token is required")
	}
	secret, err := sessionSecret(ctx)
	if err != nil {
		return err
	}
	record, err := g.DB().GetOne(ctx, `
SELECT csrf_hash
FROM public.auth_sessions
WHERE id=?
  AND revoked_at IS NULL
  AND expires_at > now()
  AND idle_expires_at > now()
LIMIT 1`, sessionID)
	if err != nil {
		return gerror.Wrap(err, "select csrf hash")
	}
	if record.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotAuthorized, "session is invalid, expired, or revoked")
	}
	expected := record["csrf_hash"].String()
	actual := hashToken(token, secret)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) != 1 {
		return gerror.NewCode(gcode.CodeNotAuthorized, "csrf token is invalid")
	}
	return nil
}

func (s *sAuthSession) Revoke(ctx context.Context, sessionID, reason string) error {
	if err := validateInternalID(sessionID); err != nil {
		return err
	}
	_, err := g.DB().Exec(ctx, `
UPDATE public.auth_sessions
SET revoked_at=COALESCE(revoked_at, now()), revoke_reason=COALESCE(NULLIF(?, ''), revoke_reason), updated_at=now()
WHERE id=?`, reason, sessionID)
	return gerror.Wrap(err, "revoke auth session")
}

func (s *sAuthSession) RevokeUserSessions(ctx context.Context, userID, exceptSessionID, reason string) error {
	if err := validateInternalID(userID); err != nil {
		return err
	}
	if exceptSessionID != "" {
		if err := validateInternalID(exceptSessionID); err != nil {
			return err
		}
	}
	_, err := g.DB().Exec(ctx, `
UPDATE public.auth_sessions
SET revoked_at=COALESCE(revoked_at, now()), revoke_reason=COALESCE(NULLIF(?, ''), revoke_reason), updated_at=now()
WHERE user_id=? AND revoked_at IS NULL AND id <> COALESCE(NULLIF(?, '')::uuid, '00000000-0000-0000-0000-000000000000'::uuid)`, reason, userID, exceptSessionID)
	return gerror.Wrap(err, "revoke user auth sessions")
}

func (s *sAuthSession) SessionCookieName(ctx context.Context) string {
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "auth.session.cookie.name", defaultSessionCookieName).String())
}

func (s *sAuthSession) CSRFCookieName(ctx context.Context) string {
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "auth.session.cookie.csrfName", defaultCSRFCookieName).String())
}

func (s *sAuthSession) ClearCookies(ctx context.Context) *service.AuthSessionCookies {
	attrs := cookieAttrs(ctx)
	expires := time.Unix(1, 0).UTC()
	return &service.AuthSessionCookies{
		Session: &http.Cookie{Name: s.SessionCookieName(ctx), Value: "", Path: attrs.Path, Domain: attrs.Domain, MaxAge: -1, Expires: expires, HttpOnly: true, Secure: attrs.Secure, SameSite: attrs.SameSite},
		CSRF:    &http.Cookie{Name: s.CSRFCookieName(ctx), Value: "", Path: attrs.Path, Domain: attrs.Domain, MaxAge: -1, Expires: expires, HttpOnly: false, Secure: attrs.Secure, SameSite: attrs.SameSite},
	}
}

func (s *sAuthSession) touch(ctx context.Context, sessionID string) error {
	idleTTL := durationConfig(ctx, "auth.session.idleTTL", defaultIdleTTL)
	_, err := g.DB().Exec(ctx, `
UPDATE public.auth_sessions
SET last_used_at=now(),
    idle_expires_at=LEAST(expires_at, now() + (?::interval)),
    updated_at=now()
WHERE id=?`, pgInterval(idleTTL), sessionID)
	return gerror.Wrap(err, "touch auth session")
}

func (s *sAuthSession) buildCookies(ctx context.Context, sessionValue, csrfValue string, expiresAt time.Time) *service.AuthSessionCookies {
	attrs := cookieAttrs(ctx)
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	return &service.AuthSessionCookies{
		Session: &http.Cookie{Name: s.SessionCookieName(ctx), Value: sessionValue, Path: attrs.Path, Domain: attrs.Domain, MaxAge: maxAge, Expires: expiresAt, HttpOnly: true, Secure: attrs.Secure, SameSite: attrs.SameSite},
		CSRF:    &http.Cookie{Name: s.CSRFCookieName(ctx), Value: csrfValue, Path: attrs.Path, Domain: attrs.Domain, MaxAge: maxAge, Expires: expiresAt, HttpOnly: false, Secure: attrs.Secure, SameSite: attrs.SameSite},
	}
}

type cookieConfig struct {
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

func cookieAttrs(ctx context.Context) cookieConfig {
	path := strings.TrimSpace(g.Cfg().MustGet(ctx, "auth.session.cookie.path", "/").String())
	if path == "" {
		path = "/"
	}
	domain := strings.TrimSpace(g.Cfg().MustGet(ctx, "auth.session.cookie.domain", "").String())
	secureDefault := !isLocalEnv(ctx)
	return cookieConfig{
		Path:     path,
		Domain:   domain,
		Secure:   g.Cfg().MustGet(ctx, "auth.session.cookie.secure", secureDefault).Bool(),
		SameSite: http.SameSiteLaxMode,
	}
}

func parseSessionCookie(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", gerror.NewCode(gcode.CodeNotAuthorized, "session cookie is invalid")
	}
	if err := validateInternalID(parts[0]); err != nil {
		return "", "", gerror.NewCode(gcode.CodeNotAuthorized, "session cookie is invalid")
	}
	return parts[0], parts[1], nil
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", gerror.Wrap(err, "generate random token")
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func sessionSecret(ctx context.Context) (string, error) {
	secret := strings.TrimSpace(g.Cfg().MustGet(ctx, "auth.session.secret", "").String())
	if secret == "" {
		return "", gerror.NewCode(gcode.CodeMissingConfiguration, "auth.session.secret is required")
	}
	return secret, nil
}

func durationConfig(ctx context.Context, key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(g.Cfg().MustGet(ctx, key, fallback.String()).String())
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func pgInterval(d time.Duration) string {
	seconds := int64(d.Seconds())
	if seconds <= 0 {
		seconds = int64(defaultIdleTTL.Seconds())
	}
	return strconv.FormatInt(seconds, 10) + " seconds"
}

func normalizeIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	if parsed := net.ParseIP(ip); parsed != nil {
		return parsed.String()
	}
	return ""
}

func trimForDB(value string) string {
	return strings.TrimSpace(value)
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.Newf("invalid internal uuid %q", id)
	}
	return nil
}

func mapSession(record gdb.Record) (*service.AuthSession, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	userID := record["user_id"].String()
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	return &service.AuthSession{
		ID:            id,
		UserID:        userID,
		UserAgent:     record["user_agent"].String(),
		IP:            record["ip"].String(),
		LastUsedAt:    record["last_used_at"].Time(),
		ExpiresAt:     record["expires_at"].Time(),
		IdleExpiresAt: record["idle_expires_at"].Time(),
		RevokedAt:     nullableTime(record["revoked_at"]),
		CreatedAt:     record["created_at"].Time(),
		UpdatedAt:     record["updated_at"].Time(),
	}, nil
}

func nullableTime(value any) *time.Time {
	v, ok := value.(interface {
		IsNil() bool
		Time(...string) time.Time
	})
	if !ok || v.IsNil() {
		return nil
	}
	t := v.Time()
	return &t
}

func isLocalEnv(ctx context.Context) bool {
	env := strings.ToLower(strings.TrimSpace(g.Cfg().MustGet(ctx, "server.env", "local").String()))
	return env == "local" || env == "test"
}
