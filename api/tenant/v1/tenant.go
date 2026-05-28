package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

type CreateReq struct {
	g.Meta       `path:"/tenants" tags:"Tenant" method:"post" summary:"Create tenant"`
	Name         string         `json:"name" v:"required"`
	Slug         string         `json:"slug" v:"required"`
	Plan         string         `json:"plan"`
	MaxRepos     int            `json:"max_repos"`
	MaxSymbols   int            `json:"max_symbols"`
	MaxStorageMB int            `json:"max_storage_mb"`
	Metadata     map[string]any `json:"metadata"`
}

type GetReq struct {
	g.Meta `path:"/tenants/{tenant}" tags:"Tenant" method:"get" summary:"Get tenant"`
	Tenant string `v:"required"`
}

type UpdateReq struct {
	g.Meta       `path:"/tenants/{tenant}" tags:"Tenant" method:"patch" summary:"Update tenant"`
	Tenant       string         `v:"required"`
	Name         string         `json:"name"`
	Slug         string         `json:"slug"`
	Plan         string         `json:"plan"`
	MaxRepos     int            `json:"max_repos"`
	MaxSymbols   int            `json:"max_symbols"`
	MaxStorageMB int            `json:"max_storage_mb"`
	Metadata     map[string]any `json:"metadata"`
}

type SuspendReq struct {
	g.Meta `path:"/tenants/{tenant}/suspend" tags:"Tenant" method:"post" summary:"Suspend tenant"`
	Tenant string `v:"required"`
}

type RestoreReq struct {
	g.Meta `path:"/tenants/{tenant}/restore" tags:"Tenant" method:"post" summary:"Restore tenant"`
	Tenant string `v:"required"`
}

type DeleteReq struct {
	g.Meta `path:"/tenants/{tenant}" tags:"Tenant" method:"delete" summary:"Soft delete tenant"`
	Tenant string `v:"required"`
}

type TenantRes struct {
	Tenant *service.Tenant `json:"tenant"`
}

type ActionRes struct {
	OK bool `json:"ok"`
}
