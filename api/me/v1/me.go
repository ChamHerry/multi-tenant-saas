package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
)

type MeReq struct {
	g.Meta `path:"/me" tags:"Me" method:"get" summary:"Get current user"`
}

type MeRes struct {
	User *service.User `json:"user"`
}

type TenantsReq struct {
	g.Meta `path:"/me/tenants" tags:"Me" method:"get" summary:"List current user tenants"`
}

type TenantsRes struct {
	Tenants []service.TenantMembershipWithTenant `json:"tenants"`
}

type AccessReq struct {
	g.Meta `path:"/me/access" tags:"Me" method:"get" summary:"Get current user access snapshot"`
}

type AccessRes struct {
	User          *service.User                 `json:"user"`
	Tenants       []service.TenantAccess        `json:"tenants"`
	PlatformAdmin *service.PlatformAdminContext `json:"platform_admin"`
}

type TenantContextReq struct {
	g.Meta `path:"/tenant-context" tags:"Me" method:"get" summary:"Get resolved tenant context"`
}

type TenantContextRes struct {
	TenantContext *service.TenantContext `json:"tenant_context"`
}

type UpdateMeReq struct {
	g.Meta      `path:"/me" tags:"Me" method:"patch" summary:"Update current user profile"`
	DisplayName string `json:"display_name" v:"required|length:1,100" dc:"Display name"`
	AvatarURL   string `json:"avatar_url" v:"url" dc:"Avatar URL"`
}

type UpdateMeRes struct {
	User *service.User `json:"user"`
}
