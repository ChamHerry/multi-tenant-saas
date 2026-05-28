package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	authapi "repomind-temp/api/auth"
	"repomind-temp/api/auth/v1"
	"repomind-temp/internal/service"
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
	if !g.Cfg().MustGet(ctx, "auth.password.enabled", true).Bool() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "password login is disabled")
	}
	r := ghttp.RequestFromCtx(ctx)
	login, err := service.PasswordAuth().Login(ctx, service.PasswordLoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IP:        requestIP(r),
		UserAgent: requestUserAgent(r),
	})
	if err != nil {
		return nil, err
	}
	setAuthCookies(r, login.Cookies)
	return &v1.AuthUserRes{User: login.User}, nil
}

func (c *PublicControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.AuthUserRes, err error) {
	if !g.Cfg().MustGet(ctx, "auth.password.registrationEnabled", false).Bool() {
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
		IP:        requestIP(ghttp.RequestFromCtx(ctx)),
		UserAgent: requestUserAgent(ghttp.RequestFromCtx(ctx)),
	})
	if err == nil {
		setAuthCookies(ghttp.RequestFromCtx(ctx), login.Cookies)
		user = login.User
	}
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
		_ = service.Audit().Write(ctx, service.AuditLogInput{UserID: identity.UserID, Action: "auth.logout", ResourceType: "auth_session", ResourceID: identity.SessionID, IP: requestIP(ghttp.RequestFromCtx(ctx)), UserAgent: requestUserAgent(ghttp.RequestFromCtx(ctx))})
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

func requestIP(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	return r.GetClientIp()
}

func requestUserAgent(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	return r.Header.Get("User-Agent")
}
