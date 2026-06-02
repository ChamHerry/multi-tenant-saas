package passwordauth

import (
	"context"
	"encoding/json"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

const (
	passwordProvider      = "password"
	invalidLoginMessage   = "invalid email or password"
	defaultPasswordMinLen = 15
	defaultBcryptCost     = 12
	defaultLockThreshold  = 10
	defaultLockDuration   = 15 * time.Minute

	firstPlatformAdminLockKey = "multi-tenant-saas:first-platform-admin-bootstrap"
	platformSuperAdminRole    = "super_admin"
	platformAdminBootstrapLog = "platform_admin.bootstrap_first_user"
)

var (
	codeConflict      = gcode.New(409001, "Conflict", nil)
	internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	dummyPasswordHash []byte
	bootstrapMu       sync.Mutex
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
	// 3-table JOIN via GoFrame DAO
	record, err := dao.Users.Ctx(ctx).
		Fields("users.*, c.password_hash").
		InnerJoin("user_identities ui", "ui.user_id = users.id").
		InnerJoin("user_password_credentials c", "c.user_id = users.id").
		Where("ui.provider", passwordProvider).
		Where("lower(ui.email) = lower(?)", email).
		Where("users.deleted_at IS NULL").
		One()
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
	hash, err := s.hashPassword(ctx, in.Password)
	if err != nil {
		return nil, err
	}
	userID := uuid.GenerateV4()
	identityID := uuid.GenerateV4()
	if err = validateInternalID(userID); err != nil {
		return nil, err
	}
	if err = validateInternalID(identityID); err != nil {
		return nil, err
	}
	rawProfile, err := marshalJSONString(map[string]any{"source": "password"}, "marshal raw identity profile")
	if err != nil {
		return nil, err
	}
	metadata, err := marshalJSONString(map[string]any{}, "marshal user metadata")
	if err != nil {
		return nil, err
	}
	displayName := strings.TrimSpace(in.DisplayName)
	hashCost := passwordCost(ctx)

	err = dao.Users.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Use Go mutex instead of pg_advisory_xact_lock
		bootstrapMu.Lock()
		defer bootstrapMu.Unlock()
		exists, err := passwordIdentityExistsTx(ctx, tx, email)
		if err != nil {
			return err
		}
		if exists {
			return registrationConflictError(email)
		}
		userCount, err := dao.Users.Ctx(ctx).TX(tx).Count()
		if err != nil {
			return gerror.Wrap(err, "count public users")
		}
		if err := insertPasswordUserTx(ctx, tx, passwordRegistrationInsert{
			UserID:       userID,
			IdentityID:   identityID,
			Email:        email,
			DisplayName:  displayName,
			Metadata:     metadata,
			RawProfile:   rawProfile,
			PasswordHash: string(hash),
			HashCost:     hashCost,
		}); err != nil {
			return err
		}
		_, err = bootstrapFirstPlatformAdminIfNeeded(ctx, tx, userID, userCount)
		return err
	})
	if err != nil {
		return nil, err
	}
	return service.UserService().GetUser(ctx, userID)
}

func passwordIdentityExists(ctx context.Context, email string) (bool, error) {
	cols := dao.UserIdentities.Columns()
	count, err := dao.UserIdentities.Ctx(ctx).
		Where(cols.Provider, passwordProvider).
		Where("lower("+cols.Email+") = lower(?)", email).
		Count()
	if err != nil {
		return false, gerror.Wrap(err, "select password identity")
	}
	return count > 0, nil
}

func registrationConflictError(email string) error {
	return gerror.NewCodef(codeConflict, "email %s is already registered", email)
}

type passwordRegistrationInsert struct {
	UserID       string
	IdentityID   string
	Email        string
	DisplayName  string
	Metadata     string
	RawProfile   string
	PasswordHash string
	HashCost     int
}

func passwordIdentityExistsTx(ctx context.Context, tx gdb.TX, email string) (bool, error) {
	cols := dao.UserIdentities.Columns()
	count, err := dao.UserIdentities.Ctx(ctx).TX(tx).
		Where(cols.Provider, passwordProvider).
		Where("lower("+cols.Email+") = lower(?)", email).
		Count()
	if err != nil {
		return false, gerror.Wrap(err, "count password identity")
	}
	return count > 0, nil
}

