package service

import "context"

type AuditLogInput struct {
	TenantID      string
	UserID        string
	Action        string
	ResourceType  string
	ResourceID    string
	IP            string
	UserAgent     string
	PlatformRole  string // filled by platform admin operations
	Metadata      map[string]any
}

type IAudit interface {
	Write(ctx context.Context, in AuditLogInput) error
}

var localAudit IAudit

func Audit() IAudit {
	if localAudit == nil {
		panic("implement not found for interface IAudit")
	}
	return localAudit
}

func RegisterAudit(i IAudit) {
	localAudit = i
}
