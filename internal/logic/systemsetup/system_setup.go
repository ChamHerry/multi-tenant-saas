package systemsetup

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/do"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/crypto"
	credis "multi-tenant-saas/utility/redis"
	"multi-tenant-saas/utility/uuid"
)

const (
	setupVersion                = "initial-setup-wizard/v1"
	passwordProvider            = "password"
	platformSuperAdminRole      = "super_admin"
	actionSetupCompleted        = "system_setup.completed"
	actionLegacyDetected        = "system_setup.legacy_detected"
	defaultPasswordMinLen       = 15
	defaultBcryptCost           = 12
	configCacheKeyFmt           = "config:%s"
	minGeneratedSecretByteCount = 32
)

var codeConflict = gcode.New(409001, "Conflict", nil)

type sSystemSetup struct{}

func init() {
	service.RegisterSystemSetup(&sSystemSetup{})
}

func (s *sSystemSetup) State(ctx context.Context) (*service.SetupStatus, error) {
	checks := s.checks(ctx)
	row, err := setupRow(ctx)
	if err != nil {
		return nil, err
	}
	initialized := !row.IsEmpty()
	legacyInitialized := false
	if !initialized {
		legacyInitialized, err = s.legacyReady(ctx)
		if err != nil {
			return nil, err
		}
	}
	missing, err := s.missing(ctx)
	if err != nil {
		return nil, err
	}
	status := &service.SetupStatus{
		Initialized:       initialized,
		RequiresSetup:     !initialized && !legacyInitialized,
		LegacyInitialized: legacyInitialized,
		Missing:           missing,
		Checks:            checks,
	}
	if initialized {
		status.Version = row["version"].String()
		status.InitializedByUserID = row["initialized_by_user_id"].String()
		if t := row["initialized_at"].Time(); !t.IsZero() {
			status.InitializedAt = &t
		}
	}
	return status, nil
}

func (s *sSystemSetup) EnsureLegacyState(ctx context.Context) error {
	row, err := setupRow(ctx)
	if err != nil {
		return err
	}
	if !row.IsEmpty() {
		return nil
	}
	ready, err := s.legacyReady(ctx)
	if err != nil {
		return err
	}
	if !ready {
		return nil
	}
	adminID, err := firstActiveSuperAdminID(ctx)
	if err != nil {
		return err
	}
	metadata, err := marshalJSON(map[string]any{"legacy_detected": true, "source": "startup"})
	if err != nil {
		return err
	}
	_, err = g.DB().Exec(ctx, `
		INSERT INTO system_setup(id, status, version, initialized_by_user_id, initialized_at, metadata)
		VALUES (1, 'initialized', ?, ?, now(), ?::jsonb)
		ON CONFLICT (id) DO NOTHING`, setupVersion, nilIfEmpty(adminID), metadata)
	if err != nil {
		return gerror.Wrap(err, "insert legacy system setup row")
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{UserID: adminID, Action: actionLegacyDetected, ResourceType: "system_setup", ResourceID: "1", Metadata: map[string]any{"legacy_detected": true}})
	return nil
}

func (s *sSystemSetup) ValidateRuntimeOrSetupPending(ctx context.Context) error {
	if err := s.EnsureLegacyState(ctx); err != nil {
		return err
	}
	state, err := s.State(ctx)
	if err != nil {
		return err
	}
	if state.RequiresSetup {
		g.Log().Warning(ctx, "[setup] system is not initialized; only setup routes are available")
		return nil
	}
	return s.ValidateRuntime(ctx)
}

func (s *sSystemSetup) ValidateRuntime(ctx context.Context) error {
	return validateRuntimeValues(runtimeValues{
		SessionSecret: service.Config().GetString(ctx, "auth.session.secret", ""),
		APIKeySecret:  service.Config().GetString(ctx, "auth.apiKey.secret", ""),
	})
}

