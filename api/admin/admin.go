package admin

import (
	"context"

	"multi-tenant-saas/api/admin/v1"
)

type IAdminV1 interface {
	Session(ctx context.Context, req *v1.SessionReq) (res *v1.SessionRes, err error)
	ListTenants(ctx context.Context, req *v1.ListTenantsReq) (res *v1.ListTenantsRes, err error)
	GetTenant(ctx context.Context, req *v1.GetTenantReq) (res *v1.TenantRes, err error)
	SuspendTenant(ctx context.Context, req *v1.SuspendTenantReq) (res *v1.ActionRes, err error)
	RestoreTenant(ctx context.Context, req *v1.RestoreTenantReq) (res *v1.ActionRes, err error)
	ListUsers(ctx context.Context, req *v1.ListUsersReq) (res *v1.ListUsersRes, err error)
	UpdateUser(ctx context.Context, req *v1.UpdateUserReq) (res *v1.ActionRes, err error)
	UnlockUser(ctx context.Context, req *v1.UnlockUserReq) (res *v1.ActionRes, err error)
	ListPlatformAdmins(ctx context.Context, req *v1.ListPlatformAdminsReq) (res *v1.PlatformAdminListRes, err error)
	GrantPlatformAdmin(ctx context.Context, req *v1.GrantPlatformAdminReq) (res *v1.ActionRes, err error)
	RevokePlatformAdmin(ctx context.Context, req *v1.RevokePlatformAdminReq) (res *v1.ActionRes, err error)
	ListAuditLogs(ctx context.Context, req *v1.ListAuditLogsReq) (res *v1.AuditListRes, err error)
	ListSystemConfigs(ctx context.Context, req *v1.ListSystemConfigsReq) (res *v1.SystemConfigListRes, err error)
	GetSystemConfig(ctx context.Context, req *v1.GetSystemConfigReq) (res *v1.SystemConfigRes, err error)
	UpsertSystemConfig(ctx context.Context, req *v1.UpsertSystemConfigReq) (res *v1.SystemConfigRes, err error)
	DeleteSystemConfig(ctx context.Context, req *v1.DeleteSystemConfigReq) (res *v1.ActionRes, err error)
}
