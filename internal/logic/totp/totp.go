package totp

import (
	"context"
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	otp "github.com/pquerna/otp"
	ptotp "github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/do"
	"multi-tenant-saas/internal/model/entity"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/crypto"
	credis "multi-tenant-saas/utility/redis"
)

const (
	defaultTOTPIssuer       = "Multi-Tenant-SaaS"
	defaultBackupCodeCount  = 10
	backupCodeChars         = "abcdefghijklmnopqrstuvwxyz0123456789"
	backupCodeLen           = 8
	backupCodeBcryptCost    = 12
	defaultRateLimitWindow  = 5 * time.Minute
	defaultRateLimitAttempt = 5
)

var (
	codeConflict        = gcode.New(409001, "Conflict", nil)
	codeTooManyAttempts = gcode.New(429001, "Too Many Requests", nil)
	totpCodePattern     = regexp.MustCompile(`^[0-9]{6}$`)
	backupCodePattern   = regexp.MustCompile(`^[a-z0-9]{8}$`)
)

type sTOTP struct{}

func init() {
	service.RegisterTOTP(&sTOTP{})
}

func (s *sTOTP) Setup(ctx context.Context, userID, password string) (*service.TOTPSetupResult, error) {
	if !s.featureEnabled(ctx) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "TOTP is disabled")
	}
	if err := s.verifyPassword(ctx, userID, password); err != nil {
		return nil, err
	}
	enabled, err := s.IsEnabled(ctx, userID)
	if err != nil {
		return nil, err
	}
	if enabled {
		return nil, gerror.NewCode(codeConflict, "TOTP is already enabled, please disable first")
	}
	user, err := service.UserService().GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	issuer := strings.TrimSpace(service.Config().GetString(ctx, "auth.totp.issuer", defaultTOTPIssuer))
	if issuer == "" {
		issuer = defaultTOTPIssuer
	}
	key, err := ptotp.Generate(ptotp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user.Email,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "generate TOTP key")
	}
	encrypted, err := crypto.Encrypt(key.Secret())
	if err != nil {
		return nil, gerror.Wrap(err, "encrypt TOTP secret")
	}
	cols := dao.UserTotpConfigs.Columns()
	_, err = dao.UserTotpConfigs.Ctx(ctx).
		Data(do.UserTotpConfigs{
			UserId:          userID,
			SecretEncrypted: encrypted,
			Enabled:         false,
			EnabledAt:       gdb.Raw("NULL"),
		}).
		OnConflict(cols.UserId).
		Save()
	if err != nil {
		return nil, gerror.Wrap(err, "save TOTP config")
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.totp.setup_initiated",
		ResourceType: "user",
		ResourceID:   userID,
	})
	return &service.TOTPSetupResult{Secret: key.Secret(), URL: key.URL()}, nil
}

func (s *sTOTP) Enable(ctx context.Context, userID, code string) ([]string, error) {
	if !s.featureEnabled(ctx) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "TOTP is disabled")
	}
	config, err := s.getConfig(ctx, userID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "TOTP setup has not been initiated")
	}
	if config.Enabled {
		return nil, gerror.NewCode(codeConflict, "TOTP is already enabled")
	}
	if err = s.verifyTOTPCode(ctx, config, code); err != nil {
		return nil, err
	}
	var backupCodes []string
	err = dao.UserTotpConfigs.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.UserTotpConfigs.Columns()
		now := time.Now().UTC()
		result, err := dao.UserTotpConfigs.Ctx(ctx).TX(tx).
			Where(cols.UserId, userID).
			Where(cols.Enabled, false).
			Data(do.UserTotpConfigs{Enabled: true, EnabledAt: now}).
			Update()
		if err != nil {
			return gerror.Wrap(err, "enable TOTP")
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return gerror.Wrap(err, "check enabled TOTP rows affected")
		}
		if affected != 1 {
			return gerror.NewCode(codeConflict, "TOTP is already enabled")
		}
		codes, err := s.generateAndStoreBackupCodesTx(ctx, tx, userID)
		if err != nil {
			return err
		}
		backupCodes = codes
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.totp.enabled",
		ResourceType: "user_totp_config",
		ResourceID:   config.Id.String(),
	})
	if err = s.revokeSessionsAfterChange(ctx, userID, "totp_enabled"); err != nil {
		return nil, err
	}
	return backupCodes, nil
}

