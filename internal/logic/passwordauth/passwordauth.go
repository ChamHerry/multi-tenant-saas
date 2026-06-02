package passwordauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
		Fields("users.*, c.password_hash, ui.email_verified").
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
	record, err := dao.AuthLoginAttempts.Ctx(ctx).
		Where(cols.LoginKey, email).
		Fields(cols.LockedUntil, cols.FailedCount, cols.LastFailedAt).
		One()
	if err != nil {
		return err
	}
	if record.IsEmpty() {
		return nil // No record at all -- definitely allowed
	}

	// Check if currently locked
	lockedUntil := record[cols.LockedUntil]
	if !lockedUntil.IsNil() {
		lt := lockedUntil.Time()
		if time.Now().Before(lt) {
			return loginError() // Generic message to prevent user enumeration
		}
	}
	return nil
}

func (s *sPasswordAuth) recordLoginFailure(ctx context.Context, email, ip string) error {
	maxAttempts := service.Config().GetInt(ctx, "auth.lockout.max_attempts", 5)
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	windowMinutes := service.Config().GetInt(ctx, "auth.lockout.window_minutes", 15)
	if windowMinutes <= 0 {
		windowMinutes = 15
	}
	lockoutMinutes := service.Config().GetInt(ctx, "auth.lockout.duration_minutes", 15)
	if lockoutMinutes <= 0 {
		lockoutMinutes = 15
	}
	lockDuration := time.Duration(lockoutMinutes) * time.Minute
	windowCutoff := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)

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
		if 1 >= maxAttempts {
			data[cols.LockedUntil] = now.Add(lockDuration)
		}
		_, err = dao.AuthLoginAttempts.Ctx(ctx).Data(data).Insert()
		if err != nil {
			return gerror.Wrap(err, "record login failure")
		}
		if 1 >= maxAttempts {
			_ = s.writeLockAudit(ctx, email, ip, 1, now.Add(lockDuration))
		}
		return nil
	}

	// Determine new count considering sliding window
	newCount := record[cols.FailedCount].Int() + 1
	lastFailedVar := record[cols.LastFailedAt]
	if !lastFailedVar.IsNil() {
		lastFailed := lastFailedVar.Time()
		if lastFailed.Before(windowCutoff) {
			// Last failure is outside the window -- reset counter
			newCount = 1
		}
	}

	data := g.Map{
		cols.FailedCount:  newCount,
		cols.LastFailedAt: now,
		cols.Ip:           ipVal,
		cols.UpdatedAt:    now,
	}
	if newCount >= maxAttempts {
		data[cols.LockedUntil] = now.Add(lockDuration)
	}
	_, err = dao.AuthLoginAttempts.Ctx(ctx).Where(cols.LoginKey, email).Data(data).Update()
	if err != nil {
		return gerror.Wrap(err, "record login failure")
	}
	if newCount >= maxAttempts {
		_ = s.writeLockAudit(ctx, email, ip, newCount, now.Add(lockDuration))
	}
	return nil
}

// writeLockAudit writes an audit log entry when an account is locked.
func (s *sPasswordAuth) writeLockAudit(ctx context.Context, email, ip string, failedCount int, lockedUntil time.Time) error {
	return service.Audit().Write(ctx, service.AuditLogInput{
		Action:       "auth.account.locked",
		ResourceType: "auth",
		IP:           normalizeIP(ip),
		Metadata: map[string]any{
			"email":        email,
			"failed_count": failedCount,
			"locked_until": lockedUntil.Format(time.RFC3339),
		},
	})
}

func (s *sPasswordAuth) recordLoginSuccess(ctx context.Context, email string) error {
	cols := dao.AuthLoginAttempts.Columns()
	_, err := dao.AuthLoginAttempts.Ctx(ctx).
		Where(cols.LoginKey, email).
		Data(g.Map{cols.FailedCount: 0, cols.LockedUntil: nil, cols.LastSuccessAt: "now()", cols.UpdatedAt: "now()"}).
		Update()
	return gerror.Wrap(err, "record login success")
}