func (s *sSystemSetup) Complete(ctx context.Context, in service.CompleteSetupInput) (*service.CompleteSetupResult, error) {
	adminEmail, err := normalizeEmail(in.Admin.Email)
	if err != nil {
		return nil, err
	}
	displayName := strings.TrimSpace(in.Admin.DisplayName)
	webBaseURL, err := normalizeWebBaseURL(in.Runtime.WebBaseURL)
	if err != nil {
		return nil, err
	}
	if err = validateSetupPassword(ctx, in.Admin.Password); err != nil {
		return nil, err
	}

	sessionSecret, err := generateSecret()
	if err != nil {
		return nil, err
	}
	apiKeySecret, err := generateSecret()
	if err != nil {
		return nil, err
	}
	if err = validateRuntimeValues(runtimeValues{
		SessionSecret: sessionSecret,
		APIKeySecret:  apiKeySecret,
	}); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Admin.Password), passwordCost(ctx))
	if err != nil {
		return nil, gerror.Wrap(err, "hash setup admin password")
	}

	var userID string
	updatedConfigKeys := []string{"web.baseUrl", "auth.session.secret", "auth.apiKey.secret"}
	err = dao.Users.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		row, err := tx.Ctx(ctx).GetOne("SELECT id FROM system_setup WHERE id=1 FOR UPDATE")
		if err != nil {
			return gerror.Wrap(err, "select system setup")
		}
		if !row.IsEmpty() {
			return gerror.NewCode(codeConflict, service.SetupCodeAlreadyInitialized)
		}

		userID, err = upsertSetupAdminTx(ctx, tx, setupAdminUpsert{
			Email:        adminEmail,
			DisplayName:  displayName,
			PasswordHash: string(hash),
			HashCost:     passwordCost(ctx),
		})
		if err != nil {
			return err
		}
		if err = upsertPlatformSuperAdminTx(ctx, tx, userID); err != nil {
			return err
		}
		if err = upsertConfigTx(ctx, tx, "web.baseUrl", webBaseURL, "string", "Frontend base URL for invitation and email links"); err != nil {
			return err
		}
		if err = upsertConfigTx(ctx, tx, "auth.session.secret", sessionSecret, "secret", "Session HMAC secret generated by initial setup"); err != nil {
			return err
		}
		if err = upsertConfigTx(ctx, tx, "auth.apiKey.secret", apiKeySecret, "secret", "API Key HMAC secret generated by initial setup"); err != nil {
			return err
		}
		metadata, err := marshalJSON(map[string]any{
			"source":         "initial_setup_wizard",
			"web_base_url":   webBaseURL,
			"generated_keys": []string{"auth.session.secret", "auth.apiKey.secret"},
		})
		if err != nil {
			return err
		}
		if _, err = tx.Ctx(ctx).Exec(`
			INSERT INTO system_setup(id, status, version, initialized_by_user_id, initialized_at, metadata)
			VALUES (1, 'initialized', ?, ?, now(), ?::jsonb)`, setupVersion, userID, metadata); err != nil {
			return gerror.Wrap(err, "insert system setup row")
		}
		return insertSetupAuditTx(ctx, tx, userID, in, webBaseURL)
	})
	if err != nil {
		return nil, err
	}
	invalidateConfigKeys(ctx, updatedConfigKeys...)
	if err = s.ValidateRuntime(ctx); err != nil {
		return nil, err
	}
	user, err := service.UserService().GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &service.CompleteSetupResult{Initialized: true, User: user}, nil
}

type runtimeValues struct {
	SessionSecret string
	APIKeySecret  string
}

func validateRuntimeValues(values runtimeValues) error {
	if !isSafeSecret(values.SessionSecret) {
		return gerror.New("auth.session.secret must be configured with a safe non-placeholder value")
	}
	if !isSafeSecret(values.APIKeySecret) {
		return gerror.New("auth.apiKey.secret must be configured with a safe non-placeholder value")
	}
	return nil
}

