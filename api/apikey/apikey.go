package apikey

import (
	"context"

	"repomind-temp/api/apikey/v1"
)

type IAPIKeyV1 interface {
	CreatePersonal(ctx context.Context, req *v1.CreatePersonalReq) (res *v1.CreateRes, err error)
	ListPersonal(ctx context.Context, req *v1.ListPersonalReq) (res *v1.ListRes, err error)
	RevokePersonal(ctx context.Context, req *v1.RevokePersonalReq) (res *v1.ActionRes, err error)
}
