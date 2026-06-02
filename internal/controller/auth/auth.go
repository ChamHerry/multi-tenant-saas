package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	authapi "multi-tenant-saas/api/auth"
	"multi-tenant-saas/api/auth/v1"
	"multi-tenant-saas/internal/service"
)

type PublicControllerV1 struct{}
type ControllerV1 struct{}

func NewPublicV1() authapi.IAuthPublicV1 {
	return &PublicControllerV1{}
}

func NewV1() authapi.IAuthV1 {
	return &ControllerV1{}
}

func (c *PublicControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.AuthUserRes, err error) {
	if !service.Config().GetBool(ctx, "auth.password.enabled", true) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "password login is disabled")
	}
	// IP and UserAgent are now auto-filled from BizContext in Audit.Write.
	login, err := service.PasswordAuth().Login(ctx, service.PasswordLoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IP:        service.BizCtx().GetClientIP(ctx),
		UserAgent: service.BizCtx().GetUserAgent(ctx),
	})
	if err != nil {
		return nil, err
	}
	if login.Requires2FA {
		return &v1.AuthUserRes{User: login.User, Requires2FA: true, TOTPToken: login.TOTPToken}, nil
	}
	setAuthCookies(ghttp.RequestFromCtx(ctx), login.Cookies)
	return &v1.AuthUserRes{User: login.User}, nil
}

func (c *PublicControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.AuthUserRes, err error) {
	if !service.Config().GetBool(ctx, "auth.password.registrationEnabled", false) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "password registration is disabled")
	}
	user, err := service.PasswordAuth().RegisterPasswordUser(ctx, service.RegisterPasswordUserInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		return nil, err
	}
	if _, err = service.TenantMembershipService().ListUserTenants(ctx, user.ID); err != nil {
		return nil, err
	}
	login, err := service.PasswordAuth().Login(ctx, service.PasswordLoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IP:        service.BizCtx().GetClientIP(ctx),
		UserAgent: service.BizCtx().GetUserAgent(ctx),
	})
	if err == nil {
		setAuthCookies(ghttp.RequestFromCtx(ctx), login.Cookies)
		user = login.User
	}

	// Send verification email asynchronously (does not block the registration response).
	go func(email, userID string) {
		sendCtx := context.Background()
		if err := service.EmailVerification().CreateAndSend(sendCtx,
			service.CreateVerificationTokenInput{UserID: userID, Email: email},
		); err != nil {
			g.Log().Errorf(sendCtx, "failed to send verification email for %s: %v", email, err)
		}
	}(req.Email, user.ID)

	return &v1.AuthUserRes{User: user}, nil
}

func (c *ControllerV1) Session(ctx context.Context, req *v1.SessionReq) (res *v1.SessionRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	user, err := service.UserService().GetUser(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	var session *service.AuthSession
	if identity.SessionID != "" {
		session, err = service.AuthSessionService().Get(ctx, identity.SessionID)
		if err != nil {
			return nil, err
		}
	}
	return &v1.SessionRes{User: user, Session: session}, nil
}

func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.ActionRes, err error) {
	identity, _ := service.AuthIdentityFromCtx(ctx)
	if identity != nil && identity.SessionID != "" {
		if err = service.AuthSessionService().Revoke(ctx, identity.SessionID, "logout"); err != nil {
			return nil, err
		}
		// IP and UserAgent auto-filled from BizContext.
		_ = service.Audit().Write(ctx, service.AuditLogInput{UserID: identity.UserID, Action: "auth.logout", ResourceType: "auth_session", ResourceID: identity.SessionID})
	}
	setAuthCookies(ghttp.RequestFromCtx(ctx), service.AuthSessionService().ClearCookies(ctx))
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) ChangePassword(ctx context.Context, req *v1.ChangePasswordReq) (res *v1.ActionRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.PasswordAuth().ChangePassword(ctx, identity.UserID, identity.SessionID, req.OldPassword, req.NewPassword); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *PublicControllerV1) VerifyEmail(ctx context.Context, req *v1.VerifyEmailReq) (res *v1.VerifyEmailRes, err error) {
	if err = service.EmailVerification().Verify(ctx, service.VerifyEmailInput{Token: req.Token}); err != nil {
		return nil, err
	}
	return &v1.VerifyEmailRes{OK: true, Message: "Email verified successfully"}, nil
}

func (c *PublicControllerV1) ForgotPassword(ctx context.Context, req *v1.ForgotPasswordReq) (res *v1.ForgotPasswordRes, err error) {
	if !service.Config().GetBool(ctx, "auth.password.enabled", true) {
		// Return OK even when password auth is disabled, to avoid information leakage.
		return &v1.ForgotPasswordRes{OK: true}, nil
	}
	ip := service.BizCtx().GetClientIP(ctx)
	if err = service.PasswordAuth().ForgotPassword(ctx, req.Email, ip); err != nil {
		return nil, err
	}
	return &v1.ForgotPasswordRes{OK: true}, nil
}

func (c *PublicControllerV1) ResetPassword(ctx context.Context, req *v1.ResetPasswordReq) (res *v1.ResetPasswordRes, err error) {
	if !service.Config().GetBool(ctx, "auth.password.enabled", true) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "password authentication is disabled")
	}
	ip := service.BizCtx().GetClientIP(ctx)
	if err = service.PasswordAuth().ResetPassword(ctx, req.Token, req.NewPassword, ip); err != nil {
		return nil, err
	}
	return &v1.ResetPasswordRes{OK: true}, nil
}

func (c *PublicControllerV1) ResendVerification(ctx context.Context, req *v1.ResendVerificationReq) (res *v1.ResendVerificationRes, err error) {
	if err = service.EmailVerification().ResendVerification(ctx, req.Email); err != nil {
		return nil, err
	}
	return &v1.ResendVerificationRes{OK: true, Message: "Verification email sent"}, nil
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
