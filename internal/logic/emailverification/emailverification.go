package emailverification

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/do"
	"multi-tenant-saas/internal/service"
)

const (
	passwordProvider      = "password"
	defaultTokenTTL       = 24 * time.Hour
	defaultResendCooldown = 1 * time.Minute
)

var (
	codeTokenExpired    = gcode.New(410, "TokenExpired", nil)
	codeTokenNotFound   = gcode.New(404, "TokenNotFound", nil)
	codeAlreadyVerified = gcode.New(409, "EmailAlreadyVerified", nil)
	codeRateLimited     = gcode.New(429, "RateLimited", nil)
)

type sEmailVerification struct{}

func init() {
	service.RegisterEmailVerification(&sEmailVerification{})
}

// CreateAndSend creates a verification token and sends the verification email asynchronously.
func (s *sEmailVerification) CreateAndSend(ctx context.Context, in service.CreateVerificationTokenInput) error {
	if in.UserID == "" || in.Email == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "userID and email are required")
	}

	// Generate random token
	rawToken, err := randomToken()
	if err != nil {
		return err
	}

	// Hash the token for storage
	secret, err := sessionSecret(ctx)
	if err != nil {
		return err
	}
	tokenHash := hashToken(rawToken, secret)

	// Calculate expiry
	ttl := tokenTTL(ctx)
	expiresAt := time.Now().Add(ttl)

	// Store token in DB
	_, err = dao.EmailVerificationTokens.Ctx(ctx).Data(do.EmailVerificationTokens{
		UserId:    in.UserID,
		Email:     strings.ToLower(strings.TrimSpace(in.Email)),
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "insert verification token")
	}

	// Build verification URL and send email asynchronously
	baseURL := strings.TrimRight(service.Config().GetString(ctx, "email.verification.base_url", defaultBaseURL(ctx)), "/")
	verifyURL := baseURL + "/verify-email?token=" + rawToken

	go func(toEmail, displayName, url string) {
		sendCtx := context.Background()
		if err := service.Email().SendEmailVerification(sendCtx, service.SendEmailVerificationInput{
			ToEmail:     toEmail,
			DisplayName: displayName,
			VerifyURL:   url,
		}); err != nil {
			g.Log().Errorf(sendCtx, "[email-verification] failed to send verification email to %s: %v", toEmail, err)
		}
	}(in.Email, "", verifyURL)

	return nil
}

// Verify validates the token and marks the email as verified.
func (s *sEmailVerification) Verify(ctx context.Context, in service.VerifyEmailInput) error {
	token := strings.TrimSpace(in.Token)
	if token == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "token is required")
	}

	// Hash the token
	secret, err := sessionSecret(ctx)
	if err != nil {
		return err
	}
	tokenHash := hashToken(token, secret)

	// Look up the token
	cols := dao.EmailVerificationTokens.Columns()
	record, err := dao.EmailVerificationTokens.Ctx(ctx).
		Where(cols.TokenHash, tokenHash).
		WhereNull(cols.UsedAt).
		Where(cols.ExpiresAt + " > NOW()").
		One()
	if err != nil {
		return gerror.Wrap(err, "select verification token")
	}
	if record.IsEmpty() {
		// Token not found — could be already used, expired, or invalid.
		// Check if it exists but expired to give a better error.
		expiredRecord, _ := dao.EmailVerificationTokens.Ctx(ctx).
			Where(cols.TokenHash, tokenHash).
			WhereNull(cols.UsedAt).
			One()
		if !expiredRecord.IsEmpty() {
			return gerror.NewCode(codeTokenExpired, "verification token has expired")
		}
		return gerror.NewCode(codeTokenNotFound, "verification token not found or already used")
	}

	userID := record[cols.UserId].String()
	email := record[cols.Email].String()

	// Mark token as used
	_, err = dao.EmailVerificationTokens.Ctx(ctx).
		Where(cols.Id, record[cols.Id].String()).
		Data(do.EmailVerificationTokens{UsedAt: "now()"}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "mark verification token used")
	}

	// Check if already verified (idempotent)
	identity, err := service.UserService().GetIdentityByEmail(ctx, passwordProvider, email)
	if err != nil {
		return err
	}
	if identity != nil && identity.EmailVerified {
		return nil // Already verified, idempotent success
	}

	// Set email as verified
	if err = service.UserService().SetEmailVerified(ctx, userID, email); err != nil {
		return err
	}

	// Write audit log
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       "auth.email.verified",
		ResourceType: "user_identity",
		ResourceID:   email,
		Metadata:     map[string]any{"email": email},
	})

	return nil
}

// ResendVerification resends the verification email with rate limiting.
func (s *sEmailVerification) ResendVerification(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "email is required")
	}

	// Look up the password identity (information-leak safe: return success regardless)
	identity, err := service.UserService().GetIdentityByEmail(ctx, passwordProvider, email)
	if err != nil {
		g.Log().Warningf(ctx, "[email-verification] resend lookup error for %s: %v", email, err)
		return nil // Return success to prevent email enumeration
	}
	if identity == nil {
		return nil // No such identity — return success to prevent enumeration
	}
	if identity.EmailVerified {
		return nil // Already verified — return success to prevent enumeration
	}

	// Check rate limit: count tokens created in the last cooldown period
	cols := dao.EmailVerificationTokens.Columns()
	cooldown := resendCooldown(ctx)
	count, err := dao.EmailVerificationTokens.Ctx(ctx).
		Where(cols.Email, email).
		Where(cols.CreatedAt+" > NOW() - INTERVAL '1 second' * ?", int(cooldown.Seconds())).
		Count()
	if err != nil {
		return gerror.Wrap(err, "check verification rate limit")
	}
	if count > 0 {
		return gerror.NewCode(codeRateLimited, "please wait before requesting another verification email")
	}

	// Create and send new token
	return s.CreateAndSend(ctx, service.CreateVerificationTokenInput{
		UserID: identity.UserID,
		Email:  email,
	})
}

// ---- helpers ----

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

func tokenTTL(ctx context.Context) time.Duration {
	ttl := service.Config().GetDuration(ctx, "email.verification.token_ttl", defaultTokenTTL)
	if ttl <= 0 {
		ttl = defaultTokenTTL
	}
	return ttl
}

func resendCooldown(ctx context.Context) time.Duration {
	cooldown := service.Config().GetDuration(ctx, "email.verification.resend_cooldown", defaultResendCooldown)
	if cooldown <= 0 {
		cooldown = defaultResendCooldown
	}
	return cooldown
}

func defaultBaseURL(ctx context.Context) string {
	// Try to construct from server config
	addr := service.Config().GetString(ctx, "server.address", ":8000")
	host := "http://127.0.0.1:8000"
	if addr != "" {
		if strings.HasPrefix(addr, ":") {
			host = "http://127.0.0.1" + addr
		} else {
			host = "http://" + addr
		}
	}
	return host
}
