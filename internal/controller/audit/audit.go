package audit

import (
	"context"
	"time"

	apiaudit "repomind-temp/api/audit"
	"repomind-temp/api/audit/v1"
	"repomind-temp/internal/service"
)

type ControllerV1 struct{}

func NewV1() apiaudit.IAuditV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) ListTenant(ctx context.Context, req *v1.ListTenantReq) (res *v1.ListRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionAuditRead); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	filter, err := auditFilter(req.UserID, tc.TenantID, req.Action, req.ResourceType, req.ResourceID, req.From, req.To, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	logs, err := service.AuditQuery().List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{Logs: logs.Logs, Total: logs.Total}, nil
}

func (c *ControllerV1) ExportTenant(ctx context.Context, req *v1.ExportTenantReq) (res *v1.ExportRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionAuditRead); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	filter, err := auditFilter(req.UserID, tc.TenantID, req.Action, req.ResourceType, req.ResourceID, req.From, req.To, 0, 0)
	if err != nil {
		return nil, err
	}
	job, err := service.AuditQuery().Export(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &v1.ExportRes{Job: job}, nil
}

func (c *ControllerV1) SecurityEvents(ctx context.Context, req *v1.SecurityEventsReq) (res *v1.ListRes, err error) {
	if err = service.RBAC().RequireAuthScope(ctx, service.PermissionUserSecurityRead); err != nil {
		return nil, err
	}
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	filter, err := auditFilter(identity.UserID, "", req.Action, "", "", req.From, req.To, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	logs, err := service.AuditQuery().List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{Logs: logs.Logs, Total: logs.Total}, nil
}

func auditFilter(userID, tenantID, action, resourceType, resourceID, from, to string, limit, offset int) (service.AuditLogFilter, error) {
	filter := service.AuditLogFilter{TenantID: tenantID, UserID: userID, Action: action, ResourceType: resourceType, ResourceID: resourceID, Limit: limit, Offset: offset}
	if from != "" {
		parsed, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return filter, err
		}
		filter.From = &parsed
	}
	if to != "" {
		parsed, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return filter, err
		}
		filter.To = &parsed
	}
	return filter, nil
}
