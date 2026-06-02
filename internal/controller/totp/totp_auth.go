package totp

import (
	"context"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	totpapi "multi-tenant-saas/api/totp"
	"multi-tenant-saas/api/totp/v1"
	"multi-tenant-saas/internal/service"
)

type AuthPublicControllerV1 struct{}

func NewAuthPublicV1() totpapi.IAuthPublicV1 {
	return &AuthPublicControllerV1{}
}

func (c *AuthPublicControllerV1) VerifyTOTP(ctx context.Context, req *v1.VerifyTOTPReq) (res *v1.AuthUserRes, err error) {
	userID, err := service.TOTP().ValidateToken(ctx, req.TOTPToken)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "TOTP token is invalid or expired")
	}
	result, err := service.TOTP().ValidateTOTP(ctx, userID, req.Code)
	if err != nil {
		if gerror.Code(err).Code() == 429001 {
			setTOTPVerifyRetryAfterHeader(ctx)
		}
		return nil, err
	}
	if result == nil || !result.Valid {
		_ = service.Audit().Write(ctx, service.AuditLogInput{
			UserID:       userID,
			Action:       "auth.totp.failed",
			ResourceType: "auth",
		})
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "Invalid TOTP code")
	}
	user, err := service.UserService().GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	session, cookies, err := service.AuthSessionService().Create(
		ctx,
		userID,
		service.BizCtx().GetUserAgent(ctx),
		service.BizCtx().GetClientIP(ctx),
	)
	if err != nil {
		return nil, err
	}
	if err = service.LoginSuccess().Complete(ctx, service.CompleteLoginSuccessInput{
		UserID:    userID,
		Email:     user.Email,
		SessionID: session.ID,
		IP:        service.BizCtx().GetClientIP(ctx),
		UserAgent: service.BizCtx().GetUserAgent(ctx),
	}); err != nil {
		return nil, err
	}
	setAuthCookies(ghttp.RequestFromCtx(ctx), cookies)
	action := "auth.totp.verified"
	if result.UsedBackupCode {
		action = "auth.totp.backup_code_used"
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       userID,
		Action:       action,
		ResourceType: "auth_session",
		ResourceID:   session.ID,
		Metadata: map[string]any{
			"backup_code_used":       result.UsedBackupCode,
			"backup_codes_remaining": result.BackupCodesRemaining,
		},
	})
	var remaining *int
	if result.UsedBackupCode {
		remaining = &result.BackupCodesRemaining
	}
	return &v1.AuthUserRes{User: user, BackupCodesRemaining: remaining}, nil
}

func setTOTPVerifyRetryAfterHeader(ctx context.Context) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return
	}
	window := service.Config().GetDuration(ctx, "auth.totp.rateLimitWindow", 5*time.Minute)
	if window <= 0 {
		window = 5 * time.Minute
	}
	r.Response.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
}

func setAuthCookies(r *ghttp.Request, cookies *service.AuthSessionCookies) {
	if r == nil || cookies == nil {
		return
	}
	if cookies.Session != nil {
		r.Cookie.SetHttpCookie(cookies.Session)
	}
	if cookies.CSRF != nil {
		r.Cookie.SetHttpCookie(cookies.CSRF)
	}
}
