package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"repomind-temp/internal/service"
)

func Require(permission service.Permission) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if err := service.RBAC().Require(r.GetCtx(), permission); err != nil {
			writeError(r, http.StatusForbidden, "ROLE_FORBIDDEN", err)
			return
		}
		r.Middleware.Next()
	}
}
