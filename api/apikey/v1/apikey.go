package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

type CreatePersonalReq struct {
	g.Meta    `path:"/api-keys" tags:"APIKey" method:"post" summary:"Create personal API key"`
	Name      string                                 `json:"name" v:"required"`
	Scopes    []string                               `json:"scopes" v:"required"`
	Grants    []service.CreateAPIKeyTenantGrantInput `json:"grants"`
	ExpiresAt string                                 `json:"expires_at"`
}

type ListPersonalReq struct {
	g.Meta `path:"/api-keys" tags:"APIKey" method:"get" summary:"List my personal API keys"`
}

type RevokePersonalReq struct {
	g.Meta `path:"/api-keys/{apiKey}" tags:"APIKey" method:"delete" summary:"Revoke my personal API key"`
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
