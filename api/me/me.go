package me

import (
	"context"

	"repomind-temp/api/me/v1"
)

type IMeV1 interface {
	Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error)
	Tenants(ctx context.Context, req *v1.TenantsReq) (res *v1.TenantsRes, err error)
	Access(ctx context.Context, req *v1.AccessReq) (res *v1.AccessRes, err error)
	TenantContext(ctx context.Context, req *v1.TenantContextReq) (res *v1.TenantContextRes, err error)
}
