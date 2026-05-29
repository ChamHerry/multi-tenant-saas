package apikey

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	apiapikey "repomind-temp/api/apikey"
	"repomind-temp/api/apikey/v1"
	"repomind-temp/internal/service"
)

type ControllerV1 struct{}

func NewV1() apiapikey.IAPIKeyV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) CreatePersonal(ctx context.Context, req *v1.CreatePersonalReq) (res *v1.CreateRes, err error) {
	identity, tc, err := requireTenantScopedInteractiveIdentity(ctx)
	if err != nil {
		return nil, err
	}
	expiresAt, err := parseExpiresAt(req.ExpiresAt)
	if err != nil {
		return nil, err
	}
	grantScopes, err := tenantGrantScopes(req.Scopes, tc.Permissions)
	if err != nil {
		return nil, err
	}
	created, err := service.APIKeyService().CreatePersonal(ctx, service.CreatePersonalAPIKeyInput{
		UserID: identity.UserID,
		Name:   req.Name,
		Scopes: req.Scopes,
		Grants: []service.CreateAPIKeyTenantGrantInput{{
			TenantID: tc.TenantID,
			Scopes:   grantScopes,
		}},
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		Action:       "api_key.create_tenant_scoped",
		ResourceType: "api_key",
		ResourceID:   created.ID,
		Metadata: map[string]any{
			"tenant_id":    tc.TenantID,
			"scopes":       req.Scopes,
			"grant_scopes": grantScopes,
		},
	})
	return &v1.CreateRes{APIKey: &created.APIKey, RawKey: created.RawKey}, nil
}

func (c *ControllerV1) ListPersonal(ctx context.Context, req *v1.ListPersonalReq) (res *v1.ListRes, err error) {
	identity, tc, err := requireTenantScopedInteractiveIdentity(ctx)
	if err != nil {
		return nil, err
	}
	items, err := service.APIKeyService().ListPersonal(ctx, identity.UserID, tc.TenantID)
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{APIKeys: items}, nil
}

func (c *ControllerV1) RevokePersonal(ctx context.Context, req *v1.RevokePersonalReq) (res *v1.ActionRes, err error) {
	identity, tc, err := requireTenantScopedInteractiveIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.APIKeyService().RevokePersonal(ctx, identity.UserID, tc.TenantID, req.APIKey); err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		Action:       "api_key.revoke_tenant_scoped",
		ResourceType: "api_key",
		ResourceID:   req.APIKey,
		Metadata:     map[string]any{"tenant_id": tc.TenantID},
	})
	return &v1.ActionRes{OK: true}, nil
}

func parseExpiresAt(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func requireInteractiveIdentity(ctx context.Context) (*service.AuthIdentity, error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if identity.Type == "api_key" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api keys cannot manage tenant api keys")
	}
	return identity, nil
}

func requireTenantScopedInteractiveIdentity(ctx context.Context) (*service.AuthIdentity, *service.TenantContext, error) {
	identity, err := requireInteractiveIdentity(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err = service.RBAC().RequireAuthScope(ctx, service.PermissionAPIKeySelfManage); err != nil {
		return nil, nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	if tc.UserID != identity.UserID {
		return nil, nil, gerror.NewCode(gcode.CodeNotAuthorized, "tenant context does not belong to current user")
	}
	return identity, tc, nil
}

func tenantGrantScopes(requestedScopes []string, allowedPermissions []service.Permission) ([]string, error) {
	allowed := make(map[string]struct{}, len(allowedPermissions))
	for _, permission := range allowedPermissions {
		allowed[string(permission)] = struct{}{}
	}
	grantScopes := make([]string, 0, len(requestedScopes))
	seen := map[string]struct{}{}
	for _, scope := range requestedScopes {
		if scope == "*" {
			return nil, gerror.NewCode(gcode.CodeNotAuthorized, "wildcard api key scope is not allowed for tenant-scoped api keys")
		}
		if !isTenantPermissionScope(scope) {
			continue
		}
		if _, ok := allowed[scope]; !ok {
			return nil, gerror.NewCodef(gcode.CodeNotAuthorized, "api key scope %q exceeds current tenant permissions", scope)
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		grantScopes = append(grantScopes, scope)
	}
	return grantScopes, nil
}

func isTenantPermissionScope(scope string) bool {
	switch service.Permission(scope) {
	case service.PermissionTenantRead,
		service.PermissionTenantManage,
		service.PermissionMemberRead,
		service.PermissionMemberManage,
		service.PermissionInvitationManage,
		service.PermissionAuditRead:
		return true
	default:
		return false
	}
}
