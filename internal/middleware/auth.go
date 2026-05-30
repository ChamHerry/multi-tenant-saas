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
	ctx := service.WithAuthIdentity(r.GetCtx(), identity)

	// Also populate BizContext.
	service.BizCtx().SetIdentity(ctx, identity)

	r.SetCtx(ctx)
	r.Middleware.Next()
}
