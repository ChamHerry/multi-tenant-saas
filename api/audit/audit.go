package audit

import (
	"context"

	"multi-tenant-saas/api/audit/v1"
)

type IAuditV1 interface {
	ListTenant(ctx context.Context, req *v1.ListTenantReq) (res *v1.ListRes, err error)
	ExportTenant(ctx context.Context, req *v1.ExportTenantReq) (res *v1.ExportRes, err error)
	SecurityEvents(ctx context.Context, req *v1.SecurityEventsReq) (res *v1.ListRes, err error)
}
