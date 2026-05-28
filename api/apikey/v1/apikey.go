package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

type CreateReq struct {
	g.Meta    `path:"/api-keys" tags:"APIKey" method:"post" summary:"Create API key"`
	Name      string   `json:"name" v:"required"`
	Scopes    []string `json:"scopes" v:"required"`
	ExpiresAt string   `json:"expires_at"`
}

type ListReq struct {
	g.Meta `path:"/api-keys" tags:"APIKey" method:"get" summary:"List API keys"`
}

type RevokeReq struct {
	g.Meta `path:"/api-keys/{apiKey}" tags:"APIKey" method:"delete" summary:"Revoke API key"`
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
