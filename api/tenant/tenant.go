package tenant

import (
	"context"

	"repomind-temp/api/tenant/v1"
)

type ITenantV1 interface {
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.TenantRes, err error)
	Get(ctx context.Context, req *v1.GetReq) (res *v1.TenantRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.TenantRes, err error)
	Suspend(ctx context.Context, req *v1.SuspendReq) (res *v1.ActionRes, err error)
	Restore(ctx context.Context, req *v1.RestoreReq) (res *v1.ActionRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.ActionRes, err error)
}
