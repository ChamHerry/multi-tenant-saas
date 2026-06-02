package me

import (
	"context"

	"multi-tenant-saas/api/me/v1"
)

type IMeV1 interface {
	Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error)
	Tenants(ctx context.Context, req *v1.TenantsReq) (res *v1.TenantsRes, err error)
	Access(ctx context.Context, req *v1.AccessReq) (res *v1.AccessRes, err error)
	TenantContext(ctx context.Context, req *v1.TenantContextReq) (res *v1.TenantContextRes, err error)
	UpdateMe(ctx context.Context, req *v1.UpdateMeReq) (res *v1.UpdateMeRes, err error)
	ListSessions(ctx context.Context, req *v1.ListSessionsReq) (res *v1.ListSessionsRes, err error)
	RevokeSession(ctx context.Context, req *v1.RevokeSessionReq) (res *v1.RevokeSessionRes, err error)
	RevokeOtherSessions(ctx context.Context, req *v1.RevokeOtherSessionsReq) (res *v1.RevokeOtherSessionsRes, err error)
}