// UnlockUser clears the lockout state for a user identified by email.
// It is idempotent -- returns nil even if the user was not locked.
func (s *sPasswordAuth) UnlockUser(ctx context.Context, in service.UnlockUserInput) error {
	if in.Email == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "email is required")
	}
	cols := dao.AuthLoginAttempts.Columns()
	_, err := dao.AuthLoginAttempts.Ctx(ctx).
		Where(cols.LoginKey, normalizeEmail(in.Email)).
		Data(g.Map{
			cols.FailedCount: 0,
			cols.LockedUntil: nil,
			cols.UpdatedAt:   "now()",
		}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "unlock user")
	}
	return nil
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
	emailVerified := false
	if ev := record["email_verified"]; ev != nil {
		emailVerified = ev.Bool()
	}
	return &service.User{
		ID:            id,
		Email:         record["email"].String(),
		DisplayName:   record["display_name"].String(),
		AvatarURL:     record["avatar_url"].String(),
		Status:        record["status"].String(),
		EmailVerified: emailVerified,
		LastLoginAt:   nullableTime(record["last_login_at"]),
		Metadata:      metadata,
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

// ---------------------------------------------------------------------------
// Password Reset
// ---------------------------------------------------------------------------

// generateResetToken creates a cryptographically secure random token.
// Returns a 64-character hex-encoded string (32 bytes = 256 bits entropy).
func generateResetToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", gerror.Wrap(err, "generate reset token")
	}
	return hex.EncodeToString(buf), nil
}

// hashResetToken returns the SHA-256 hex digest of a token.
func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

// findPasswordUserByEmail returns the user ID for a password-identity email,
// or empty string if not found.
func findPasswordUserByEmail(ctx context.Context, email string) (string, error) {
	cols := dao.UserIdentities.Columns()
	value, err := dao.UserIdentities.Ctx(ctx).
		Where(cols.Provider, passwordProvider).
		Where("lower("+cols.Email+") = lower(?)", email).
		Value(cols.UserId)
	if err != nil {
		return "", gerror.Wrap(err, "select user by email")
	}
	if value.IsNil() || value.String() == "" {
		return "", nil
	}
	return value.String(), nil
}

// checkResetRateLimit checks whether the given email has exceeded the
// password reset rate limit (3 requests per 5 minutes per email).
func (s *sPasswordAuth) checkResetRateLimit(ctx context.Context, email string) error {
	redisEnabled := service.Config().GetBool(ctx, "redis.enabled", false)
	if !redisEnabled {
		return nil
	}
	key := fmt.Sprintf("password-reset-rate:%s", normalizeEmail(email))
	count, err := g.Redis().Incr(ctx, key)
	if err != nil {
		g.Log().Warningf(ctx, "[password-reset] rate limit check failed for %s: %v", email, err)
		return nil // fail open — don't block legitimate requests
	}
	if count == 1 {
		_, _ = g.Redis().Expire(ctx, key, int64(5*time.Minute.Seconds()))
	}
	maxRequests := service.Config().GetInt(ctx, "auth.password.resetRateLimit", 3)
	if count > int64(maxRequests) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "too many reset requests; please try again later")
	}
	return nil
}

// resetURL builds the password reset page URL.
func resetURL(ctx context.Context, token string) string {
	base := strings.TrimRight(service.Config().GetString(ctx, "web.baseUrl", ""), "/")
	if base == "" {
		return ""
	}
	return base + "/reset-password?token=" + token
}

// ForgotPassword initiates the password reset flow by email.
// It always returns nil error to prevent user enumeration, even when the
// email does not exist. Rate limiting is the only case that returns an error.
func (s *sPasswordAuth) ForgotPassword(ctx context.Context, email, ip string) error {
	email = normalizeEmail(email)
	if email == "" {
		// Don't reveal whether email is valid
		return nil
	}

	// Check rate limit (per email)
	if err := s.checkResetRateLimit(ctx, email); err != nil {
		return err
	}

	// Look up the user — silently return if not found (prevent enumeration)
	userID, err := findPasswordUserByEmail(ctx, email)
	if err != nil {
		g.Log().Warningf(ctx, "[password-reset] lookup error for %s: %v", email, err)
		return nil
	}
	if userID == "" {
		// User not found — log and return success to prevent enumeration
		_ = service.Audit().Write(ctx, service.AuditLogInput{
			Action:       "auth.password.forgot",
			ResourceType: "auth",
			IP:           normalizeIP(ip),
			Metadata:     map[string]any{"email": email, "result": "email_not_found"},
		})
		return nil
	}

	// Generate token
	token, err := generateResetToken()
	if err != nil {
		return err
	}

	// Store hashed token
	cols := dao.PasswordResetTokens.Columns()
	_, err = dao.PasswordResetTokens.Ctx(ctx).Data(g.Map{
		cols.UserId:      userID,
		cols.TokenHash:   hashResetToken(token),
		cols.ExpiresAt:   time.Now().Add(1 * time.Hour),
		cols.RequestedIp: attemptIP(ip),
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "insert password reset token")
	}

	// Get user display name for email
	userName := email
	if user, err := service.UserService().GetUser(ctx, userID); err == nil && user != nil {
		if user.DisplayName != "" {
			userName = user.DisplayName
		}
	}

	// Send email asynchronously
	targetURL := resetURL(ctx, token)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Warningf(ctx, "[email] panic sending password reset: %v", r)
			}
		}()
		if err := service.Email().SendPasswordReset(ctx, service.SendPasswordResetInput{
			ToEmail:  email,
			UserName: userName,
			ResetURL: targetURL,
		}); err != nil {
			g.Log().Warningf(ctx, "[email] failed to send password reset to %s: %v", email, err)
		}
	}()

	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.password.forgot",
		ResourceType: "auth",
		IP:           normalizeIP(ip),
		Metadata:     map[string]any{"email": email},
	})

	return nil
}

