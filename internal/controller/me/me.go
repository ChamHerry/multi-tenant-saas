package me

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	apime "multi-tenant-saas/api/me"
	"multi-tenant-saas/api/me/v1"
	"multi-tenant-saas/internal/service"
)

type ControllerV1 struct{}

func NewV1() apime.IMeV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error) {
	if err = service.RBAC().RequireAuthScope(ctx, service.PermissionUserRead); err != nil {
		return nil, err
	}
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	user, err := service.UserService().GetUser(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.MeRes{User: user}, nil
}

func (c *ControllerV1) Tenants(ctx context.Context, req *v1.TenantsReq) (res *v1.TenantsRes, err error) {
	if err = service.RBAC().RequireAuthScope(ctx, service.PermissionUserTenantRead); err != nil {
		return nil, err
	}
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	tenants, err := service.TenantMembershipService().ListUserTenants(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.TenantsRes{Tenants: tenants}, nil
}

func (c *ControllerV1) Access(ctx context.Context, req *v1.AccessReq) (res *v1.AccessRes, err error) {
	if err = service.RBAC().RequireAuthScope(ctx, service.PermissionUserTenantRead); err != nil {
		return nil, err
	}
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	access, err := service.Access().Snapshot(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.AccessRes{
		User:          access.User,
		Tenants:       access.Tenants,
		PlatformAdmin: access.PlatformAdmin,
	}, nil
}

func (c *ControllerV1) TenantContext(ctx context.Context, req *v1.TenantContextReq) (res *v1.TenantContextRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionTenantRead); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.TenantContextRes{TenantContext: tc}, nil
}

func (c *ControllerV1) UpdateMe(ctx context.Context, req *v1.UpdateMeReq) (res *v1.UpdateMeRes, err error) {
	if err = service.RBAC().RequireAuthScope(ctx, service.PermissionUserRead); err != nil {
		return nil, err
	}
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	user, err := service.UserService().UpdateProfile(ctx, service.UpdateProfileInput{
		UserID:      identity.UserID,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateMeRes{User: user}, nil
}

func (c *ControllerV1) ListSessions(ctx context.Context, req *v1.ListSessionsReq) (res *v1.ListSessionsRes, err error) {
	identity, err := requireSessionIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.RBAC().RequireAuthScope(ctx, service.PermissionUserSecurityRead); err != nil {
		return nil, err
	}
	sessions, err := service.AuthSessionService().ListUserSessions(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	items := make([]v1.SessionItem, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, toSessionItem(session, identity.SessionID))
	}
	return &v1.ListSessionsRes{Sessions: items}, nil
}

func (c *ControllerV1) RevokeSession(ctx context.Context, req *v1.RevokeSessionReq) (res *v1.RevokeSessionRes, err error) {
	identity, err := requireSessionIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Session) == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "session is required")
	}
	if req.Session == identity.SessionID {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "current session cannot be revoked here; use logout")
	}
	session, err := service.AuthSessionService().Get(ctx, req.Session)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotFound, "session not found")
	}
	if session.UserID != identity.UserID {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "session does not belong to current user")
	}
	now := time.Now().UTC()
	if session.RevokedAt != nil || !session.ExpiresAt.After(now) || !session.IdleExpiresAt.After(now) {
		return nil, gerror.NewCode(gcode.CodeNotFound, "session not found")
	}
	if err = service.AuthSessionService().Revoke(ctx, req.Session, "user_revoke"); err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       identity.UserID,
		Action:       "auth.session.revoke",
		ResourceType: "auth_session",
		ResourceID:   req.Session,
	})
	return &v1.RevokeSessionRes{OK: true}, nil
}

func (c *ControllerV1) RevokeOtherSessions(ctx context.Context, req *v1.RevokeOtherSessionsReq) (res *v1.RevokeOtherSessionsRes, err error) {
	identity, err := requireSessionIdentity(ctx)
	if err != nil {
		return nil, err
	}
	sessions, err := service.AuthSessionService().ListUserSessions(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	revokedCount := 0
	for _, session := range sessions {
		if session.ID != identity.SessionID {
			revokedCount++
		}
	}
	if err = service.AuthSessionService().RevokeUserSessions(ctx, identity.UserID, identity.SessionID, "user_revoke_others"); err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       identity.UserID,
		Action:       "auth.session.revoke_others",
		ResourceType: "auth_session",
		ResourceID:   identity.SessionID,
		Metadata:     map[string]any{"revoked_count": revokedCount},
	})
	return &v1.RevokeOtherSessionsRes{OK: true, RevokedCount: revokedCount}, nil
}

func requireSessionIdentity(ctx context.Context) (*service.AuthIdentity, error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if identity.Type != "session" || strings.TrimSpace(identity.SessionID) == "" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "session authentication is required")
	}
	return identity, nil
}
