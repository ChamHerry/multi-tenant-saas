package me

import (
	"context"

	apime "repomind-temp/api/me"
	"repomind-temp/api/me/v1"
	"repomind-temp/internal/service"
)

type ControllerV1 struct{}

func NewV1() apime.IMeV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	user, err := service.UserService().GetUser(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.MeRes{User: user}, nil
}

func (c *ControllerV1) Tenants(ctx context.Context, req *v1.TenantsReq) (res *v1.TenantsRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	tenants, err := service.TenantMembershipService().ListUserTenants(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.TenantsRes{Tenants: tenants}, nil
}

func (c *ControllerV1) Access(ctx context.Context, req *v1.AccessReq) (res *v1.AccessRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	access, err := service.Access().Snapshot(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.AccessRes{
		User:          access.User,
		Tenants:       access.Tenants,
		PlatformAdmin: access.PlatformAdmin,
	}, nil
}

func (c *ControllerV1) TenantContext(ctx context.Context, req *v1.TenantContextReq) (res *v1.TenantContextRes, err error) {
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.TenantContextRes{TenantContext: tc}, nil
}
