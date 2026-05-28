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
	PermissionBillingRead      Permission = "tenant:billing:read"
	PermissionBillingManage    Permission = "tenant:billing:manage"
	PermissionAPIKeyManage     Permission = "api_key:manage"
)

type IRBAC interface {
	Can(ctx context.Context, tc *TenantContext, permission Permission) bool
	Require(ctx context.Context, permission Permission) error
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