func insertPasswordUserTx(ctx context.Context, tx gdb.TX, in passwordRegistrationInsert) error {
	userCols := dao.Users.Columns()
	if _, err := dao.Users.Ctx(ctx).TX(tx).Data(g.Map{
		userCols.Id:          in.UserID,
		userCols.Email:       in.Email,
		userCols.DisplayName: nilIfEmptyString(in.DisplayName),
		userCols.AvatarUrl:   nil,
		userCols.Status:      "active",
		userCols.Metadata:    in.Metadata,
		userCols.LastLoginAt: "now()",
		userCols.CreatedAt:   "now()",
		userCols.UpdatedAt:   "now()",
	}).Insert(); err != nil {
		return gerror.Wrap(err, "insert public user")
	}
	identCols := dao.UserIdentities.Columns()
	if _, err := dao.UserIdentities.Ctx(ctx).TX(tx).Data(g.Map{
		identCols.Id:            in.IdentityID,
		identCols.UserId:        in.UserID,
		identCols.Provider:      passwordProvider,
		identCols.AuthId:        in.Email,
		identCols.Email:         in.Email,
		identCols.EmailVerified: false,
		identCols.RawProfile:    in.RawProfile,
		identCols.LastLoginAt:   "now()",
		identCols.CreatedAt:     "now()",
		identCols.UpdatedAt:     "now()",
	}).Insert(); err != nil {
		return gerror.Wrap(err, "insert user identity")
	}
	if err := upsertCredentialTx(ctx, tx, in.UserID, in.PasswordHash, in.HashCost); err != nil {
		return err
	}
	return nil
}

func upsertCredentialTx(ctx context.Context, tx gdb.TX, userID, passwordHash string, cost int) error {
	if err := validateInternalID(userID); err != nil {
		return err
	}
	cols := dao.UserPasswordCredentials.Columns()
	now := time.Now()

	existing, err := dao.UserPasswordCredentials.Ctx(ctx).TX(tx).Where(cols.UserId, userID).One()
	if err != nil {
		return gerror.Wrap(err, "select password credential")
	}

	if existing.IsEmpty() {
		_, err = dao.UserPasswordCredentials.Ctx(ctx).TX(tx).Data(g.Map{
			cols.UserId:            userID,
			cols.PasswordHash:      passwordHash,
			cols.HashCost:          cost,
			cols.PasswordChangedAt: now,
			cols.CreatedAt:         now,
			cols.UpdatedAt:         now,
		}).Insert()
	} else {
		_, err = dao.UserPasswordCredentials.Ctx(ctx).TX(tx).
			Where(cols.UserId, userID).
			Data(g.Map{
				cols.PasswordHash:      passwordHash,
				cols.HashCost:          cost,
				cols.PasswordChangedAt: now,
				cols.UpdatedAt:         now,
			}).Update()
	}
	return gerror.Wrap(err, "upsert password credential")
}

func bootstrapFirstPlatformAdminIfNeeded(ctx context.Context, tx gdb.TX, userID string, userCount int) (bool, error) {
	if !shouldBootstrapFirstPlatformAdmin(userCount) {
		return false, nil
	}
	if err := validateInternalID(userID); err != nil {
		return false, err
	}
	cols := dao.PlatformAdmins.Columns()
	now := time.Now()

	existing, err := dao.PlatformAdmins.Ctx(ctx).TX(tx).Where(cols.UserId, userID).One()
	if err != nil {
		return false, gerror.Wrap(err, "select platform admin")
	}

	if existing.IsEmpty() {
		_, err = dao.PlatformAdmins.Ctx(ctx).TX(tx).Data(g.Map{
			cols.UserId:          userID,
			cols.Role:            platformSuperAdminRole,
			cols.Status:          "active",
			cols.CreatedByUserId: userID,
			cols.CreatedAt:       now,
			cols.UpdatedAt:       now,
		}).Insert()
	} else {
		_, err = dao.PlatformAdmins.Ctx(ctx).TX(tx).
			Where(cols.UserId, userID).
			Data(g.Map{
				cols.Role:      platformSuperAdminRole,
				cols.Status:    "active",
				cols.UpdatedAt: now,
			}).Update()
	}
	if err != nil {
		return false, gerror.Wrap(err, "bootstrap first platform admin")
	}
	if err = insertBootstrapAuditTx(ctx, tx, userID); err != nil {
		return false, err
	}
	return true, nil
}

