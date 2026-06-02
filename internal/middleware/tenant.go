package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

type tenantSelectorCandidate struct {
	Source string
	Value  string
}

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
	selectors := tenantSelectors(r)
	if len(selectors) == 0 && identity.Type != "api_key" && identity.TenantID != "" {
		selectors = append(selectors, tenantSelectorCandidate{Source: "identity", Value: identity.TenantID})
	}
	if len(selectors) == 0 {
		writeError(r, http.StatusBadRequest, "TENANT_REQUIRED", gerror.New("tenant selector is required"))
		return
	}
	primary := selectors[0]
	tc, err := service.TenantMembershipService().ResolveTenantContext(r.GetCtx(), identity.UserID, primary.Value)
	if err != nil {
		writeError(r, http.StatusForbidden, "TENANT_FORBIDDEN", err)
		return
	}
	for _, selector := range selectors[1:] {
		resolved, err := service.TenantMembershipService().ResolveTenantContext(r.GetCtx(), identity.UserID, selector.Value)
		if err != nil {
			writeError(r, http.StatusForbidden, "TENANT_FORBIDDEN", err)
			return
		}
		if resolved.TenantID != tc.TenantID {
			writeError(r, http.StatusBadRequest, "TENANT_MISMATCH", gerror.Newf("tenant selector %s does not match %s", selector.Source, primary.Source))
			return
		}
	}
	if identity.TenantID != "" && identity.Type != "api_key" && tc.TenantID != identity.TenantID {
		writeError(r, http.StatusForbidden, "TENANT_FORBIDDEN", gerror.New("identity cannot access another tenant"))
		return
	}
	tc.AuthType = identity.Type
	tc.Scopes = append([]string(nil), identity.Scopes...)
	tc.APIKeyID = identity.APIKeyID
	if identity.Type == "api_key" {
		grant, err := service.APIKeyService().ResolveTenantGrant(r.GetCtx(), identity.APIKeyID, tc.TenantID)
		if err != nil {
			writeError(r, http.StatusForbidden, "TENANT_GRANT_FORBIDDEN", err)
			return
		}
		tc.APIKeyGrantID = grant.ID
		tc.Scopes = intersectScopes(identity.Scopes, grant.Scopes)
	}
	tc.RequestID = service.RequestIDFromCtx(r.GetCtx())
	tc.Permissions = service.RBAC().PermissionsForContext(r.GetCtx(), tc)
	ctx := service.WithTenantContext(r.GetCtx(), tc)

	// Also populate BizContext.
	service.BizCtx().SetTenant(ctx, tc)

	r.SetCtx(ctx)
	r.Middleware.Next()
}

func tenantOptionalRoute(r *ghttp.Request) bool {
	return tenantOptionalPath(r.URL.Path, r.Method)
}

// tenantOptionalRoutes is the centralized registry of routes that do not
// require a tenant selector. Add new entries here instead of scattering
// hardcoded path checks.
var tenantOptionalRoutes = []struct {
	Method string
	Path   string // exact match after trimming trailing slash
	Prefix string // prefix match (mutually exclusive with Path)
	Suffix string // suffix match (used with Prefix)
}{
	{Method: http.MethodGet, Path: "/api/v1/me"},
	{Method: http.MethodGet, Path: "/api/v1/me/tenants"},
	{Method: http.MethodGet, Path: "/api/v1/me/access"},
	{Method: http.MethodGet, Path: "/api/v1/me/invitations"},
	{Method: http.MethodPost, Prefix: "/api/v1/me/invitations/", Suffix: "/decline"},
	{Method: http.MethodGet, Path: "/api/v1/me/security-events"},
	{Method: http.MethodGet, Path: "/api/v1/me/sessions"},
	{Method: http.MethodDelete, Path: "/api/v1/me/sessions"},
	{Method: http.MethodDelete, Prefix: "/api/v1/me/sessions/"},
	{Method: http.MethodGet, Path: "/api/v1/me/totp/status"},
	{Method: http.MethodPost, Path: "/api/v1/me/totp/setup"},
	{Method: http.MethodPost, Path: "/api/v1/me/totp/enable"},
	{Method: http.MethodPost, Path: "/api/v1/me/totp/disable"},
	{Method: http.MethodPost, Path: "/api/v1/me/totp/backup-codes/regenerate"},
	{Method: http.MethodPost, Path: "/api/v1/invitations/accept"},
	{Method: http.MethodPost, Path: "/api/v1/tenants"},
}

func tenantOptionalPath(rawPath, method string) bool {
	path := strings.TrimRight(rawPath, "/")
	if path == "" {
		path = "/"
	}
	for _, route := range tenantOptionalRoutes {
		if method != route.Method {
			continue
		}
		if route.Path != "" && path == route.Path {
			return true
		}
		if route.Prefix != "" && strings.HasPrefix(path, route.Prefix) {
			if route.Suffix == "" || strings.HasSuffix(path, route.Suffix) {
				return true
			}
		}
	}
	return false
}

func tenantSelector(r *ghttp.Request) string {
	selectors := tenantSelectors(r)
	if len(selectors) == 0 {
		return ""
	}
	return selectors[0].Value
}

func tenantSelectors(r *ghttp.Request) []tenantSelectorCandidate {
	selectors := make([]tenantSelectorCandidate, 0, 4)
	if value := tenantFromPath(r.URL.Path); value != "" {
		selectors = append(selectors, tenantSelectorCandidate{Source: "path", Value: value})
	}
	for _, header := range []string{"X-Tenant-ID", "X-Tenant-Slug"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			selectors = append(selectors, tenantSelectorCandidate{Source: header, Value: value})
		}
	}
	if value := strings.TrimSpace(r.GetQuery("tenant").String()); value != "" {
		selectors = append(selectors, tenantSelectorCandidate{Source: "query", Value: value})
	}
	return selectors
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

func intersectScopes(keyScopes, grantScopes []string) []string {
	if len(grantScopes) == 0 {
		return append([]string(nil), keyScopes...)
	}
	keyHasStar := containsScope(keyScopes, "*")
	grantHasStar := containsScope(grantScopes, "*")
	switch {
	case keyHasStar && grantHasStar:
		return []string{"*"}
	case keyHasStar:
		return uniqueScopes(grantScopes)
	case grantHasStar:
		return uniqueScopes(keyScopes)
	}
	allowed := map[string]struct{}{}
	for _, scope := range grantScopes {
		allowed[scope] = struct{}{}
	}
	out := make([]string, 0, len(keyScopes))
	seen := map[string]struct{}{}
	for _, scope := range keyScopes {
		if _, ok := allowed[scope]; !ok {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return out
}

func containsScope(scopes []string, target string) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func uniqueScopes(scopes []string) []string {
	out := make([]string, 0, len(scopes))
	seen := map[string]struct{}{}
	for _, scope := range scopes {
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return out
}
