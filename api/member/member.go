package member

import (
	"context"

	"repomind-temp/api/member/v1"
)

type IMemberV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Add(ctx context.Context, req *v1.AddReq) (res *v1.MemberRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.MemberRes, err error)
	Remove(ctx context.Context, req *v1.RemoveReq) (res *v1.ActionRes, err error)
}
