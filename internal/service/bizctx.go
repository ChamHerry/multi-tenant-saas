package service

import "context"

// BizContext aggregates all request-scoped context data into a single struct.
// It is created once in the RequestContext middleware and progressively filled
// by Auth, TenantResolver, and PlatformAdmin middlewares.
type BizContext struct {
	// Request metadata — filled by RequestContext middleware.
	RequestID string
	TraceID   string // reserved for OpenTelemetry
	ClientIP  string
	UserAgent string

	// Identity — filled by Auth middleware.
	Identity *AuthIdentity

	// Tenant — filled by TenantResolver middleware. Nil for personal routes.
	Tenant *TenantContext

	// PlatformAdmin — filled by PlatformAdmin middleware. Nil for non-admin users.
	PlatformAdmin *PlatformAdminContext
}

// BizContextKey is the context key used to store BizContext.
type BizContextKey string

const bizCtxKey BizContextKey = "repomind.biz_context"

// WithBizContext injects a new BizContext into ctx.
// Called once at the start of the middleware chain (RequestContext).
func WithBizContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, bizCtxKey, &BizContext{})
}

// BizContextFromCtx retrieves the BizContext from ctx (internal helper).
func BizContextFromCtx(ctx context.Context) (*BizContext, bool) {
	bc, ok := ctx.Value(bizCtxKey).(*BizContext)
	return bc, ok && bc != nil
}

// IBizCtx provides a unified interface for reading and writing request-scoped
// business context data. All Get methods return zero values when the context
// is not initialised.
type IBizCtx interface {
	// Get returns the full BizContext (never nil — returns empty struct if missing).
	Get(ctx context.Context) *BizContext

	// Request metadata.
	GetRequestID(ctx context.Context) string
	GetTraceID(ctx context.Context) string
	GetClientIP(ctx context.Context) string
	GetUserAgent(ctx context.Context) string

	// Identity convenience methods.
	GetUserID(ctx context.Context) string
	GetAuthType(ctx context.Context) string // "session" | "api_key"

	// Tenant convenience methods.
	GetTenantID(ctx context.Context) string
	GetRole(ctx context.Context) string
	GetPermissions(ctx context.Context) []Permission

	// Platform admin convenience methods.
	GetPlatformRole(ctx context.Context) string
	IsPlatformAdmin(ctx context.Context) bool

	// Setters — called from middlewares.
	SetRequestID(ctx context.Context, id string) context.Context
	SetTraceID(ctx context.Context, id string) context.Context
	SetClientIP(ctx context.Context, ip string) context.Context
	SetUserAgent(ctx context.Context, ua string) context.Context
	SetIdentity(ctx context.Context, identity *AuthIdentity) context.Context
	SetTenant(ctx context.Context, tc *TenantContext) context.Context
	SetPlatformAdmin(ctx context.Context, pac *PlatformAdminContext) context.Context
}

var localBizCtx IBizCtx

// BizCtx returns the IBizCtx implementation.
func BizCtx() IBizCtx {
	if localBizCtx == nil {
		panic("implement not found for interface IBizCtx")
	}
	return localBizCtx
}

// RegisterBizCtx registers the IBizCtx implementation.
func RegisterBizCtx(i IBizCtx) {
	localBizCtx = i
}
