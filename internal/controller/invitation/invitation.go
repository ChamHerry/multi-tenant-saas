package invitation

import (
	"context"
	"time"

	apiinvitation "multi-tenant-saas/api/invitation"
	"multi-tenant-saas/api/invitation/v1"
	"multi-tenant-saas/internal/service"
)

type ControllerV1 struct{}

func NewV1() apiinvitation.IInvitationV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) ListTenant(ctx context.Context, req *v1.ListTenantReq) (res *v1.ListRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionMemberRead); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := service.TenantInvitationService().ListTenantInvitations(ctx, tc.TenantID, service.TenantInvitationFilter{Status: req.Status, Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{Invitations: items.Invitations, Total: items.Total}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionInvitationManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	expires := time.Duration(req.ExpiresHours) * time.Hour
	created, err := service.TenantInvitationService().Create(ctx, service.CreateTenantInvitationInput{TenantID: tc.TenantID, InviteeEmail: req.InviteeEmail, Role: req.Role, Message: req.Message, InvitedByUserID: tc.UserID, ExpiresIn: expires})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Invitation: created.Invitation, Token: created.Token, AcceptURL: created.AcceptURL}, nil
}

func (c *ControllerV1) Resend(ctx context.Context, req *v1.ResendReq) (res *v1.CreateRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionInvitationManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	created, err := service.TenantInvitationService().Resend(ctx, tc.TenantID, req.Invitation, tc.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Invitation: created.Invitation, Token: created.Token, AcceptURL: created.AcceptURL}, nil
}

func (c *ControllerV1) Revoke(ctx context.Context, req *v1.RevokeReq) (res *v1.ActionRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionInvitationManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.TenantInvitationService().Revoke(ctx, tc.TenantID, req.Invitation, tc.UserID); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) ListMine(ctx context.Context, req *v1.ListMineReq) (res *v1.ListRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	user, err := service.UserService().GetUser(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	items, err := service.TenantInvitationService().ListMyInvitations(ctx, identity.UserID, user.Email, service.TenantInvitationFilter{Status: req.Status, Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{Invitations: items.Invitations, Total: items.Total}, nil
}

func (c *ControllerV1) Accept(ctx context.Context, req *v1.AcceptReq) (res *v1.AcceptRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	member, err := service.TenantInvitationService().Accept(ctx, req.Token, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.AcceptRes{Member: member}, nil
}

func (c *ControllerV1) Decline(ctx context.Context, req *v1.DeclineReq) (res *v1.ActionRes, err error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.TenantInvitationService().Decline(ctx, req.Invitation, identity.UserID); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) BatchCreate(ctx context.Context, req *v1.BatchCreateReq) (res *v1.BatchCreateRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionInvitationManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	created := 0
	var results []service.TenantInvitation
	var errs []v1.BatchCreateError
	for _, inv := range req.Invitations {
		item, createErr := service.TenantInvitationService().Create(ctx, service.CreateTenantInvitationInput{
			TenantID:        tc.TenantID,
			InviteeEmail:    inv.InviteeEmail,
			Role:            inv.Role,
			Message:         inv.Message,
			InvitedByUserID: tc.UserID,
		})
		if createErr != nil {
			errs = append(errs, v1.BatchCreateError{InviteeEmail: inv.InviteeEmail, Error: createErr.Error()})
		} else {
			created++
			results = append(results, item.Invitation)
		}
	}
	return &v1.BatchCreateRes{Created: created, Results: results, Errors: errs}, nil
}
