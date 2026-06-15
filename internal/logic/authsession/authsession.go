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
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/do"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

const (
	defaultSessionCookieName = "saas_template_session"
	defaultCSRFCookieName    = "saas_template_csrf"
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

	// Batch-read session configuration (single round-trip via Redis MGET).
	cfg, err := service.Config().GetStrings(ctx, []string{
		"auth.session.secret",
		"auth.session.absoluteTTL",
		"auth.session.idleTTL",
		"auth.session.cookie.name",
		"auth.session.cookie.csrfName",
		"auth.session.cookie.path",
		"auth.session.cookie.domain",
	})
	if err != nil {
		return nil, nil, err
	}

	sessionSecret := strings.TrimSpace(cfg["auth.session.secret"])
	if sessionSecret == "" {
		return nil, nil, gerror.NewCode(gcode.CodeMissingConfiguration, "auth.session.secret is required")
	}

	absoluteTTL, _ := time.ParseDuration(cfg["auth.session.absoluteTTL"])
	if absoluteTTL <= 0 {
		absoluteTTL = defaultAbsoluteTTL
	}
	idleTTL, _ := time.ParseDuration(cfg["auth.session.idleTTL"])
	if idleTTL <= 0 {
		idleTTL = defaultIdleTTL
	}

	sessionCookieName := strings.TrimSpace(cfg["auth.session.cookie.name"])
	if sessionCookieName == "" {
		sessionCookieName = defaultSessionCookieName
	}
	csrfCookieName := strings.TrimSpace(cfg["auth.session.cookie.csrfName"])
	if csrfCookieName == "" {
		csrfCookieName = defaultCSRFCookieName
	}
	cookiePath := strings.TrimSpace(cfg["auth.session.cookie.path"])
	if cookiePath == "" {
		cookiePath = "/"
	}
	cookieDomain := strings.TrimSpace(cfg["auth.session.cookie.domain"])

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
	now := time.Now().UTC()
	expiresAt := now.Add(absoluteTTL)
	idleExpiresAt := now.Add(idleTTL)
	if idleExpiresAt.After(expiresAt) {
		idleExpiresAt = expiresAt
	}
	cols := dao.AuthSessions.Columns()
	var ipVal any
	if normalized := normalizeIP(ip); normalized != "" {
		ipVal = normalized
	}
	_, err = dao.AuthSessions.Ctx(ctx).Data(do.AuthSessions{
		Id:            sessionID,
		UserId:        userID,
		SecretHash:    hashToken(secretToken, sessionSecret),
		CsrfHash:      hashToken(csrfToken, sessionSecret),
		UserAgent:     trimForDB(userAgent),
		Ip:            ipVal,
		ExpiresAt:     expiresAt,
		IdleExpiresAt: idleExpiresAt,
	}).Insert()
	if err != nil {
		return nil, nil, gerror.Wrap(err, "insert auth session")
	}
	record, err := dao.AuthSessions.Ctx(ctx).Where(cols.Id, sessionID).One()
	if err != nil {
		return nil, nil, gerror.Wrap(err, "select created auth session")
	}
	session, err := mapSession(record)
	if err != nil {
		return nil, nil, err
	}

	// Build cookies using batch-read values for name, path, and domain.
	// Secure and SameSite still come from cookieAttrs (which reads auth.session.cookie.secure individually).
	attrs := cookieAttrs(ctx)
	attrs.Path = cookiePath
	attrs.Domain = cookieDomain
	cookies := buildCookiesWithConfig(sessionCookieName, csrfCookieName, attrs, sessionID+"."+secretToken, csrfToken, expiresAt)
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
	cols := dao.AuthSessions.Columns()
	record, err := dao.AuthSessions.Ctx(ctx).
		Where(cols.Id, sessionID).
		Where(cols.SecretHash, hashToken(secretToken, secret)).
		Where(cols.RevokedAt + " IS NULL").
		Where(cols.ExpiresAt + " > NOW()").
		Where(cols.IdleExpiresAt + " > NOW()").
		One()
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
	record, err := dao.AuthSessions.Ctx(ctx).Where(dao.AuthSessions.Columns().Id, sessionID).One()
	if err != nil {
		return nil, gerror.Wrap(err, "select auth session")
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("auth session %s not found", sessionID)
	}
	return mapSession(record)
}

