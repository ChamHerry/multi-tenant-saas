package service

import (
	"context"
	"time"
)

type AuditLog struct {
	ID           string         `json:"id"`
	TenantID     string         `json:"tenant_id,omitempty"`
	UserID       string         `json:"user_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id,omitempty"`
	IP           string         `json:"ip,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	Metadata     map[string]any `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
}

type AuditLogFilter struct {
	TenantID     string
	UserID       string
	Action       string
	ResourceType string
	ResourceID   string
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

type AuditLogList struct {
	Logs  []AuditLog `json:"logs"`
	Total int        `json:"total"`
}

type AuditExportJob struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Format      string `json:"format,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	FileName    string `json:"file_name,omitempty"`
	Content     string `json:"content,omitempty"`
}

type IAuditQuery interface {
	List(ctx context.Context, filter AuditLogFilter) (*AuditLogList, error)
	Export(ctx context.Context, filter AuditLogFilter) (*AuditExportJob, error)
}

var localAuditQuery IAuditQuery

func AuditQuery() IAuditQuery {
	if localAuditQuery == nil {
		panic("implement not found for interface IAuditQuery")
	}
	return localAuditQuery
}

func RegisterAuditQuery(i IAuditQuery) {
	localAuditQuery = i
}
