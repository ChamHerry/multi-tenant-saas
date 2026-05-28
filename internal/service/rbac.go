package service

import "context"

type Permission string

const (
	PermissionTenantRead       Permission = "tenant:read"
	PermissionTenantManage     Permission = "tenant:manage"
	PermissionMemberRead       Permission = "member:read"
	PermissionMemberManage     Permission = "member:manage"
	PermissionInvitationManage Permission = "tenant:invitation:manage"
	PermissionAuditRead        Permission = "tenant:audit:read"
	PermissionUserRead         Permission = "user:read"
	PermissionUserTenantRead   Permission = "user:tenant:read"
	PermissionUserSecurityRead Permission = "user:security:read"
	PermissionAPIKeySelfManage Permission = "api_key:self_manage"
)

type IRBAC interface {
	Can(ctx context.Context, tc *TenantContext, permission Permission) bool
	Require(ctx context.Context, permission Permission) error
	RequireAuthScope(ctx context.Context, permission Permission) error
	PermissionsForRole(role string) []Permission
	PermissionsForContext(ctx context.Context, tc *TenantContext) []Permission
}

var localRBAC IRBAC

func RBAC() IRBAC {
	if localRBAC == nil {
		panic("implement not found for interface IRBAC")
	}
	return localRBAC
}

func RegisterRBAC(i IRBAC) {
	localRBAC = i
}
