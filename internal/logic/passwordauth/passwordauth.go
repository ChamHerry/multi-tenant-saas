package passwordauth

import (
	"context"
	"encoding/json"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	"repomind-temp/internal/service"
)

const (
	passwordProvider      = "password"
	invalidLoginMessage   = "invalid email or password"
	defaultPasswordMinLen = 15
	defaultBcryptCost     = 12
	defaultLockThreshold  = 10
	defaultLockDuration   = 15 * time.Minute
)

var (
	codeConflict      = gcode.New(409001, "Conflict", nil)
	internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	dummyPasswordHash []byte
)

type sPasswordAuth struct{}

func init() {
	dummyPasswordHash, _ = bcrypt.GenerateFromPassword([]byte("repomind-dummy-password"), bcrypt.MinCost)
	service.RegisterPasswordAuth(&sPasswordAuth{})
}

func (s *sPasswordAuth) Login(ctx context.Context, in service.PasswordLoginInput) (*service.PasswordLoginResult, error) {
	email := normalizeEmail(in.Email)
	if email == "" || strings.TrimSpace(in.Password) == "" {
		return nil, loginError()
	}
	if err := s.ensureLoginAllowed(ctx, email); err != nil {
		return nil, err
	}
	record, err := g.DB().GetOne(ctx, `
SELECT u.id, u.email, u.display_name, u.avatar_url, u.status, u.last_login_at, u.metadata, u.created_at, u.updated_at,
       c.password_hash
FROM public.user_identities i
JOIN public.users u ON u.id = i.user_id
JOIN public.user_password_credentials c ON c.user_id = u.id
WHERE i.provider='password'
  AND lower(i.auth_id)=?
  AND u.deleted_at IS NULL
LIMIT 1`, email)
	if err != nil {
		return nil, gerror.Wrap(err, "select password credential")
	}
	if record.IsEmpty() {
		performDummyCompare(in.Password)
		_ = s.recordLoginFailure(ctx, email, in.IP)
		_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "auth.login.failed", ResourceType: "auth", IP: normalizeIP(in.IP), UserAgent: in.UserAgent, Metadata: map[string]any{"email": email}})
		return nil, loginError()
	}
	user, err := mapUser(record)
	if err != nil {
		return nil, err
	}
	if user.Status != "active" {
		performDummyCompare(in.Password)
		_ = s.recordLoginFailure(ctx, email, in.IP)
		_ = service.Audit().Write(ctx, service.AuditLogInput{UserID: user.ID, Action: "auth.login.failed", ResourceType: "auth", IP: normalizeIP(in.IP), UserAgent: in.UserAgent, Metadata: map[string]any{"email": email, "reason": "inactive_user"}})
		return nil, loginError()
	}
	if err = bcrypt.CompareHashAndPassword([]byte(record["password_hash"].String()), []byte(in.Password)); err != nil {
		_ = s.recordLoginFailure(ctx, email, in.IP)
		_ = service.Audit().Write(ctx, service.AuditLogInput{UserID: user.ID, Action: "auth.login.failed", ResourceType: "auth", IP: normalizeIP(in.IP), UserAgent: in.UserAgent, Metadata: map[string]any{"email": email}})
		return nil, loginError()
	}
	if err = s.recordLoginSuccess(ctx, email); err != nil {
		return nil, err
	}
	if err = touchPasswordIdentityLogin(ctx, user.ID, email); err != nil {
		return nil, err
	}
	session, cookies, err := service.AuthSessionService().Create(ctx, user.ID, in.UserAgent, in.IP)
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{UserID: user.ID, Action: "auth.login.success", ResourceType: "auth_session", ResourceID: session.ID, IP: normalizeIP(in.IP), UserAgent: in.UserAgent, Metadata: map[string]any{"email": email}})
	user, _ = service.UserService().GetUser(ctx, user.ID)
	return &service.PasswordLoginResult{User: user, Session: session, Cookies: cookies}, nil
}

func (s *sPasswordAuth) RegisterPasswordUser(ctx context.Context, in service.RegisterPasswordUserInput) (*service.User, error) {
	email := normalizeEmail(in.Email)
	if email == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "email is required")
	}
	exists, err := passwordIdentityExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, registrationConflictError(email)
	}
	return s.CreatePasswordUser(ctx, service.CreatePasswordUserInput{
		Email:         email,
		Password:      in.Password,
		DisplayName:   in.DisplayName,
		EmailVerified: false,
	})
}

