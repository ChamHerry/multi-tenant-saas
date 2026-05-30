package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

const csrfHeader = "X-CSRF-Token"

func CSRF(r *ghttp.Request) {
	if csrfSafeMethod(r.Method) {
		r.Middleware.Next()
		return
	}
	identity, err := service.MustAuthIdentity(r.GetCtx())
	if err != nil {
		writeError(r, http.StatusUnauthorized, "UNAUTHENTICATED", err)
		return
	}
	if identity.Type != "session" {
		r.Middleware.Next()
		return
	}
	headerToken := strings.TrimSpace(r.Header.Get(csrfHeader))
	cookie := r.Cookie.Get(service.AuthSessionService().CSRFCookieName(r.GetCtx()))
	cookieToken := ""
	if cookie != nil {
		cookieToken = strings.TrimSpace(cookie.String())
	}
	if headerToken == "" || cookieToken == "" || headerToken != cookieToken {
		writeError(r, http.StatusForbidden, "CSRF_INVALID", gerror.New("csrf token is required or invalid"))
		return
	}
	if err = service.AuthSessionService().ValidateCSRF(r.GetCtx(), identity.SessionID, headerToken); err != nil {
		writeError(r, http.StatusForbidden, "CSRF_INVALID", err)
		return
	}
	r.Middleware.Next()
}

func csrfSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
