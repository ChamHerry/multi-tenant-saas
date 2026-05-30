package middleware

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

// Tracing logs every API request with structured fields: method, path, status,
// latency and request_id. It is placed after RequestContext so the X-Request-ID
// is already available.
func Tracing(r *ghttp.Request) {
	start := time.Now()
	requestID := service.RequestIDFromCtx(r.GetCtx())
	r.Middleware.Next()
	latency := time.Since(start)
	status := r.Response.Status
	if status == 0 {
		status = 200
	}
	g.Log().Infof(r.GetCtx(),
		"[trace] method=%s path=%s status=%d latency=%s request_id=%s",
		r.Method, r.URL.Path, status, latency.Round(time.Millisecond), requestID,
	)
}
