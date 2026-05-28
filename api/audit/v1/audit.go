package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

type ListTenantReq struct {
	g.Meta       `path:"/tenants/{tenant}/audit-logs" tags:"Audit" method:"get" summary:"List tenant audit logs"`
	Tenant       string `v:"required"`
	UserID       string `json:"user_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	From         string `json:"from"`
	To           string `json:"to"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

type ExportTenantReq struct {
	g.Meta       `path:"/tenants/{tenant}/audit-logs/export" tags:"Audit" method:"post" summary:"Export tenant audit logs"`
	Tenant       string `v:"required"`
	UserID       string `json:"user_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	From         string `json:"from"`
	To           string `json:"to"`
}

type SecurityEventsReq struct {
	g.Meta `path:"/me/security-events" tags:"Audit" method:"get" summary:"List my security events"`
	Action string `json:"action"`
	From   string `json:"from"`
	To     string `json:"to"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type ListRes struct {
	Logs  []service.AuditLog `json:"logs"`
	Total int                `json:"total"`
}

type ExportRes struct {
	Job *service.AuditExportJob `json:"job"`
}
