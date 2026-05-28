package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"repomind-temp/internal/service"
)

func PlatformAdmin(r *ghttp.Request) {
	identity, err := service.MustAuthIdentity(r.GetCtx())
	if err != nil {
		writeError(r, http.StatusUnauthorized, "UNAUTHENTICATED", err)
		return
	}
	if identity.Type == "api_key" {
		writeError(r, http.StatusForbidden, "PLATFORM_ADMIN_SESSION_REQUIRED", nil)
		return
	}
	pac, err := service.PlatformAdminService().Resolve(r.GetCtx(), identity.UserID)
	if err != nil {
		writeError(r, http.StatusForbidden, "PLATFORM_ADMIN_FORBIDDEN", err)
		return
	}
	r.SetCtx(service.WithPlatformAdminContext(r.GetCtx(), pac))
	r.Middleware.Next()
}
