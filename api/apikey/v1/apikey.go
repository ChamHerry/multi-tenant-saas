package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
)

type CreatePersonalReq struct {
	g.Meta    `path:"/tenants/{tenant}/api-keys" tags:"APIKey" method:"post" summary:"Create tenant-scoped API key"`
	Tenant    string   `v:"required"`
	Name      string   `json:"name" v:"required"`
	Scopes    []string `json:"scopes" v:"required"`
	ExpiresAt string   `json:"expires_at"`
}

type ListPersonalReq struct {
	g.Meta `path:"/tenants/{tenant}/api-keys" tags:"APIKey" method:"get" summary:"List my tenant-scoped API keys"`
	Tenant string `v:"required"`
}

type RevokePersonalReq struct {
	g.Meta `path:"/tenants/{tenant}/api-keys/{apiKey}" tags:"APIKey" method:"delete" summary:"Revoke my tenant-scoped API key"`
	Tenant string `v:"required"`
	APIKey string `v:"required"`
}

type CreateRes struct {
	APIKey *service.APIKey `json:"api_key"`
	RawKey string          `json:"raw_key"`
}

type ListRes struct {
	APIKeys []service.APIKey `json:"api_keys"`
}

type ActionRes struct {
	OK bool `json:"ok"`
}
