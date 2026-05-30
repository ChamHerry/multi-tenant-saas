package member

import (
	"context"

	apimember "multi-tenant-saas/api/member"
	"multi-tenant-saas/api/member/v1"
	"multi-tenant-saas/internal/service"
)

type ControllerV1 struct{}

func NewV1() apimember.IMemberV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionMemberRead); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	members, err := service.TenantMembershipService().ListTenantMembers(ctx, tc.TenantID)
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{Members: members}, nil
}

func (c *ControllerV1) Add(ctx context.Context, req *v1.AddReq) (res *v1.MemberRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionMemberManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	member, err := service.TenantMembershipService().AddMember(ctx, service.AddTenantMemberInput{
		TenantID:        tc.TenantID,
		UserID:          req.UserID,
		Role:            req.Role,
		Status:          req.Status,
		InvitedByUserID: tc.UserID,
	})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "member.add", ResourceType: "tenant_member", ResourceID: req.UserID})
	return &v1.MemberRes{Member: member}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.MemberRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionMemberManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	member, err := service.TenantMembershipService().UpdateMember(ctx, tc.TenantID, req.User, service.UpdateTenantMemberInput{Role: req.Role, Status: req.Status})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "member.update", ResourceType: "tenant_member", ResourceID: req.User})
	return &v1.MemberRes{Member: member}, nil
}

func (c *ControllerV1) Remove(ctx context.Context, req *v1.RemoveReq) (res *v1.ActionRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionMemberManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.TenantMembershipService().RemoveMember(ctx, tc.TenantID, req.User); err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "member.remove", ResourceType: "tenant_member", ResourceID: req.User})
	return &v1.ActionRes{OK: true}, nil
}

func (c *ControllerV1) BatchAdd(ctx context.Context, req *v1.BatchAddReq) (res *v1.BatchAddRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionMemberManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	added := 0
	var errs []v1.BatchAddError
	for _, m := range req.Members {
		_, addErr := service.TenantMembershipService().AddMember(ctx, service.AddTenantMemberInput{
			TenantID:        tc.TenantID,
			UserID:          m.UserID,
			Role:            m.Role,
			Status:          "active",
			InvitedByUserID: tc.UserID,
		})
		if addErr != nil {
			errs = append(errs, v1.BatchAddError{UserID: m.UserID, Error: addErr.Error()})
		} else {
			added++
		}
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "member.batch_add", ResourceType: "tenant_member", Metadata: map[string]any{"added": added, "errors": len(errs)}})
	return &v1.BatchAddRes{Added: added, Errors: errs}, nil
}
