package invitation

import (
	"context"

	"repomind-temp/api/invitation/v1"
)

type IInvitationV1 interface {
	ListTenant(ctx context.Context, req *v1.ListTenantReq) (res *v1.ListRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Resend(ctx context.Context, req *v1.ResendReq) (res *v1.CreateRes, err error)
	Revoke(ctx context.Context, req *v1.RevokeReq) (res *v1.ActionRes, err error)
	ListMine(ctx context.Context, req *v1.ListMineReq) (res *v1.ListRes, err error)
	Accept(ctx context.Context, req *v1.AcceptReq) (res *v1.AcceptRes, err error)
	Decline(ctx context.Context, req *v1.DeclineReq) (res *v1.ActionRes, err error)
}
