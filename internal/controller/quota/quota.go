package quota

import (
	"context"

	apiquota "repomind-temp/api/quota"
	"repomind-temp/api/quota/v1"
	"repomind-temp/internal/service"
)

type ControllerV1 struct{}

func NewV1() apiquota.IQuotaV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionBillingRead); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	quota, err := service.Quota().GetTenantQuota(ctx, tc.TenantID)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{Quota: quota}, nil
}

func (c *ControllerV1) Recalculate(ctx context.Context, req *v1.RecalculateReq) (res *v1.ActionRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionBillingManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.Quota().Recalculate(ctx, tc.TenantID); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}
