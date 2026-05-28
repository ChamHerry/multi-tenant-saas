package tenant

import (
	"context"

	apitenant "repomind-temp/api/tenant"
	"repomind-temp/api/tenant/v1"
	"repomind-temp/internal/service"
)

type ControllerV1 struct{}

func NewV1() apitenant.ITenantV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.TenantRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	tenant, err := service.TenantProvision().CreateTenant(ctx, service.CreateTenantInput{
		Name:         req.Name,
		Slug:         req.Slug,
		OwnerUserID:  identity.UserID,
		Plan:         req.Plan,
		MaxRepos:     req.MaxRepos,
		MaxSymbols:   req.MaxSymbols,
		MaxStorageMB: req.MaxStorageMB,
		Metadata:     req.Metadata,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TenantRes{Tenant: tenant}, nil
}

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.TenantRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionTenantRead); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	tenant, err := service.TenantAdmin().GetTenant(ctx, tc.TenantID)
	if err != nil {
		return nil, err
	}
	return &v1.TenantRes{Tenant: tenant}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.TenantRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionTenantManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	tenant, err := service.TenantAdmin().UpdateTenant(ctx, tc.TenantID, service.UpdateTenantInput{
		Name:         req.Name,
		Slug:         req.Slug,
		Plan:         req.Plan,
		MaxRepos:     req.MaxRepos,
		MaxSymbols:   req.MaxSymbols,
		MaxStorageMB: req.MaxStorageMB,
		Metadata:     req.Metadata,
	})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "tenant.update", ResourceType: "tenant", ResourceID: tc.TenantID})
	return &v1.TenantRes{Tenant: tenant}, nil
}

func (c *ControllerV1) Suspend(ctx context.Context, req *v1.SuspendReq) (res *v1.ActionRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionTenantManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.TenantAdmin().SuspendTenant(ctx, tc.TenantID, tc.UserID); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) Restore(ctx context.Context, req *v1.RestoreReq) (res *v1.ActionRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionTenantManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.TenantAdmin().RestoreTenant(ctx, tc.TenantID, tc.UserID); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.ActionRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionTenantManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.TenantAdmin().DeleteTenant(ctx, tc.TenantID, tc.UserID); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}
