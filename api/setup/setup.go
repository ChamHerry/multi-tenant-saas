package setup

import (
	"context"

	"multi-tenant-saas/api/setup/v1"
)

type ISetupV1 interface {
	State(ctx context.Context, req *v1.StateReq) (res *v1.StateRes, err error)
	Complete(ctx context.Context, req *v1.CompleteReq) (res *v1.CompleteRes, err error)
}
