package access

import (
	"context"

	"multi-tenant-saas/internal/service"
)

type sAccess struct{}

func init() {
	service.RegisterAccess(&sAccess{})
}

func (s *sAccess) Snapshot(ctx context.Context, userID string) (*service.AccessSnapshot, error) {
	user, err := service.UserService().GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	memberships, err := service.TenantMembershipService().ListUserTenants(ctx, userID)
	if err != nil {
		return nil, err
	}
	tenants := make([]service.TenantAccess, 0, len(memberships))
	for _, membership := range memberships {
		permissions := []service.Permission{}
		if membership.Status == "active" && membership.TenantStatus == "active" {
			permissions = service.RBAC().PermissionsForRole(membership.Role)
		}
		tenants = append(tenants, service.TenantAccess{
			TenantID:     membership.TenantID,
			TenantName:   membership.TenantName,
			TenantSlug:   membership.TenantSlug,
			TenantStatus: membership.TenantStatus,
			Role:         membership.Role,
			Status:       membership.Status,
			Permissions:  permissions,
		})
	}
	platformAdmin, err := service.PlatformAdminService().ResolveOptional(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &service.AccessSnapshot{
		User:          user,
		Tenants:       tenants,
		PlatformAdmin: platformAdmin,
	}, nil
}