func (s *sSystemSetup) checks(ctx context.Context) service.SetupChecks {
	checks := service.SetupChecks{
		Redis: service.SetupCheck{OK: true, Message: "optional"},
	}
	if _, err := g.DB().GetOne(ctx, "SELECT 1"); err != nil {
		checks.Database = service.SetupCheck{OK: false, Message: err.Error()}
	} else {
		checks.Database = service.SetupCheck{OK: true}
	}
	version, dirty, err := service.AutoMigrate().Status(ctx)
	if err != nil {
		checks.Migrations = service.SetupCheck{OK: false, Message: err.Error()}
	} else if dirty {
		checks.Migrations = service.SetupCheck{OK: false, Message: "dirty migration state"}
	} else {
		checks.Migrations = service.SetupCheck{OK: true, Message: fmt.Sprint(version)}
	}
	if _, err := crypto.Encrypt("setup-check"); err != nil {
		checks.Encryption = service.SetupCheck{OK: false, Message: err.Error()}
	} else {
		checks.Encryption = service.SetupCheck{OK: true}
	}
	if adapter := credis.GetCacheManager().GetAdapter("default"); adapter != nil {
		if err := adapter.Ping(ctx); err != nil {
			checks.Redis = service.SetupCheck{OK: false, Message: err.Error()}
		} else {
			checks.Redis = service.SetupCheck{OK: true}
		}
	}
	return checks
}

func setupRow(ctx context.Context) (gdb.Record, error) {
	row, err := g.DB().GetOne(ctx, `SELECT id, status, version, initialized_by_user_id, initialized_at, metadata FROM system_setup WHERE id=1`)
	if err != nil {
		return nil, gerror.Wrap(err, "select system setup")
	}
	return row, nil
}

func (s *sSystemSetup) legacyReady(ctx context.Context) (bool, error) {
	hasAdmin, err := hasActiveSuperAdmin(ctx)
	if err != nil {
		return false, err
	}
	if !hasAdmin {
		return false, nil
	}
	return isSafeSecret(service.Config().GetString(ctx, "auth.session.secret", "")) &&
		isSafeSecret(service.Config().GetString(ctx, "auth.apiKey.secret", "")), nil
}

func (s *sSystemSetup) missing(ctx context.Context) ([]string, error) {
	missing := []string{}
	hasAdmin, err := hasActiveSuperAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !hasAdmin {
		missing = append(missing, "admin")
	}
	if !isSafeSecret(service.Config().GetString(ctx, "auth.session.secret", "")) {
		missing = append(missing, "auth.session.secret")
	}
	if !isSafeSecret(service.Config().GetString(ctx, "auth.apiKey.secret", "")) {
		missing = append(missing, "auth.apiKey.secret")
	}
	return missing, nil
}

func hasActiveSuperAdmin(ctx context.Context) (bool, error) {
	count, err := dao.PlatformAdmins.Ctx(ctx).
		Where(dao.PlatformAdmins.Columns().Role, platformSuperAdminRole).
		Where(dao.PlatformAdmins.Columns().Status, "active").
		Count()
	return count > 0, gerror.Wrap(err, "count active platform super admins")
}

func firstActiveSuperAdminID(ctx context.Context) (string, error) {
	row, err := dao.PlatformAdmins.Ctx(ctx).
		Where(dao.PlatformAdmins.Columns().Role, platformSuperAdminRole).
		Where(dao.PlatformAdmins.Columns().Status, "active").
		OrderAsc(dao.PlatformAdmins.Columns().CreatedAt).
		One()
	if err != nil {
		return "", gerror.Wrap(err, "select active platform super admin")
	}
	if row.IsEmpty() {
		return "", nil
	}
	return row[dao.PlatformAdmins.Columns().UserId].String(), nil
}

type setupAdminUpsert struct {
	Email        string
	DisplayName  string
	PasswordHash string
	HashCost     int
}

