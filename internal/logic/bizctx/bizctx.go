package bizctx

import (
	"context"

	"multi-tenant-saas/internal/service"
)

type sBizCtx struct{}

func init() {
	service.RegisterBizCtx(&sBizCtx{})
}

func (s *sBizCtx) Get(ctx context.Context) *service.BizContext {
	if bc, ok := ctx.Value(service.BizContextKey("repomind.biz_context")).(*service.BizContext); ok && bc != nil {
		return bc
	}
	return &service.BizContext{}
}

// --- Request metadata getters ---

func (s *sBizCtx) GetRequestID(ctx context.Context) string { return s.Get(ctx).RequestID }
func (s *sBizCtx) GetTraceID(ctx context.Context) string   { return s.Get(ctx).TraceID }
func (s *sBizCtx) GetClientIP(ctx context.Context) string   { return s.Get(ctx).ClientIP }
func (s *sBizCtx) GetUserAgent(ctx context.Context) string  { return s.Get(ctx).UserAgent }

// --- Identity convenience methods ---

func (s *sBizCtx) GetUserID(ctx context.Context) string {
	if bc := s.Get(ctx); bc.Identity != nil {
		return bc.Identity.UserID
	}
	return ""
}

func (s *sBizCtx) GetAuthType(ctx context.Context) string {
	if bc := s.Get(ctx); bc.Identity != nil {
		return bc.Identity.Type
	}
	return ""
}

// --- Tenant convenience methods ---

func (s *sBizCtx) GetTenantID(ctx context.Context) string {
	if bc := s.Get(ctx); bc.Tenant != nil {
		return bc.Tenant.TenantID
	}
	return ""
}

func (s *sBizCtx) GetRole(ctx context.Context) string {
	if bc := s.Get(ctx); bc.Tenant != nil {
		return bc.Tenant.Role
	}
	return ""
}

func (s *sBizCtx) GetPermissions(ctx context.Context) []service.Permission {
	if bc := s.Get(ctx); bc.Tenant != nil {
		return bc.Tenant.Permissions
	}
	return nil
}

// --- Platform admin convenience methods ---

func (s *sBizCtx) GetPlatformRole(ctx context.Context) string {
	if bc := s.Get(ctx); bc.PlatformAdmin != nil {
		return bc.PlatformAdmin.Role
	}
	return ""
}

func (s *sBizCtx) IsPlatformAdmin(ctx context.Context) bool {
	return s.Get(ctx).PlatformAdmin != nil
}

// --- Setters (called from middlewares) ---

func (s *sBizCtx) SetRequestID(ctx context.Context, id string) context.Context {
	s.Get(ctx).RequestID = id
	return ctx
}

func (s *sBizCtx) SetTraceID(ctx context.Context, id string) context.Context {
	s.Get(ctx).TraceID = id
	return ctx
}

func (s *sBizCtx) SetClientIP(ctx context.Context, ip string) context.Context {
	s.Get(ctx).ClientIP = ip
	return ctx
}

func (s *sBizCtx) SetUserAgent(ctx context.Context, ua string) context.Context {
	s.Get(ctx).UserAgent = ua
	return ctx
}

func (s *sBizCtx) SetIdentity(ctx context.Context, identity *service.AuthIdentity) context.Context {
	s.Get(ctx).Identity = identity
	return ctx
}

func (s *sBizCtx) SetTenant(ctx context.Context, tc *service.TenantContext) context.Context {
	s.Get(ctx).Tenant = tc
	return ctx
}

func (s *sBizCtx) SetPlatformAdmin(ctx context.Context, pac *service.PlatformAdminContext) context.Context {
	s.Get(ctx).PlatformAdmin = pac
	return ctx
}
