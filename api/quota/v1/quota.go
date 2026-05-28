package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

type GetReq struct {
	g.Meta `path:"/tenants/{tenant}/quota" tags:"Quota" method:"get" summary:"Get tenant quota"`
	Tenant string `v:"required"`
}

type RecalculateReq struct {
	g.Meta `path:"/tenants/{tenant}/usage/recalculate" tags:"Quota" method:"post" summary:"Recalculate tenant usage"`
	Tenant string `v:"required"`
}

type GetRes struct {
	Quota *service.TenantQuotaView `json:"quota"`
}

type ActionRes struct {
	OK bool `json:"ok"`
}
