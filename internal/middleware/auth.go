package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

func Auth(r *ghttp.Request) {
	identity, err := service.Auth().Authenticate(r.GetCtx(), r)
	if err != nil {
		writeError(r, http.StatusUnauthorized, "UNAUTHENTICATED", err)
		return
	}
	r.SetCtx(service.WithAuthIdentity(r.GetCtx(), identity))
	r.Middleware.Next()
}
