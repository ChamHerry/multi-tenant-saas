package admin

import (
	"context"
	"time"

	apiadmin "repomind-temp/api/admin"
	"repomind-temp/api/admin/v1"
	"repomind-temp/internal/service"
)

type ControllerV1 struct{}

func NewV1() apiadmin.IAdminV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) Session(ctx context.Context, req *v1.SessionReq) (res *v1.SessionRes, err error) {
	pac, err := service.MustPlatformAdminContext(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.SessionRes{PlatformAdmin: pac}, nil
}

func (c *ControllerV1) ListTenants(ctx context.Context, req *v1.ListTenantsReq) (res *v1.ListTenantsRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionTenantRead); err != nil {
		return nil, err
	}
	items, err := service.PlatformAdminService().ListTenants(ctx, service.PlatformTenantFilter{Query: req.Query, Status: req.Status, Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return nil, err
	}
	return &v1.ListTenantsRes{Items: items.Items, Total: items.Total}, nil
}

func (c *ControllerV1) GetTenant(ctx context.Context, req *v1.GetTenantReq) (res *v1.TenantRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionTenantRead); err != nil {
		return nil, err
	}
	tenant, err := service.TenantAdmin().GetTenant(ctx, req.Tenant)
	if err != nil {
		return nil, err
	}
	return &v1.TenantRes{Tenant: tenant}, nil
}

func (c *ControllerV1) SuspendTenant(ctx context.Context, req *v1.SuspendTenantReq) (res *v1.ActionRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionTenantManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	if err = service.TenantAdmin().SuspendTenant(ctx, req.Tenant, pac.UserID); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) RestoreTenant(ctx context.Context, req *v1.RestoreTenantReq) (res *v1.ActionRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionTenantManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	if err = service.TenantAdmin().RestoreTenant(ctx, req.Tenant, pac.UserID); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) ListUsers(ctx context.Context, req *v1.ListUsersReq) (res *v1.ListUsersRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionUserRead); err != nil {
		return nil, err
	}
	items, err := service.PlatformAdminService().ListUsers(ctx, service.PlatformUserFilter{Query: req.Query, Status: req.Status, Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return nil, err
	}
	return &v1.ListUsersRes{Items: items.Items, Total: items.Total}, nil
}

func (c *ControllerV1) UpdateUser(ctx context.Context, req *v1.UpdateUserReq) (res *v1.ActionRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionUserManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	if err = service.PlatformAdminService().UpdateUserStatus(ctx, pac.UserID, req.User, req.Status); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) ListPlatformAdmins(ctx context.Context, req *v1.ListPlatformAdminsReq) (res *v1.PlatformAdminListRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionAdminManage); err != nil {
		return nil, err
	}
	items, err := service.PlatformAdminService().ListPlatformAdmins(ctx, service.PlatformAdminFilter{Query: req.Query, Role: req.Role, Status: req.Status, Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return nil, err
	}
	return &v1.PlatformAdminListRes{Items: items.Items, Total: items.Total}, nil
}

func (c *ControllerV1) GrantPlatformAdmin(ctx context.Context, req *v1.GrantPlatformAdminReq) (res *v1.ActionRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionAdminManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	if err = service.PlatformAdminService().Grant(ctx, pac.UserID, req.UserID, req.Role); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) RevokePlatformAdmin(ctx context.Context, req *v1.RevokePlatformAdminReq) (res *v1.ActionRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionAdminManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	if err = service.PlatformAdminService().Revoke(ctx, pac.UserID, req.User); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) ListAuditLogs(ctx context.Context, req *v1.ListAuditLogsReq) (res *v1.AuditListRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionAuditRead); err != nil {
		return nil, err
	}
	filter := service.AuditLogFilter{TenantID: req.TenantID, UserID: req.UserID, Action: req.Action, ResourceType: req.ResourceType, ResourceID: req.ResourceID, Limit: req.Limit, Offset: req.Offset}
	if req.From != "" {
		parsed, err := time.Parse(time.RFC3339, req.From)
		if err != nil {
			return nil, err
		}
		filter.From = &parsed
	}
	if req.To != "" {
		parsed, err := time.Parse(time.RFC3339, req.To)
		if err != nil {
			return nil, err
		}
		filter.To = &parsed
	}
	logs, err := service.AuditQuery().List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &v1.AuditListRes{Logs: logs.Logs, Total: logs.Total}, nil
}

func (c *ControllerV1) ListPlans(ctx context.Context, req *v1.ListPlansReq) (res *v1.PlanListRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionBillingManage); err != nil {
		return nil, err
	}
	items, err := service.PlatformAdminService().ListPlans(ctx, service.PlanEntitlementFilter{Plan: req.Plan, FeatureKey: req.FeatureKey, Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return nil, err
	}
	return &v1.PlanListRes{Items: items.Items, Total: items.Total}, nil
}

func (c *ControllerV1) UpdateTenantPlan(ctx context.Context, req *v1.UpdateTenantPlanReq) (res *v1.ActionRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionBillingManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	if err = service.PlatformAdminService().UpdateTenantPlan(ctx, pac.UserID, req.Tenant, req.Plan); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) UpdateTenantQuota(ctx context.Context, req *v1.UpdateTenantQuotaReq) (res *v1.ActionRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionBillingManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	if err = service.PlatformAdminService().UpdateTenantQuota(ctx, pac.UserID, req.Tenant, service.UpdateTenantQuotaInput{MaxRepos: req.MaxRepos, MaxSymbols: req.MaxSymbols, MaxStorageMB: req.MaxStorageMB}); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}
