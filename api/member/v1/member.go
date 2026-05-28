package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

type ListReq struct {
	g.Meta `path:"/tenants/{tenant}/members" tags:"Member" method:"get" summary:"List tenant members"`
	Tenant string `v:"required"`
}

type AddReq struct {
	g.Meta `path:"/tenants/{tenant}/members" tags:"Member" method:"post" summary:"Add tenant member"`
	Tenant string `v:"required"`
	UserID string `json:"user_id" v:"required"`
	Role   string `json:"role" v:"required"`
	Status string `json:"status"`
}

type UpdateReq struct {
	g.Meta `path:"/tenants/{tenant}/members/{user}" tags:"Member" method:"patch" summary:"Update tenant member"`
	Tenant string `v:"required"`
	User   string `v:"required"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

type RemoveReq struct {
	g.Meta `path:"/tenants/{tenant}/members/{user}" tags:"Member" method:"delete" summary:"Remove tenant member"`
	Tenant string `v:"required"`
	User   string `v:"required"`
}

type ListRes struct {
	Members []service.TenantMembershipWithUser `json:"members"`
}

type MemberRes struct {
	Member *service.TenantMembership `json:"member"`
}

type ActionRes struct {
	OK bool `json:"ok"`
}
