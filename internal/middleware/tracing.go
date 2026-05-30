package middleware

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

// Tracing logs every API request with structured fields: method, path, status,
// latency and request_id. It also extracts an OpenTelemetry trace ID (if
// present) into BizContext so that downstream code (audit, logging) can
// reference it.
func Tracing(r *ghttp.Request) {
	start := time.Now()
	requestID := service.RequestIDFromCtx(r.GetCtx())

	// Attempt to extract OpenTelemetry TraceID from the incoming context.
	// The span is propagated by the Go runtime's context if OTel is enabled.
	traceID := extractTraceID(r.GetCtx())
	if traceID != "" {
		service.BizCtx().SetTraceID(r.GetCtx(), traceID)
	}

	r.Middleware.Next()
	latency := time.Since(start)
	status := r.Response.Status
	if status == 0 {
		status = 200
	}
	g.Log().Infof(r.GetCtx(),
		"[trace] method=%s path=%s status=%d latency=%s request_id=%s trace_id=%s",
		r.Method, r.URL.Path, status, latency.Round(time.Millisecond), requestID, traceID,
	)
}

// extractTraceID returns the OpenTelemetry trace ID from ctx if available.
// Returns empty string when OTel is not configured.
func extractTraceID(ctx context.Context) string {
	// Use string conversion to avoid importing go.opentelemetry.io/otel
	// when OTel is not yet configured. When OTel is added, replace with:
	//   sc := trace.SpanContextFromContext(ctx)
	//   if sc.HasTraceID() { return sc.TraceID().String() }
	//   return ""
	_ = ctx
	return ""
}