func (s *sTOTP) Disable(ctx context.Context, userID, password, code string) error {
	if strings.TrimSpace(password) == "" && strings.TrimSpace(code) == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "Either password or code is required")
	}
	config, err := s.getConfig(ctx, userID)
	if err != nil {
		return err
	}
	if config == nil || !config.Enabled {
		return gerror.NewCode(gcode.CodeNotFound, "TOTP is not enabled")
	}
	if strings.TrimSpace(password) != "" {
		if err = s.verifyPassword(ctx, userID, password); err != nil {
			return err
		}
	}
	if strings.TrimSpace(code) != "" {
		if err = s.verifyTOTPCode(ctx, config, code); err != nil {
			return err
		}
	}
	err = dao.UserTotpConfigs.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.UserTotpConfigs.Columns()
		if _, err := dao.UserTotpConfigs.Ctx(ctx).TX(tx).
			Where(cols.UserId, userID).
			Data(do.UserTotpConfigs{Enabled: false, EnabledAt: gdb.Raw("NULL")}).
			Update(); err != nil {
			return gerror.Wrap(err, "disable TOTP")
		}
		if _, err := dao.UserTotpBackupCodes.Ctx(ctx).TX(tx).
			Where(dao.UserTotpBackupCodes.Columns().UserId, userID).
			Delete(); err != nil {
			return gerror.Wrap(err, "delete backup codes")
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.totp.disabled",
		ResourceType: "user_totp_config",
		ResourceID:   config.Id.String(),
	})
	if err = s.revokeSessionsAfterChange(ctx, userID, "totp_disabled"); err != nil {
		return err
	}
	return nil
}

func (s *sTOTP) ValidateTOTP(ctx context.Context, userID, code string) (*service.TOTPVerifyResult, error) {
	if !s.featureEnabled(ctx) {
		return &service.TOTPVerifyResult{Valid: false}, nil
	}
	if err := s.checkVerifyAttempts(ctx, userID); err != nil {
		return nil, err
	}
	config, err := s.getConfig(ctx, userID)
	if err != nil {
		return nil, err
	}
	if config == nil || !config.Enabled {
		s.recordVerifyAttempt(ctx, userID, false)
		return &service.TOTPVerifyResult{Valid: false}, nil
	}
	if s.verifyTOTPCodeValue(ctx, config, code) {
		s.recordVerifyAttempt(ctx, userID, true)
		return &service.TOTPVerifyResult{Valid: true}, nil
	}
	result, err := s.validateBackupCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	s.recordVerifyAttempt(ctx, userID, result.Valid)
	return result, nil
}

func (s *sTOTP) IsEnabled(ctx context.Context, userID string) (bool, error) {
	if !s.featureEnabled(ctx) {
		return false, nil
	}
	config, err := s.getConfig(ctx, userID)
	if err != nil || config == nil {
		return false, err
	}
	return config.Enabled, nil
}

func (s *sTOTP) GetStatus(ctx context.Context, userID string) (*service.TOTPStatus, error) {
	config, err := s.getConfig(ctx, userID)
	if err != nil {
		return nil, err
	}
	status := &service.TOTPStatus{}
	if config == nil {
		return status, nil
	}
	status.SetupInitiated = true
	status.Enabled = config.Enabled
	if !config.EnabledAt.IsZero() {
		enabledAt := config.EnabledAt
		status.EnabledAt = &enabledAt
	}
	if !config.CreatedAt.IsZero() {
		createdAt := config.CreatedAt
		status.CreatedAt = &createdAt
	}
	count, err := dao.UserTotpBackupCodes.Ctx(ctx).
		Where(dao.UserTotpBackupCodes.Columns().UserId, userID).
		Where(dao.UserTotpBackupCodes.Columns().UsedAt + " IS NULL").
		Count()
	if err != nil {
		return nil, gerror.Wrap(err, "count backup codes")
	}
	status.BackupCodesRemaining = count
	return status, nil
}

