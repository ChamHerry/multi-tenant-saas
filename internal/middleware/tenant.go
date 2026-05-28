package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"repomind-temp/internal/service"
)

func TenantResolver(r *ghttp.Request) {
	if tenantOptionalRoute(r) {
		r.Middleware.Next()
		return
	}
	identity, err := service.MustAuthIdentity(r.GetCtx())
	if err != nil {
		writeError(r, http.StatusUnauthorized, "UNAUTHENTICATED", err)
		return
	}
	selector := tenantSelector(r)
	if selector == "" && identity.TenantID != "" {
		selector = identity.TenantID
	}
	if selector == "" {
		writeError(r, http.StatusBadRequest, "TENANT_REQUIRED", gerror.New("tenant selector is required"))
		return
	}
	tc, err := service.TenantMembershipService().ResolveTenantContext(r.GetCtx(), identity.UserID, selector)
	if err != nil {
		writeError(r, http.StatusForbidden, "TENANT_FORBIDDEN", err)
		return
	}
	if identity.TenantID != "" && tc.TenantID != identity.TenantID {
		writeError(r, http.StatusForbidden, "TENANT_FORBIDDEN", gerror.New("api key cannot access another tenant"))
		return
	}
	tc.AuthType = identity.Type
	tc.Scopes = append([]string(nil), identity.Scopes...)
	tc.APIKeyID = identity.APIKeyID
	tc.RequestID = service.RequestIDFromCtx(r.GetCtx())
	tc.Permissions = service.RBAC().PermissionsForContext(r.GetCtx(), tc)
	r.SetCtx(service.WithTenantContext(r.GetCtx(), tc))
	r.Middleware.Next()
}

func tenantOptionalRoute(r *ghttp.Request) bool {
	return tenantOptionalPath(r.URL.Path, r.Method)
}

func tenantOptionalPath(rawPath, method string) bool {
	path := strings.TrimRight(rawPath, "/")
	if path == "" {
		path = "/"
	}
	switch {
	case path == "/api/v1/me" && method == http.MethodGet:
		return true
	case path == "/api/v1/me/tenants" && method == http.MethodGet:
		return true
	case path == "/api/v1/me/access" && method == http.MethodGet:
		return true
	case path == "/api/v1/me/invitations" && method == http.MethodGet:
		return true
	case strings.HasPrefix(path, "/api/v1/me/invitations/") && strings.HasSuffix(path, "/decline") && method == http.MethodPost:
		return true
	case path == "/api/v1/me/security-events" && method == http.MethodGet:
		return true
	case path == "/api/v1/invitations/accept" && method == http.MethodPost:
		return true
	case path == "/api/v1/tenants" && method == http.MethodPost:
		return true
	default:
		return false
	}
}

func tenantSelector(r *ghttp.Request) string {
	for _, header := range []string{"X-Tenant-ID", "X-Tenant-Slug"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	if value := strings.TrimSpace(r.GetQuery("tenant").String()); value != "" {
		return value
	}
	return tenantFromPath(r.URL.Path)
}

func tenantFromPath(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "tenants" {
			return parts[i+1]
		}
	}
	return ""
}
