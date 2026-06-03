package auth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

type sAuth struct{}

func init() {
	service.RegisterAuth(&sAuth{})
}

func (s *sAuth) Authenticate(ctx context.Context, r *ghttp.Request) (*service.AuthIdentity, error) {
	if token := bearerToken(r.Header.Get("Authorization")); token != "" {
		return service.APIKeyService().Authenticate(ctx, token)
	}
	if cookie := r.Cookie.Get(service.AuthSessionService().SessionCookieName(ctx)); cookie != nil && strings.TrimSpace(cookie.String()) != "" {
		return service.AuthSessionService().Authenticate(ctx, cookie.String())
	}
	return nil, gerror.New("authentication required")
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
