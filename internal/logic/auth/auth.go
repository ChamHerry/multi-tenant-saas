package auth

import (
	"context"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"repomind-temp/internal/service"
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

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
	if !devHeaderEnabled(ctx) {
		return nil, gerror.New("authentication required")
	}
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userID == "" {
		return nil, gerror.New("authentication required")
	}
	if !internalIDPattern.MatchString(userID) {
		return nil, gerror.New("invalid X-User-ID")
	}
	user, err := service.UserService().GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Status != "active" {
		return nil, gerror.Newf("user %s is not active", userID)
	}
	return &service.AuthIdentity{UserID: userID, Type: "dev_header"}, nil
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

func devHeaderEnabled(ctx context.Context) bool {
	if !g.Cfg().MustGet(ctx, "auth.devHeader.enabled", false).Bool() {
		return false
	}
	env := strings.ToLower(strings.TrimSpace(g.Cfg().MustGet(ctx, "server.env", "local").String()))
	return env == "local" || env == "test"
}
