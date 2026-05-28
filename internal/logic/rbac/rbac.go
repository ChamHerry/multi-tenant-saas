package rbac

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"repomind-temp/internal/service"
)

type sRBAC struct{}

func init() {
	service.RegisterRBAC(&sRBAC{})
}

var permissionOrder = []service.Permission{
	service.PermissionTenantRead,
	service.PermissionTenantManage,
	service.PermissionMemberRead,
	service.PermissionMemberManage,
	service.PermissionInvitationManage,
	service.PermissionAuditRead,
}

var userScopePermissions = map[service.Permission]struct{}{
	service.PermissionUserRead:         {},
	service.PermissionUserTenantRead:   {},
	service.PermissionUserSecurityRead: {},
	service.PermissionAPIKeySelfManage: {},
}

var rolePermissions = map[string]map[service.Permission]struct{}{
	"owner": {
		service.PermissionTenantRead:       {},
		service.PermissionTenantManage:     {},
		service.PermissionMemberRead:       {},
		service.PermissionMemberManage:     {},
		service.PermissionInvitationManage: {},
		service.PermissionAuditRead:        {},
	},
	"admin": {
		service.PermissionTenantRead:       {},
		service.PermissionMemberRead:       {},
		service.PermissionMemberManage:     {},
		service.PermissionInvitationManage: {},
		service.PermissionAuditRead:        {},
	},
	"member": {
		service.PermissionTenantRead: {},
		service.PermissionMemberRead: {},
	},
	"viewer": {
		service.PermissionTenantRead: {},
	},
}

func (s *sRBAC) Can(ctx context.Context, tc *service.TenantContext, permission service.Permission) bool {
	if tc == nil {
		return false
	}
	allowed, ok := rolePermissions[tc.Role]
	if !ok {
		return false
	}
	if _, ok = allowed[permission]; !ok {
		return false
	}
	if tc.AuthType != "api_key" {
		return true
	}
	return scopeAllows(tc.Scopes, permission)
}

func (s *sRBAC) Require(ctx context.Context, permission service.Permission) error {
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return err
	}
	if !s.Can(ctx, tc, permission) {
		return gerror.NewCodef(gcode.CodeNotAuthorized, "role %s is not allowed to perform %s", tc.Role, permission)
	}
	return nil
}

func (s *sRBAC) RequireAuthScope(ctx context.Context, permission service.Permission) error {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return err
	}
	if identity.Type != "api_key" {
		return nil
	}
	if !scopeAllows(identity.Scopes, permission) {
		return gerror.NewCodef(gcode.CodeNotAuthorized, "api key scope is not allowed to perform %s", permission)
	}
	return nil
}

func (s *sRBAC) PermissionsForRole(role string) []service.Permission {
	allowed, ok := rolePermissions[role]
	if !ok {
		return []service.Permission{}
	}
	permissions := make([]service.Permission, 0, len(allowed))
	for _, permission := range permissionOrder {
		if _, ok := allowed[permission]; ok {
			permissions = append(permissions, permission)
		}
	}
	return permissions
}

func (s *sRBAC) PermissionsForContext(ctx context.Context, tc *service.TenantContext) []service.Permission {
	if tc == nil {
		return []service.Permission{}
	}
	permissions := make([]service.Permission, 0, len(permissionOrder))
	for _, permission := range permissionOrder {
		if s.Can(ctx, tc, permission) {
			permissions = append(permissions, permission)
		}
	}
	return permissions
}

func scopeAllows(scopes []string, permission service.Permission) bool {
	if len(scopes) == 0 {
		return false
	}
	for _, scope := range scopes {
		if scope == "*" || scope == string(permission) {
			return true
		}
	}
	return false
}

func KnownPermission(scope string) bool {
	if scope == "*" {
		return true
	}
	if _, ok := userScopePermissions[service.Permission(scope)]; ok {
		return true
	}
	for _, permissions := range rolePermissions {
		if _, ok := permissions[service.Permission(scope)]; ok {
			return true
		}
	}
	return false
}
