package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"repomind-temp/internal/service"
	"repomind-temp/utility/uuid"
)

const RequestIDHeader = "X-Request-ID"

func RequestContext(r *ghttp.Request) {
	requestID := r.Header.Get(RequestIDHeader)
	if requestID == "" {
		requestID = uuid.GenerateV4()
	}
	r.Response.Header().Set(RequestIDHeader, requestID)
	r.SetCtx(service.WithRequestID(r.GetCtx(), requestID))
	r.Middleware.Next()
}