func passwordIdentityExists(ctx context.Context, email string) (bool, error) {
	record, err := g.DB().GetOne(ctx, `
SELECT 1
FROM public.user_identities i
JOIN public.users u ON u.id = i.user_id
WHERE i.provider='password' AND lower(i.auth_id)=? AND u.deleted_at IS NULL
LIMIT 1`, email)
	if err != nil {
		return false, gerror.Wrap(err, "select password identity")
	}
	return !record.IsEmpty(), nil
}

func registrationConflictError(email string) error {
	return gerror.NewCodef(codeConflict, "email %s is already registered", email)
}

func (s *sPasswordAuth) CreatePasswordUser(ctx context.Context, in service.CreatePasswordUserInput) (*service.User, error) {
	email := normalizeEmail(in.Email)
	if email == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "email is required")
	}
	hash, err := s.hashPassword(ctx, in.Password)
	if err != nil {
		return nil, err
	}
	user, err := service.UserService().EnsureUserByIdentity(ctx, service.EnsureUserByIdentityInput{
		Provider:      passwordProvider,
		AuthID:        email,
		Email:         email,
		EmailVerified: in.EmailVerified,
		DisplayName:   strings.TrimSpace(in.DisplayName),
		RawProfile:    map[string]any{"source": "password"},
	})
	if err != nil {
		return nil, err
	}
	if err = upsertCredential(ctx, user.ID, string(hash), passwordCost(ctx)); err != nil {
		return nil, err
	}
	return service.UserService().GetUser(ctx, user.ID)
}

func (s *sPasswordAuth) SetPassword(ctx context.Context, email, password string) (*service.User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "email is required")
	}
	record, err := g.DB().GetOne(ctx, `
SELECT u.id
FROM public.user_identities i
JOIN public.users u ON u.id = i.user_id
WHERE i.provider='password' AND lower(i.auth_id)=? AND u.deleted_at IS NULL
LIMIT 1`, email)
	if err != nil {
		return nil, gerror.Wrap(err, "select password user")
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("password user %s not found", email)
	}
	hash, err := s.hashPassword(ctx, password)
	if err != nil {
		return nil, err
	}
	userID := record["id"].String()
	if err = upsertCredential(ctx, userID, string(hash), passwordCost(ctx)); err != nil {
		return nil, err
	}
	_ = service.AuthSessionService().RevokeUserSessions(ctx, userID, "", "password_set")
	return service.UserService().GetUser(ctx, userID)
}

func (s *sPasswordAuth) ChangePassword(ctx context.Context, userID, currentSessionID, oldPassword, newPassword string) error {
	if err := validateInternalID(userID); err != nil {
		return err
	}
	record, err := g.DB().GetOne(ctx, `
SELECT password_hash
FROM public.user_password_credentials
WHERE user_id=?`, userID)
	if err != nil {
		return gerror.Wrap(err, "select user password credential")
	}
	if record.IsEmpty() {
		performDummyCompare(oldPassword)
		return loginError()
	}
	if err = bcrypt.CompareHashAndPassword([]byte(record["password_hash"].String()), []byte(oldPassword)); err != nil {
		return loginError()
	}
	hash, err := s.hashPassword(ctx, newPassword)
	if err != nil {
		return err
	}
	if err = upsertCredential(ctx, userID, string(hash), passwordCost(ctx)); err != nil {
		return err
	}
	if err = service.AuthSessionService().RevokeUserSessions(ctx, userID, currentSessionID, "password_changed"); err != nil {
		return err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{UserID: userID, Action: "auth.password.changed", ResourceType: "user", ResourceID: userID})
	return nil
}

func (s *sPasswordAuth) hashPassword(ctx context.Context, password string) ([]byte, error) {
	if err := validatePassword(ctx, password); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost(ctx))
	if err != nil {
		return nil, gerror.Wrap(err, "hash password")
	}
	return hash, nil
}

func (s *sPasswordAuth) ensureLoginAllowed(ctx context.Context, email string) error {
	record, err := g.DB().GetOne(ctx, `
SELECT locked_until
FROM public.auth_login_attempts
WHERE login_key=? AND locked_until IS NOT NULL AND locked_until > now()
ORDER BY locked_until DESC
LIMIT 1`, email)
	if err != nil {
		return gerror.Wrap(err, "select login attempts")
	}
	if !record.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotAuthorized, "too many failed login attempts; try again later")
	}
	return nil
}

