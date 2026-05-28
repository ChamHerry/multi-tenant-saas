package member

import (
	"context"

	apimember "repomind-temp/api/member"
	"repomind-temp/api/member/v1"
	"repomind-temp/internal/service"
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
