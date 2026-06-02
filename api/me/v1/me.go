package v1

import (
	"time"

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

type SessionDevice struct {
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version,omitempty"`
	OS             string `json:"os"`
	IsMobile       bool   `json:"is_mobile"`
}

type SessionItem struct {
	SessionID    string        `json:"session_id"`
	IsCurrent    bool          `json:"is_current"`
	Device       SessionDevice `json:"device"`
	IP           string        `json:"ip"`
	Location     *string       `json:"location"`
	LastActiveAt time.Time     `json:"last_active_at"`
	CreatedAt    time.Time     `json:"created_at"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

type ListSessionsReq struct {
	g.Meta `path:"/me/sessions" tags:"Me" method:"get" summary:"List current user active sessions"`
}

type ListSessionsRes struct {
	Sessions []SessionItem `json:"sessions"`
}

type RevokeSessionReq struct {
	g.Meta  `path:"/me/sessions/{session}" tags:"Me" method:"delete" summary:"Revoke another current user session"`
	Session string `v:"required"`
}

type RevokeSessionRes struct {
	OK bool `json:"ok"`
}

type RevokeOtherSessionsReq struct {
	g.Meta `path:"/me/sessions" tags:"Me" method:"delete" summary:"Revoke all other current user sessions"`
}

type RevokeOtherSessionsRes struct {
	OK           bool `json:"ok"`
	RevokedCount int  `json:"revoked_count"`
}