// ResetPassword verifies a reset token and sets a new password.
// On success, all existing sessions for the user are revoked and a
// password-changed notification is sent.
func (s *sPasswordAuth) ResetPassword(ctx context.Context, token, newPassword, ip string) error {
	token = strings.TrimSpace(token)
	// Validate token format: must be 64 hex characters
	if len(token) != 64 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "invalid token format")
	}

	// Hash the provided token
	hashedToken := hashResetToken(token)

	// Look up the token
	cols := dao.PasswordResetTokens.Columns()
	record, err := dao.PasswordResetTokens.Ctx(ctx).
		Where(cols.TokenHash, hashedToken).
		WhereNull(cols.UsedAt).
		One()
	if err != nil {
		return gerror.Wrap(err, "select reset token")
	}
	if record.IsEmpty() {
		performDummyCompare(token)
		_ = service.Audit().Write(ctx, service.AuditLogInput{
			Action:       "auth.password.reset.failed",
			ResourceType: "auth",
			IP:           normalizeIP(ip),
			Metadata:     map[string]any{"reason": "invalid_token"},
		})
		return gerror.NewCode(gcode.CodeNotAuthorized, "invalid or expired reset token")
	}

	// Check expiration
	if time.Now().After(record[cols.ExpiresAt].Time()) {
		performDummyCompare(token)
		_ = service.Audit().Write(ctx, service.AuditLogInput{
			Action:       "auth.password.reset.failed",
			ResourceType: "auth",
			IP:           normalizeIP(ip),
			Metadata:     map[string]any{"reason": "expired_token"},
		})
		return gerror.NewCode(gcode.CodeNotAuthorized, "invalid or expired reset token")
	}

	userID := record[cols.UserId].String()
	if err = validateInternalID(userID); err != nil {
		return err
	}

	// Hash the new password
	passwordHash, err := s.hashPassword(ctx, newPassword)
	if err != nil {
		return err
	}

	hashCost := passwordCost(ctx)

	// Execute in transaction: mark token used + update password
	err = dao.PasswordResetTokens.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Mark current token as used
		_, err := dao.PasswordResetTokens.Ctx(ctx).TX(tx).
			Where(cols.TokenHash, hashedToken).
			WhereNull(cols.UsedAt).
			Data(g.Map{cols.UsedAt: "now()"}).
			Update()
		if err != nil {
			return gerror.Wrap(err, "mark reset token used")
		}

		// Revoke all other unused tokens for this user
		_, err = dao.PasswordResetTokens.Ctx(ctx).TX(tx).
			Where(cols.UserId, userID).
			WhereNull(cols.UsedAt).
			Where(cols.TokenHash+" != ?", hashedToken).
			Data(g.Map{cols.UsedAt: "now()"}).
			Update()
		if err != nil {
			return gerror.Wrap(err, "revoke other reset tokens")
		}

		// Update password credential
		if err = upsertCredentialTx(ctx, tx, userID, string(passwordHash), hashCost); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Revoke all existing sessions for the user
	if err = service.AuthSessionService().RevokeUserSessions(ctx, userID, "", "password_reset"); err != nil {
		g.Log().Warningf(ctx, "[password-reset] failed to revoke sessions for %s: %v", userID, err)
		// Non-fatal: password is already changed
	}

	// Send password changed notification asynchronously
	userEmail := ""
	displayName := ""
	if user, err := service.UserService().GetUser(ctx, userID); err == nil && user != nil {
		userEmail = user.Email
		if user.DisplayName != "" {
			displayName = user.DisplayName
		} else {
			displayName = userEmail
		}
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Warningf(ctx, "[email] panic sending password changed: %v", r)
			}
		}()
		if err := service.Email().SendPasswordChanged(ctx, service.SendPasswordChangedInput{
			ToEmail:  userEmail,
			UserName: displayName,
		}); err != nil {
			g.Log().Warningf(ctx, "[email] failed to send password changed to %s: %v", userEmail, err)
		}
	}()

	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.password.reset",
		ResourceType: "user",
		ResourceID:   userID,
		IP:           normalizeIP(ip),
	})

	return nil
}