func shouldBootstrapFirstPlatformAdmin(userCount int) bool {
	return userCount == 0
}

func insertBootstrapAuditTx(ctx context.Context, tx gdb.TX, userID string) error {
	metadata, err := marshalJSONString(map[string]any{
		"role":   platformSuperAdminRole,
		"source": "first_user_registration",
	}, "marshal platform admin bootstrap metadata")
	if err != nil {
		return err
	}
	auditCols := dao.AuditLogs.Columns()
	_, err = dao.AuditLogs.Ctx(ctx).TX(tx).Data(g.Map{
		auditCols.TenantId:     nil,
		auditCols.UserId:       userID,
		auditCols.Action:       platformAdminBootstrapLog,
		auditCols.ResourceType: "platform_admin",
		auditCols.ResourceId:   userID,
		auditCols.Metadata:     metadata,
		auditCols.CreatedAt:    "now()",
	}).Insert()
	return gerror.Wrap(err, "insert platform admin bootstrap audit log")
}

func marshalJSONString(in map[string]any, message string) (string, error) {
	payload, err := json.Marshal(in)
	if err != nil {
		return "", gerror.Wrap(err, message)
	}
	return string(payload), nil
}

func nilIfEmptyString(value string) any {
	if value == "" {
		return nil
	}
	return value
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
	// Find password user by identity email
	identCols := dao.UserIdentities.Columns()
	identRecord, err := dao.UserIdentities.Ctx(ctx).
		Where(identCols.Provider, passwordProvider).
		Where("lower("+identCols.Email+") = lower(?)", email).
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select password user")
	}
	if identRecord.IsEmpty() {
		return nil, gerror.Newf("password user %s not found", email)
	}
	hash, err := s.hashPassword(ctx, password)
	if err != nil {
		return nil, err
	}
	userID := identRecord["user_id"].String()
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
	cols := dao.UserPasswordCredentials.Columns()
	value, err := dao.UserPasswordCredentials.Ctx(ctx).
		Where(cols.UserId, userID).
		Value(cols.PasswordHash)
	if err != nil {
		return err
	}
	if value == nil || value.String() == "" {
		performDummyCompare(oldPassword)
		return loginError()
	}
	if err = bcrypt.CompareHashAndPassword([]byte(value.String()), []byte(oldPassword)); err != nil {
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
	cols := dao.AuthLoginAttempts.Columns()
	value, err := dao.AuthLoginAttempts.Ctx(ctx).
		Where(cols.LoginKey, email).
		Where(cols.LockedUntil+" > NOW()").
		Value(cols.LockedUntil)
	if err != nil {
		return err
	}
	if value != nil && value.String() != "" {
		return gerror.NewCode(gcode.CodeNotAuthorized, "too many failed login attempts; try again later")
	}
	return nil
}