func (s *sTOTP) RegenerateBackupCodes(ctx context.Context, userID, password string) ([]string, error) {
	if err := s.verifyPassword(ctx, userID, password); err != nil {
		return nil, err
	}
	enabled, err := s.IsEnabled(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, gerror.NewCode(gcode.CodeNotFound, "TOTP is not enabled")
	}
	var backupCodes []string
	err = dao.UserTotpBackupCodes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		codes, err := s.generateAndStoreBackupCodesTx(ctx, tx, userID)
		if err != nil {
			return err
		}
		backupCodes = codes
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.totp.backup_codes_regenerated",
		ResourceType: "user",
		ResourceID:   userID,
	})
	if err = s.revokeSessionsAfterChange(ctx, userID, "totp_backup_codes_regenerated"); err != nil {
		return nil, err
	}
	return backupCodes, nil
}

func (s *sTOTP) revokeSessionsAfterChange(ctx context.Context, userID, reason string) error {
	if !service.Config().GetBool(ctx, "auth.totp.revokeSessionsOnChange", false) {
		return nil
	}
	if err := service.AuthSessionService().RevokeUserSessions(ctx, userID, "", reason); err != nil {
		return err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.totp.sessions_revoked",
		ResourceType: "auth_session",
		Metadata:     map[string]any{"reason": reason},
	})
	return nil
}

func (s *sTOTP) featureEnabled(ctx context.Context) bool {
	return service.Config().GetBool(ctx, "auth.totp.enabled", false)
}

func (s *sTOTP) getConfig(ctx context.Context, userID string) (*entity.UserTotpConfigs, error) {
	var config *entity.UserTotpConfigs
	err := dao.UserTotpConfigs.Ctx(ctx).
		Where(dao.UserTotpConfigs.Columns().UserId, userID).
		Scan(&config)
	if err != nil {
		return nil, gerror.Wrap(err, "select TOTP config")
	}
	return config, nil
}

func (s *sTOTP) verifyPassword(ctx context.Context, userID, password string) error {
	if strings.TrimSpace(password) == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "password is required")
	}
	cols := dao.UserPasswordCredentials.Columns()
	record, err := dao.UserPasswordCredentials.Ctx(ctx).
		Fields(cols.PasswordHash).
		Where(cols.UserId, userID).
		One()
	if err != nil {
		return gerror.Wrap(err, "select password credential")
	}
	if record.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotAuthorized, "Invalid password")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(record[cols.PasswordHash].String()), []byte(password)); err != nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "Invalid password")
	}
	return nil
}

func (s *sTOTP) verifyTOTPCode(ctx context.Context, config *entity.UserTotpConfigs, code string) error {
	if strings.TrimSpace(code) == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "code is required")
	}
	if !s.verifyTOTPCodeValue(ctx, config, code) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "Invalid TOTP code")
	}
	return nil
}

