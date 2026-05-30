package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

const RequestIDHeader = "X-Request-ID"

func RequestContext(r *ghttp.Request) {
	requestID := r.Header.Get(RequestIDHeader)
	if requestID == "" {
		requestID = uuid.GenerateV4()
	}

	// Initialise BizContext and populate request-scoped metadata.
	ctx := service.WithBizContext(r.GetCtx())
	service.BizCtx().SetRequestID(ctx, requestID)
	service.BizCtx().SetClientIP(ctx, r.GetClientIp())
	service.BizCtx().SetUserAgent(ctx, r.Header.Get("User-Agent"))

	// Preserve legacy independent context key (backward compatible).
	ctx = service.WithRequestID(ctx, requestID)

	r.Response.Header().Set(RequestIDHeader, requestID)
	r.SetCtx(ctx)
	r.Middleware.Next()
}