func (s *sPasswordAuth) recordLoginFailure(ctx context.Context, email, ip string) error {
	threshold := g.Cfg().MustGet(ctx, "auth.password.lockThreshold", defaultLockThreshold).Int()
	if threshold <= 0 {
		threshold = defaultLockThreshold
	}
	lockDuration := durationConfig(ctx, "auth.password.lockDuration", defaultLockDuration)
	_, err := g.DB().Exec(ctx, `
INSERT INTO public.auth_login_attempts(login_key, ip, failed_count, locked_until, last_failed_at, created_at, updated_at)
VALUES (?, ?::inet, 1, CASE WHEN 1 >= ? THEN now() + (?::interval) ELSE NULL END, now(), now(), now())
ON CONFLICT (login_key, ip) DO UPDATE
SET failed_count=public.auth_login_attempts.failed_count + 1,
    locked_until=CASE
        WHEN public.auth_login_attempts.failed_count + 1 >= ? THEN now() + (?::interval)
        ELSE public.auth_login_attempts.locked_until
    END,
    last_failed_at=now(),
    updated_at=now()`, email, attemptIP(ip), threshold, pgInterval(lockDuration), threshold, pgInterval(lockDuration))
	return gerror.Wrap(err, "record login failure")
}

func (s *sPasswordAuth) recordLoginSuccess(ctx context.Context, email string) error {
	_, err := g.DB().Exec(ctx, `
UPDATE public.auth_login_attempts
SET failed_count=0, locked_until=NULL, last_success_at=now(), updated_at=now()
WHERE login_key=?`, email)
	return gerror.Wrap(err, "record login success")
}

func touchPasswordIdentityLogin(ctx context.Context, userID, email string) error {
	if _, err := g.DB().Exec(ctx, `
UPDATE public.user_identities
SET last_login_at=now(), updated_at=now()
WHERE provider='password' AND lower(auth_id)=?`, email); err != nil {
		return gerror.Wrap(err, "touch password identity login")
	}
	if _, err := g.DB().Exec(ctx, `
UPDATE public.users
SET last_login_at=now(), updated_at=now()
WHERE id=? AND deleted_at IS NULL`, userID); err != nil {
		return gerror.Wrap(err, "touch password user login")
	}
	return nil
}

func upsertCredential(ctx context.Context, userID, passwordHash string, cost int) error {
	if err := validateInternalID(userID); err != nil {
		return err
	}
	_, err := g.DB().Exec(ctx, `
INSERT INTO public.user_password_credentials(user_id, password_hash, hash_alg, hash_cost, password_changed_at, created_at, updated_at)
VALUES (?, ?, 'bcrypt', ?, now(), now(), now())
ON CONFLICT (user_id) DO UPDATE
SET password_hash=EXCLUDED.password_hash,
    hash_alg='bcrypt',
    hash_cost=EXCLUDED.hash_cost,
    password_changed_at=now(),
    updated_at=now()`, userID, passwordHash, cost)
	return gerror.Wrap(err, "upsert password credential")
}

func validatePassword(ctx context.Context, password string) error {
	if password == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "password is required")
	}
	minLen := g.Cfg().MustGet(ctx, "auth.password.minLength", defaultPasswordMinLen).Int()
	if minLen <= 0 {
		minLen = defaultPasswordMinLen
	}
	if len([]rune(password)) < minLen {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "password must be at least %d characters", minLen)
	}
	if len([]byte(password)) > 72 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "password must not exceed 72 bytes for bcrypt")
	}
	return nil
}

func passwordCost(ctx context.Context) int {
	cost := g.Cfg().MustGet(ctx, "auth.password.bcryptCost", defaultBcryptCost).Int()
	if cost < 10 || cost > 16 {
		return defaultBcryptCost
	}
	return cost
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
		seconds = int64(defaultLockDuration.Seconds())
	}
	return strconv.FormatInt(seconds, 10) + " seconds"
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if parsed := net.ParseIP(ip); parsed != nil {
		return parsed.String()
	}
	return ""
}

func attemptIP(ip string) string {
	if normalized := normalizeIP(ip); normalized != "" {
		return normalized
	}
	return "0.0.0.0"
}

func loginError() error {
	return gerror.NewCode(gcode.CodeNotAuthorized, invalidLoginMessage)
}

func performDummyCompare(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.Newf("invalid internal uuid %q", id)
	}
	return nil
}

func mapUser(record gdb.Record) (*service.User, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	metadata := map[string]any{}
	if raw := record["metadata"].String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &metadata)
	}
	return &service.User{
		ID:          id,
		Email:       record["email"].String(),
		DisplayName: record["display_name"].String(),
		AvatarURL:   record["avatar_url"].String(),
		Status:      record["status"].String(),
		LastLoginAt: nullableTime(record["last_login_at"]),
		Metadata:    metadata,
		CreatedAt:   record["created_at"].Time(),
		UpdatedAt:   record["updated_at"].Time(),
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