func (s *sAuthSession) ListUserSessions(ctx context.Context, userID string) ([]*service.AuthSession, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	cols := dao.AuthSessions.Columns()
	records, err := dao.AuthSessions.Ctx(ctx).
		Where(cols.UserId, userID).
		Where(cols.RevokedAt + " IS NULL").
		Where(cols.ExpiresAt + " > NOW()").
		Where(cols.IdleExpiresAt + " > NOW()").
		OrderDesc(cols.LastUsedAt).
		All()
	if err != nil {
		return nil, gerror.Wrap(err, "select user auth sessions")
	}
	sessions := make([]*service.AuthSession, 0, len(records))
	for _, record := range records {
		session, err := mapSession(record)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
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
	cols := dao.AuthSessions.Columns()
	record, err := dao.AuthSessions.Ctx(ctx).
		Fields(cols.CsrfHash).
		Where(cols.Id, sessionID).
		Where(cols.RevokedAt + " IS NULL").
		One()
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
	cols := dao.AuthSessions.Columns()
	_, err := dao.AuthSessions.Ctx(ctx).
		Where(cols.Id, sessionID).
		Where(cols.RevokedAt + " IS NULL").
		Data(do.AuthSessions{RevokedAt: time.Now().UTC(), RevokeReason: reason}).
		Update()
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
	cols := dao.AuthSessions.Columns()
	m := dao.AuthSessions.Ctx(ctx).
		Where(cols.UserId, userID).
		Where(cols.RevokedAt + " IS NULL").
		Data(do.AuthSessions{
			RevokedAt:    time.Now().UTC(),
			RevokeReason: reason,
		})
	if exceptSessionID != "" {
		m = m.Where(cols.Id+" <> ?", exceptSessionID)
	}
	_, err := m.Update()
	return gerror.Wrap(err, "revoke user auth sessions")
}

func (s *sAuthSession) SessionCookieName(ctx context.Context) string {
	return strings.TrimSpace(service.Config().GetString(ctx, "auth.session.cookie.name", defaultSessionCookieName))
}

func (s *sAuthSession) CSRFCookieName(ctx context.Context) string {
	return strings.TrimSpace(service.Config().GetString(ctx, "auth.session.cookie.csrfName", defaultCSRFCookieName))
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
	cols := dao.AuthSessions.Columns()

	// Read current idle_expires_at and expires_at to compute new value in Go
	record, err := dao.AuthSessions.Ctx(ctx).
		Fields(cols.IdleExpiresAt, cols.ExpiresAt).
		Where(cols.Id, sessionID).
		Where(cols.RevokedAt + " IS NULL").
		One()
	if err != nil {
		return gerror.Wrap(err, "select session for touch")
	}
	if record.IsEmpty() {
		return nil
	}

	currentIdle := record[cols.IdleExpiresAt].Time()
	expiresAt := record[cols.ExpiresAt].Time()
	newIdle := currentIdle.Add(idleTTL)
	if newIdle.After(expiresAt) {
		newIdle = expiresAt
	}

	_, err = dao.AuthSessions.Ctx(ctx).
		Where(cols.Id, sessionID).
		Where(cols.RevokedAt + " IS NULL").
		Data(do.AuthSessions{
			LastUsedAt:    time.Now().UTC(),
			IdleExpiresAt: newIdle,
		}).Update()
	return gerror.Wrap(err, "touch auth session")
}

func (s *sAuthSession) buildCookies(ctx context.Context, sessionValue, csrfValue string, expiresAt time.Time) *service.AuthSessionCookies {
	return buildCookiesWithConfig(s.SessionCookieName(ctx), s.CSRFCookieName(ctx), cookieAttrs(ctx), sessionValue, csrfValue, expiresAt)
}

func buildCookiesWithConfig(sessionCookieName, csrfCookieName string, attrs cookieConfig, sessionValue, csrfValue string, expiresAt time.Time) *service.AuthSessionCookies {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	return &service.AuthSessionCookies{
		Session: &http.Cookie{Name: sessionCookieName, Value: sessionValue, Path: attrs.Path, Domain: attrs.Domain, MaxAge: maxAge, Expires: expiresAt, HttpOnly: true, Secure: attrs.Secure, SameSite: attrs.SameSite},
		CSRF:    &http.Cookie{Name: csrfCookieName, Value: csrfValue, Path: attrs.Path, Domain: attrs.Domain, MaxAge: maxAge, Expires: expiresAt, HttpOnly: false, Secure: attrs.Secure, SameSite: attrs.SameSite},
	}
}

type cookieConfig struct {
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

func cookieAttrs(ctx context.Context) cookieConfig {
	path := strings.TrimSpace(service.Config().GetString(ctx, "auth.session.cookie.path", "/"))
	if path == "" {
		path = "/"
	}
	domain := strings.TrimSpace(service.Config().GetString(ctx, "auth.session.cookie.domain", ""))
	return cookieConfig{
		Path:     path,
		Domain:   domain,
		Secure:   service.Config().GetBool(ctx, "auth.session.cookie.secure", true),
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
	secret := strings.TrimSpace(service.Config().GetString(ctx, "auth.session.secret", ""))
	if secret == "" {
		return "", gerror.NewCode(gcode.CodeMissingConfiguration, "auth.session.secret is required")
	}
	return secret, nil
}

func durationConfig(ctx context.Context, key string, fallback time.Duration) time.Duration {
	d := service.Config().GetDuration(ctx, key, fallback)
	if d <= 0 {
		return fallback
	}
	return d
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
