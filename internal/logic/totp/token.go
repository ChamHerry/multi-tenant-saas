package totp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"multi-tenant-saas/internal/service"
)

const defaultTOTPTokenTTL = 5 * time.Minute

func (s *sTOTP) GenerateToken(ctx context.Context, userID string) (string, error) {
	secret, err := totpTokenSecret(ctx)
	if err != nil {
		return "", err
	}
	ttl := service.Config().GetDuration(ctx, "auth.totp.tokenTTL", defaultTOTPTokenTTL)
	if ttl <= 0 {
		ttl = defaultTOTPTokenTTL
	}
	return generateTokenWithSecret(userID, time.Now().UTC(), ttl, secret), nil
}

func (s *sTOTP) ValidateToken(ctx context.Context, token string) (string, error) {
	secret, err := totpTokenSecret(ctx)
	if err != nil {
		return "", err
	}
	return validateTokenWithSecret(token, time.Now().UTC(), secret)
}

func generateTokenWithSecret(userID string, now time.Time, ttl time.Duration, secret []byte) string {
	expiresAt := now.Add(ttl).Unix()
	expiresBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(expiresBytes, uint64(expiresAt))

	userPart := base64.RawURLEncoding.EncodeToString([]byte(userID))
	expirePart := base64.RawURLEncoding.EncodeToString(expiresBytes)
	sigPart := base64.RawURLEncoding.EncodeToString(signTOTPToken([]byte(userID), expiresBytes, secret))
	return fmt.Sprintf("%s.%s.%s", userPart, expirePart, sigPart)
}

func validateTokenWithSecret(token string, now time.Time, secret []byte) (string, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "TOTP token is invalid or expired")
	}
	userBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(userBytes) == 0 {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "TOTP token is invalid or expired")
	}
	expiresBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(expiresBytes) != 8 {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "TOTP token is invalid or expired")
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "TOTP token is invalid or expired")
	}
	if !hmac.Equal(sigBytes, signTOTPToken(userBytes, expiresBytes, secret)) {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "TOTP token is invalid or expired")
	}
	expiresAt := int64(binary.BigEndian.Uint64(expiresBytes))
	if now.Unix() > expiresAt {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "TOTP token is invalid or expired")
	}
	return string(userBytes), nil
}

func signTOTPToken(userBytes, expiresBytes, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(userBytes)
	mac.Write(expiresBytes)
	return mac.Sum(nil)
}

func totpTokenSecret(ctx context.Context) ([]byte, error) {
	sessionSecret := strings.TrimSpace(service.Config().GetString(ctx, "auth.session.secret", ""))
	if sessionSecret == "" {
		return nil, gerror.NewCode(gcode.CodeMissingConfiguration, "auth.session.secret is required")
	}
	mac := hmac.New(sha256.New, []byte("totp-token-v1"))
	mac.Write([]byte(sessionSecret))
	return mac.Sum(nil), nil
}

func sameBytes(a, b []byte) bool {
	return bytes.Equal(a, b)
}
