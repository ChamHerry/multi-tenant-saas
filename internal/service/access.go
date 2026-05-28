package service

import "context"

type TenantAccess struct {
	TenantID     string       `json:"tenant_id"`
	TenantName   string       `json:"tenant_name"`
	TenantSlug   string       `json:"tenant_slug"`
	TenantStatus string       `json:"tenant_status"`
	Role         string       `json:"role"`
	Status       string       `json:"status"`
	Permissions  []Permission `json:"permissions"`
}

type AccessSnapshot struct {
	User          *User                 `json:"user"`
	Tenants       []TenantAccess        `json:"tenants"`
	PlatformAdmin *PlatformAdminContext `json:"platform_admin"`
}

type IAccess interface {
	Snapshot(ctx context.Context, userID string) (*AccessSnapshot, error)
}

var localAccess IAccess

func Access() IAccess {
	if localAccess == nil {
		panic("implement not found for interface IAccess")
	}
	return localAccess
}

func RegisterAccess(i IAccess) {
	localAccess = i
}
