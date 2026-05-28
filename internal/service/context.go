package service

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

type contextKey string

const (
	authIdentityContextKey  contextKey = "repomind.auth_identity"
	tenantContextContextKey contextKey = "repomind.tenant_context"
	platformAdminContextKey contextKey = "repomind.platform_admin_context"
	requestIDContextKey     contextKey = "repomind.request_id"
)

type AuthIdentity struct {
	UserID    string
	TenantID  string
	Type      string
	Scopes    []string
	APIKeyID  string
	SessionID string
}

func WithAuthIdentity(ctx context.Context, identity *AuthIdentity) context.Context {
	return context.WithValue(ctx, authIdentityContextKey, identity)
}

func AuthIdentityFromCtx(ctx context.Context) (*AuthIdentity, bool) {
	identity, ok := ctx.Value(authIdentityContextKey).(*AuthIdentity)
	return identity, ok && identity != nil
}

func MustAuthIdentity(ctx context.Context) (*AuthIdentity, error) {
	identity, ok := AuthIdentityFromCtx(ctx)
	if !ok {
		return nil, gerror.New("authenticated identity not found in context")
	}
	return identity, nil
}

func WithTenantContext(ctx context.Context, tc *TenantContext) context.Context {
	return context.WithValue(ctx, tenantContextContextKey, tc)
}

func TenantContextFromCtx(ctx context.Context) (*TenantContext, bool) {
	tc, ok := ctx.Value(tenantContextContextKey).(*TenantContext)
	return tc, ok && tc != nil
}

func MustTenantContext(ctx context.Context) (*TenantContext, error) {
	tc, ok := TenantContextFromCtx(ctx)
	if !ok {
		return nil, gerror.New("tenant context not found")
	}
	return tc, nil
}

func WithPlatformAdminContext(ctx context.Context, pac *PlatformAdminContext) context.Context {
	return context.WithValue(ctx, platformAdminContextKey, pac)
}

func PlatformAdminContextFromCtx(ctx context.Context) (*PlatformAdminContext, bool) {
	pac, ok := ctx.Value(platformAdminContextKey).(*PlatformAdminContext)
	return pac, ok && pac != nil
}

func MustPlatformAdminContext(ctx context.Context) (*PlatformAdminContext, error) {
	pac, ok := PlatformAdminContextFromCtx(ctx)
	if !ok {
		return nil, gerror.New("platform admin context not found")
	}
	return pac, nil
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestIDFromCtx(ctx context.Context) string {
	if value, ok := ctx.Value(requestIDContextKey).(string); ok {
		return value
	}
	return ""
}
