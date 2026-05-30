package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
)

type SessionReq struct {
	g.Meta `path:"/admin/session" tags:"Admin" method:"get" summary:"Get platform admin session"`
}

type ListTenantsReq struct {
	g.Meta `path:"/admin/tenants" tags:"Admin" method:"get" summary:"List all tenants"`
	Query  string `json:"query"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type GetTenantReq struct {
	g.Meta `path:"/admin/tenants/{tenant}" tags:"Admin" method:"get" summary:"Get tenant as platform admin"`
	Tenant string `v:"required"`
}

type SuspendTenantReq struct {
	g.Meta `path:"/admin/tenants/{tenant}/suspend" tags:"Admin" method:"post" summary:"Suspend tenant as platform admin"`
	Tenant string `v:"required"`
}

type RestoreTenantReq struct {
	g.Meta `path:"/admin/tenants/{tenant}/restore" tags:"Admin" method:"post" summary:"Restore tenant as platform admin"`
	Tenant string `v:"required"`
}

type ListUsersReq struct {
	g.Meta `path:"/admin/users" tags:"Admin" method:"get" summary:"List users as platform admin"`
	Query  string `json:"query"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type UpdateUserReq struct {
	g.Meta `path:"/admin/users/{user}/status" tags:"Admin" method:"patch" summary:"Update user status as platform admin"`
	User   string `v:"required"`
	Status string `json:"status" v:"required"`
}

type ListPlatformAdminsReq struct {
	g.Meta `path:"/admin/platform-admins" tags:"Admin" method:"get" summary:"List platform admins"`
	Query  string `json:"query"`
	Role   string `json:"role"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type GrantPlatformAdminReq struct {
	g.Meta `path:"/admin/platform-admins" tags:"Admin" method:"post" summary:"Grant platform admin"`
	UserID string `json:"user_id" v:"required"`
	Role   string `json:"role" v:"required"`
}

type RevokePlatformAdminReq struct {
	g.Meta `path:"/admin/platform-admins/{user}" tags:"Admin" method:"delete" summary:"Revoke platform admin"`
	User   string `v:"required"`
}

type ListAuditLogsReq struct {
	g.Meta       `path:"/admin/audit-logs" tags:"Admin" method:"get" summary:"List platform audit logs"`
	TenantID     string `json:"tenant_id"`
	UserID       string `json:"user_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	From         string `json:"from"`
	To           string `json:"to"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

type SessionRes struct {
	PlatformAdmin *service.PlatformAdminContext `json:"platform_admin"`
}

type ListTenantsRes struct {
	Items []service.Tenant `json:"items"`
	Total int              `json:"total"`
}

type TenantRes struct {
	Tenant *service.Tenant `json:"tenant"`
}

type ListUsersRes struct {
	Items []service.User `json:"items"`
	Total int            `json:"total"`
}

type PlatformAdminListRes struct {
	Items []service.PlatformAdmin `json:"items"`
	Total int                     `json:"total"`
}

type AuditListRes struct {
	Logs  []service.AuditLog `json:"logs"`
	Total int                `json:"total"`
}

type ActionRes struct {
	OK bool `json:"ok"`
}