func upsertSetupAdminTx(ctx context.Context, tx gdb.TX, in setupAdminUpsert) (string, error) {
	identCols := dao.UserIdentities.Columns()
	identity, err := dao.UserIdentities.Ctx(ctx).TX(tx).
		Where(identCols.Provider, passwordProvider).
		Where("lower("+identCols.Email+") = lower(?)", in.Email).
		One()
	if err != nil {
		return "", gerror.Wrap(err, "select setup admin identity")
	}
	if !identity.IsEmpty() {
		userID := identity[identCols.UserId].String()
		if err := updateExistingSetupAdminTx(ctx, tx, userID, in); err != nil {
			return "", err
		}
		return userID, nil
	}

	userID := uuid.GenerateV4()
	identityID := uuid.GenerateV4()
	metadata, err := marshalJSON(map[string]any{"source": "initial_setup_wizard"})
	if err != nil {
		return "", err
	}
	rawProfile := metadata
	if _, err = dao.Users.Ctx(ctx).TX(tx).Data(do.Users{
		Id:          userID,
		Email:       in.Email,
		DisplayName: nilIfEmpty(in.DisplayName),
		Status:      "active",
		Metadata:    gjson.New(metadata),
	}).Insert(); err != nil {
		return "", gerror.Wrap(err, "insert setup admin user")
	}
	if _, err = dao.UserIdentities.Ctx(ctx).TX(tx).Data(do.UserIdentities{
		Id:            identityID,
		UserId:        userID,
		Provider:      passwordProvider,
		AuthId:        in.Email,
		Email:         in.Email,
		EmailVerified: true,
		RawProfile:    gjson.New(rawProfile),
	}).Insert(); err != nil {
		return "", gerror.Wrap(err, "insert setup admin identity")
	}
	if err = upsertCredentialTx(ctx, tx, userID, in.PasswordHash, in.HashCost); err != nil {
		return "", err
	}
	return userID, nil
}

func updateExistingSetupAdminTx(ctx context.Context, tx gdb.TX, userID string, in setupAdminUpsert) error {
	userCols := dao.Users.Columns()
	data := do.Users{Status: "active"}
	if in.DisplayName != "" {
		data.DisplayName = in.DisplayName
	}
	if _, err := dao.Users.Ctx(ctx).TX(tx).Where(userCols.Id, userID).Data(data).Update(); err != nil {
		return gerror.Wrap(err, "update setup admin user")
	}
	identCols := dao.UserIdentities.Columns()
	if _, err := dao.UserIdentities.Ctx(ctx).TX(tx).
		Where(identCols.Provider, passwordProvider).
		Where("lower("+identCols.Email+") = lower(?)", in.Email).
		Data(do.UserIdentities{EmailVerified: true}).
		Update(); err != nil {
		return gerror.Wrap(err, "verify setup admin identity")
	}
	return upsertCredentialTx(ctx, tx, userID, in.PasswordHash, in.HashCost)
}

func upsertCredentialTx(ctx context.Context, tx gdb.TX, userID, passwordHash string, cost int) error {
	cols := dao.UserPasswordCredentials.Columns()
	now := time.Now().UTC()
	existing, err := dao.UserPasswordCredentials.Ctx(ctx).TX(tx).Where(cols.UserId, userID).One()
	if err != nil {
		return gerror.Wrap(err, "select setup admin password credential")
	}
	data := do.UserPasswordCredentials{
		PasswordHash:      passwordHash,
		HashCost:          cost,
		PasswordChangedAt: now,
	}
	if existing.IsEmpty() {
		data.UserId = userID
		_, err = dao.UserPasswordCredentials.Ctx(ctx).TX(tx).Data(data).Insert()
	} else {
		_, err = dao.UserPasswordCredentials.Ctx(ctx).TX(tx).Where(cols.UserId, userID).Data(data).Update()
	}
	return gerror.Wrap(err, "upsert setup admin password credential")
}

func upsertPlatformSuperAdminTx(ctx context.Context, tx gdb.TX, userID string) error {
	cols := dao.PlatformAdmins.Columns()
	existing, err := dao.PlatformAdmins.Ctx(ctx).TX(tx).Where(cols.UserId, userID).One()
	if err != nil {
		return gerror.Wrap(err, "select setup platform admin")
	}
	data := do.PlatformAdmins{Role: platformSuperAdminRole, Status: "active"}
	if existing.IsEmpty() {
		data.UserId = userID
		data.CreatedByUserId = userID
		_, err = dao.PlatformAdmins.Ctx(ctx).TX(tx).Data(data).Insert()
	} else {
		_, err = dao.PlatformAdmins.Ctx(ctx).TX(tx).Where(cols.UserId, userID).Data(data).Update()
	}
	return gerror.Wrap(err, "upsert setup platform super admin")
}