func (s *sPasswordAuth) recordLoginFailure(ctx context.Context, email, ip string) error {
	threshold := service.Config().GetInt(ctx, "auth.password.lockThreshold", defaultLockThreshold)
	if threshold <= 0 {
		threshold = defaultLockThreshold
	}
	lockDuration := durationConfig(ctx, "auth.password.lockDuration", defaultLockDuration)

	cols := dao.AuthLoginAttempts.Columns()
	now := time.Now()
	ipVal := attemptIP(ip)

	// Check existing attempt record
	record, err := dao.AuthLoginAttempts.Ctx(ctx).Where(cols.LoginKey, email).One()
	if err != nil {
		return gerror.Wrap(err, "select login attempt")
	}

	if record.IsEmpty() {
		// First failure for this login key
		data := g.Map{
			cols.LoginKey:     email,
			cols.Ip:           ipVal,
			cols.FailedCount:  1,
			cols.LastFailedAt: now,
			cols.CreatedAt:    now,
			cols.UpdatedAt:    now,
		}
		if 1 >= threshold {
			data[cols.LockedUntil] = now.Add(lockDuration)
		}
		_, err = dao.AuthLoginAttempts.Ctx(ctx).Data(data).Insert()
		return gerror.Wrap(err, "record login failure")
	}

	// Increment existing failure count
	newCount := record[cols.FailedCount].Int() + 1
	data := g.Map{
		cols.FailedCount:  newCount,
		cols.LastFailedAt: now,
		cols.Ip:           ipVal,
		cols.UpdatedAt:    now,
	}
	if newCount >= threshold {
		data[cols.LockedUntil] = now.Add(lockDuration)
	}
	_, err = dao.AuthLoginAttempts.Ctx(ctx).Where(cols.LoginKey, email).Data(data).Update()
	return gerror.Wrap(err, "record login failure")
}

func (s *sPasswordAuth) recordLoginSuccess(ctx context.Context, email string) error {
	cols := dao.AuthLoginAttempts.Columns()
	_, err := dao.AuthLoginAttempts.Ctx(ctx).
		Where(cols.LoginKey, email).
		Data(g.Map{cols.FailedCount: 0, cols.LockedUntil: nil, cols.LastSuccessAt: "now()", cols.UpdatedAt: "now()"}).
		Update()
	return gerror.Wrap(err, "record login success")
}

func touchPasswordIdentityLogin(ctx context.Context, userID, email string) error {
	identCols := dao.UserIdentities.Columns()
	_, err := dao.UserIdentities.Ctx(ctx).
		Where(identCols.Provider, passwordProvider).
		Where("lower("+identCols.AuthId+") = lower(?)", email).
		Data(g.Map{identCols.LastLoginAt: "now()", identCols.UpdatedAt: "now()"}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "touch password identity login")
	}
	userCols := dao.Users.Columns()
	_, err = dao.Users.Ctx(ctx).
		Where(userCols.Id, userID).
		Where("deleted_at IS NULL").
		Data(g.Map{userCols.LastLoginAt: "now()", userCols.UpdatedAt: "now()"}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "touch password user login")
	}
	return nil
}

func upsertCredential(ctx context.Context, userID, passwordHash string, cost int) error {
	if err := validateInternalID(userID); err != nil {
		return err
	}
	cols := dao.UserPasswordCredentials.Columns()
	now := time.Now()

	existing, err := dao.UserPasswordCredentials.Ctx(ctx).Where(cols.UserId, userID).One()
	if err != nil {
		return gerror.Wrap(err, "select password credential")
	}

	if existing.IsEmpty() {
		_, err = dao.UserPasswordCredentials.Ctx(ctx).Data(g.Map{
			cols.UserId:            userID,
			cols.PasswordHash:      passwordHash,
			cols.HashCost:          cost,
			cols.PasswordChangedAt: now,
			cols.CreatedAt:         now,
			cols.UpdatedAt:         now,
		}).Insert()
	} else {
		_, err = dao.UserPasswordCredentials.Ctx(ctx).
			Where(cols.UserId, userID).
			Data(g.Map{
				cols.PasswordHash:      passwordHash,
				cols.HashCost:          cost,
				cols.PasswordChangedAt: now,
				cols.UpdatedAt:         now,
			}).Update()
	}
	return gerror.Wrap(err, "upsert password credential")
}

func validatePassword(ctx context.Context, password string) error {
	if password == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "password is required")
	}
	minLen := service.Config().GetInt(ctx, "auth.password.minLength", defaultPasswordMinLen)
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
	cost := service.Config().GetInt(ctx, "auth.password.bcryptCost", defaultBcryptCost)
	if cost < 10 || cost > 16 {
		return defaultBcryptCost
	}
	return cost
}

func durationConfig(ctx context.Context, key string, fallback time.Duration) time.Duration {
	return service.Config().GetDuration(ctx, key, fallback)
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
