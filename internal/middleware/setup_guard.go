package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

func InitGuard(r *ghttp.Request) {
	if setupAllowedRequest(r) {
		r.Middleware.Next()
		return
	}
	state, err := service.SystemSetup().State(r.GetCtx())
	if err != nil {
		writeError(r, http.StatusServiceUnavailable, "SYSTEM_SETUP_STATUS_UNAVAILABLE", err)
		return
	}
	if state.RequiresSetup {
		writeError(r, http.StatusServiceUnavailable, service.SetupCodeRequired, gerror.New("system setup is required before using this API"))
		return
	}
	r.Middleware.Next()
}

func setupAllowedRequest(r *ghttp.Request) bool {
	if strings.EqualFold(r.Method, http.MethodOptions) {
		return true
	}
	path := r.URL.Path
	return path == "/api/v1/setup/state" || path == "/api/v1/setup/complete"
}
