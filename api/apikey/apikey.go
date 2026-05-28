package apikey

import (
	"context"

	"repomind-temp/api/apikey/v1"
)

type IAPIKeyV1 interface {
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Revoke(ctx context.Context, req *v1.RevokeReq) (res *v1.ActionRes, err error)
}
