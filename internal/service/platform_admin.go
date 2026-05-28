package service

import (
	"context"
	"time"
)

type PlatformPermission string

const (
	PlatformPermissionTenantRead   PlatformPermission = "platform:tenant:read"
	PlatformPermissionTenantManage PlatformPermission = "platform:tenant:manage"
	PlatformPermissionUserRead     PlatformPermission = "platform:user:read"
	PlatformPermissionUserManage   PlatformPermission = "platform:user:manage"
	PlatformPermissionAuditRead    PlatformPermission = "platform:audit:read"
	PlatformPermissionAdminManage  PlatformPermission = "platform:admin:manage"
)

type PlatformAdminContext struct {
	UserID      string               `json:"user_id"`
	Role        string               `json:"role"`
	Status      string               `json:"status"`
	Permissions []PlatformPermission `json:"permissions"`
}

type PlatformAdmin struct {
	UserID          string    `json:"user_id"`
	Role            string    `json:"role"`
	Status          string    `json:"status"`
	CreatedByUserID string    `json:"created_by_user_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PlatformTenantFilter struct {
	Query  string
	Status string
	Limit  int
	Offset int
}

type PlatformUserFilter struct {
	Query  string
	Status string
	Limit  int
	Offset int
}

type PlatformTenantList struct {
	Items []Tenant `json:"items"`
	Total int      `json:"total"`
}

type PlatformUserList struct {
	Items []User `json:"items"`
	Total int    `json:"total"`
}

type PlatformAdminFilter struct {
	Query  string
	Role   string
	Status string
	Limit  int
	Offset int
}

type PlatformAdminList struct {
	Items []PlatformAdmin `json:"items"`
	Total int             `json:"total"`
}

type IPlatformAdmin interface {
	Resolve(ctx context.Context, userID string) (*PlatformAdminContext, error)
	ResolveOptional(ctx context.Context, userID string) (*PlatformAdminContext, error)
	PermissionsForRole(role string) []PlatformPermission
	Require(ctx context.Context, permission PlatformPermission) error
	ListTenants(ctx context.Context, filter PlatformTenantFilter) (*PlatformTenantList, error)
	ListUsers(ctx context.Context, filter PlatformUserFilter) (*PlatformUserList, error)
	UpdateUserStatus(ctx context.Context, actorUserID, targetUserID, status string) error
	ListPlatformAdmins(ctx context.Context, filter PlatformAdminFilter) (*PlatformAdminList, error)
	Grant(ctx context.Context, actorUserID, targetUserID, role string) error
	Revoke(ctx context.Context, actorUserID, targetUserID string) error
}

var localPlatformAdmin IPlatformAdmin

func PlatformAdminService() IPlatformAdmin {
	if localPlatformAdmin == nil {
		panic("implement not found for interface IPlatformAdmin")
	}
	return localPlatformAdmin
}

func RegisterPlatformAdmin(i IPlatformAdmin) {
	localPlatformAdmin = i
}