func upsertConfigTx(ctx context.Context, tx gdb.TX, key, value, valueType, description string) error {
	storedValue := value
	isEncrypted := false
	if valueType == "secret" {
		encrypted, err := crypto.Encrypt(value)
		if err != nil {
			return gerror.Wrap(err, "encrypt setup config secret")
		}
		storedValue = encrypted
		isEncrypted = true
	}
	cols := dao.SystemConfig.Columns()
	_, err := dao.SystemConfig.Ctx(ctx).TX(tx).
		Data(do.SystemConfig{
			Key:         key,
			Value:       storedValue,
			ValueType:   valueType,
			Description: description,
			IsEncrypted: isEncrypted,
		}).
		OnConflict(cols.Key).
		Save()
	return gerror.Wrapf(err, "upsert setup config %s", key)
}

func insertSetupAuditTx(ctx context.Context, tx gdb.TX, userID string, in service.CompleteSetupInput, webBaseURL string) error {
	metadata, err := marshalJSON(map[string]any{
		"email":          strings.ToLower(strings.TrimSpace(in.Admin.Email)),
		"web_base_url":   webBaseURL,
		"generated_keys": []string{"auth.session.secret", "auth.apiKey.secret"},
	})
	if err != nil {
		return err
	}
	_, err = dao.AuditLogs.Ctx(ctx).TX(tx).Data(do.AuditLogs{
		UserId:       userID,
		Action:       actionSetupCompleted,
		ResourceType: "system_setup",
		ResourceId:   "1",
		Ip:           nilIfEmpty(in.IP),
		UserAgent:    nilIfEmpty(in.UserAgent),
		Metadata:     gjson.New(metadata),
	}).Insert()
	return gerror.Wrap(err, "insert setup completion audit log")
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, "admin email is required")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || strings.ToLower(addr.Address) != email {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "admin email is invalid")
	}
	return email, nil
}

func normalizeWebBaseURL(value string) (string, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, "web base URL is required")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "web base URL must be an absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "web base URL must use http or https")
	}
	return value, nil
}

func validateSetupPassword(ctx context.Context, password string) error {
	minLen := service.Config().GetInt(ctx, "auth.password.minLength", defaultPasswordMinLen)
	if minLen < defaultPasswordMinLen {
		minLen = defaultPasswordMinLen
	}
	if len(password) < minLen {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "admin password must be at least %d characters", minLen)
	}
	return nil
}

func passwordCost(ctx context.Context) int {
	cost := service.Config().GetInt(ctx, "auth.password.bcryptCost", defaultBcryptCost)
	if cost < bcrypt.MinCost || cost > 16 {
		return defaultBcryptCost
	}
	return cost
}

func generateSecret() (string, error) {
	buf := make([]byte, minGeneratedSecretByteCount)
	if _, err := rand.Read(buf); err != nil {
		return "", gerror.Wrap(err, "generate setup secret")
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func isSafeSecret(value string) bool {
	value = strings.TrimSpace(value)
	return len(value) >= 32 && !containsChangeMe(value)
}

func containsChangeMe(value string) bool {
	return strings.Contains(strings.ToLower(value), "change-me")
}

func marshalJSON(in map[string]any) (string, error) {
	payload, err := json.Marshal(in)
	if err != nil {
		return "", gerror.Wrap(err, "marshal setup metadata")
	}
	return string(payload), nil
}

func nilIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func invalidateConfigKeys(ctx context.Context, keys ...string) {
	adapter := credis.GetCacheManager().GetAdapter("default")
	if adapter == nil {
		return
	}
	for _, key := range keys {
		if _, err := adapter.Del(ctx, fmt.Sprintf(configCacheKeyFmt, key)); err != nil {
			g.Log().Warningf(ctx, "setup config cache del[%s] failed: %v", key, err)
		}
	}
}