func (s *sTOTP) verifyTOTPCodeValue(ctx context.Context, config *entity.UserTotpConfigs, code string) bool {
	code = normalizeTOTPCode(code)
	if !totpCodePattern.MatchString(code) {
		return false
	}
	secret, err := crypto.Decrypt(config.SecretEncrypted)
	if err != nil {
		g.Log().Errorf(ctx, "decrypt TOTP secret for user %s failed: %v", config.UserId.String(), err)
		return false
	}
	valid, err := ptotp.ValidateCustom(code, secret, time.Now().UTC(), ptotp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return err == nil && valid
}

func (s *sTOTP) generateAndStoreBackupCodesTx(ctx context.Context, tx gdb.TX, userID string) ([]string, error) {
	backupCols := dao.UserTotpBackupCodes.Columns()
	if _, err := dao.UserTotpBackupCodes.Ctx(ctx).TX(tx).
		Where(backupCols.UserId, userID).
		Delete(); err != nil {
		return nil, gerror.Wrap(err, "delete old backup codes")
	}
	count := service.Config().GetInt(ctx, "auth.totp.backupCodeCount", defaultBackupCodeCount)
	if count <= 0 || count > 50 {
		count = defaultBackupCodeCount
	}
	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code, err := generateRandomBackupCode()
		if err != nil {
			return nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(code), backupCodeBcryptCost)
		if err != nil {
			return nil, gerror.Wrap(err, "hash backup code")
		}
		if _, err = dao.UserTotpBackupCodes.Ctx(ctx).TX(tx).Data(do.UserTotpBackupCodes{
			UserId:   userID,
			CodeHash: string(hash),
		}).Insert(); err != nil {
			return nil, gerror.Wrap(err, "insert backup code")
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func (s *sTOTP) validateBackupCode(ctx context.Context, userID, code string) (*service.TOTPVerifyResult, error) {
	code = normalizeBackupCode(code)
	if !backupCodePattern.MatchString(code) {
		return &service.TOTPVerifyResult{Valid: false}, nil
	}
	backupCols := dao.UserTotpBackupCodes.Columns()
	var result service.TOTPVerifyResult
	var matchedID string
	err := dao.UserTotpBackupCodes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var unusedCodes []entity.UserTotpBackupCodes
		err := dao.UserTotpBackupCodes.Ctx(ctx).TX(tx).
			Where(backupCols.UserId, userID).
			Where(backupCols.UsedAt + " IS NULL").
			Scan(&unusedCodes)
		if err != nil {
			return gerror.Wrap(err, "query backup codes")
		}
		for _, candidate := range unusedCodes {
			if bcrypt.CompareHashAndPassword([]byte(candidate.CodeHash), []byte(code)) != nil {
				continue
			}
			updateResult, err := dao.UserTotpBackupCodes.Ctx(ctx).TX(tx).
				Where(backupCols.Id, candidate.Id.String()).
				Where(backupCols.UsedAt + " IS NULL").
				Data(do.UserTotpBackupCodes{UsedAt: time.Now().UTC()}).
				Update()
			if err != nil {
				return gerror.Wrap(err, "mark backup code used")
			}
			affected, err := updateResult.RowsAffected()
			if err != nil {
				return gerror.Wrap(err, "check backup code rows affected")
			}
			if affected != 1 {
				continue
			}
			remaining, err := dao.UserTotpBackupCodes.Ctx(ctx).TX(tx).
				Where(backupCols.UserId, userID).
				Where(backupCols.UsedAt + " IS NULL").
				Count()
			if err != nil {
				return gerror.Wrap(err, "count remaining backup codes")
			}
			matchedID = candidate.Id.String()
			result = service.TOTPVerifyResult{Valid: true, UsedBackupCode: true, BackupCodesRemaining: remaining}
			return nil
		}
		result = service.TOTPVerifyResult{Valid: false}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if result.Valid && matchedID != "" {
		_ = service.Audit().Write(ctx, service.AuditLogInput{
			UserID:       userID,
			Action:       "auth.totp.backup_code_used",
			ResourceType: "user_totp_backup_code",
			ResourceID:   matchedID,
			Metadata:     map[string]any{"remaining": result.BackupCodesRemaining},
		})
	}
	return &result, nil
}

func normalizeTOTPCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")
	return code
}

func normalizeBackupCode(code string) string {
	code = strings.TrimSpace(strings.ToLower(code))
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")
	return code
}

func generateRandomBackupCode() (string, error) {
	buf := make([]byte, backupCodeLen)
	max := big.NewInt(int64(len(backupCodeChars)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", gerror.Wrap(err, "generate backup code")
		}
		buf[i] = backupCodeChars[n.Int64()]
	}
	return string(buf), nil
}

type verifyRateLimiter interface {
	Check(ctx context.Context, userID string, maxAttempts int, window time.Duration) (retryAfter time.Duration, err error)
	Record(ctx context.Context, userID string, success bool, window time.Duration) error
}

type verifyAttemptState struct {
	Count   int
	ResetAt time.Time
}

var verifyAttempts = struct {
	sync.Mutex
	items map[string]verifyAttemptState
}{items: map[string]verifyAttemptState{}}

func (s *sTOTP) checkVerifyAttempts(ctx context.Context, userID string) error {
	maxAttempts, window := s.verifyAttemptConfig(ctx)
	retryAfter, err := s.verifyLimiter(ctx).Check(ctx, userID, maxAttempts, window)
	if err != nil {
		g.Log().Warningf(ctx, "TOTP shared verify limiter failed, falling back to process limiter: %v", err)
		retryAfter, err = processVerifyLimiter.Check(ctx, userID, maxAttempts, window)
	}
	if err != nil {
		return err
	}
	if retryAfter > 0 {
		return gerror.NewCode(codeTooManyAttempts, "Too many attempts, please try again later")
	}
	return nil
}

func (s *sTOTP) recordVerifyAttempt(ctx context.Context, userID string, success bool) {
	if userID == "" {
		return
	}
	_, window := s.verifyAttemptConfig(ctx)
	if err := s.verifyLimiter(ctx).Record(ctx, userID, success, window); err != nil {
		g.Log().Warningf(ctx, "TOTP shared verify limiter record failed, falling back to process limiter: %v", err)
		_ = processVerifyLimiter.Record(ctx, userID, success, window)
	}
}

func (s *sTOTP) verifyLimiter(ctx context.Context) verifyRateLimiter {
	if adapter := credis.GetCacheManager().GetAdapter("default"); adapter != nil {
		return redisVerifyLimiter{adapter: adapter}
	}
	return processVerifyLimiter
}

func (s *sTOTP) verifyAttemptConfig(ctx context.Context) (int, time.Duration) {
	maxAttempts := service.Config().GetInt(ctx, "auth.totp.rateLimitAttempts", defaultRateLimitAttempt)
	if maxAttempts <= 0 {
		maxAttempts = defaultRateLimitAttempt
	}
	window := service.Config().GetDuration(ctx, "auth.totp.rateLimitWindow", defaultRateLimitWindow)
	if window <= 0 {
		window = defaultRateLimitWindow
	}
	return maxAttempts, window
}

type memoryVerifyLimiter struct{}

var processVerifyLimiter memoryVerifyLimiter

func (memoryVerifyLimiter) Check(_ context.Context, userID string, maxAttempts int, _ time.Duration) (time.Duration, error) {
	now := time.Now().UTC()
	verifyAttempts.Lock()
	defer verifyAttempts.Unlock()
	state := verifyAttempts.items[userID]
	if state.ResetAt.IsZero() || now.After(state.ResetAt) {
		delete(verifyAttempts.items, userID)
		return 0, nil
	}
	if state.Count >= maxAttempts {
		return state.ResetAt.Sub(now), nil
	}
	return 0, nil
}

func (memoryVerifyLimiter) Record(_ context.Context, userID string, success bool, window time.Duration) error {
	if success {
		verifyAttempts.Lock()
		delete(verifyAttempts.items, userID)
		verifyAttempts.Unlock()
		return nil
	}
	now := time.Now().UTC()
	verifyAttempts.Lock()
	defer verifyAttempts.Unlock()
	state := verifyAttempts.items[userID]
	if state.ResetAt.IsZero() || now.After(state.ResetAt) {
		state = verifyAttemptState{ResetAt: now.Add(window)}
	}
	state.Count++
	verifyAttempts.items[userID] = state
	return nil
}

type redisVerifyLimiter struct {
	adapter *credis.RedisAdapter
}

func (l redisVerifyLimiter) Check(ctx context.Context, userID string, maxAttempts int, window time.Duration) (time.Duration, error) {
	key := totpVerifyAttemptKey(userID)
	count, err := l.adapter.GetInt64(ctx, key)
	if err != nil {
		return 0, err
	}
	if count < int64(maxAttempts) {
		return 0, nil
	}
	ttl, err := l.adapter.TTL(ctx, key)
	if err != nil {
		return 0, err
	}
	if ttl <= 0 {
		ttl = window
	}
	return ttl, nil
}

func (l redisVerifyLimiter) Record(ctx context.Context, userID string, success bool, window time.Duration) error {
	key := totpVerifyAttemptKey(userID)
	if success {
		_, err := l.adapter.Del(ctx, key)
		return err
	}
	_, err := l.adapter.IncrementWithTTL(ctx, key, window)
	return err
}

func totpVerifyAttemptKey(userID string) string {
	return "totp:verify_attempts:" + userID
}
