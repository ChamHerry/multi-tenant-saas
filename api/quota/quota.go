package quota

import (
	"context"

	"repomind-temp/api/quota/v1"
)

type IQuotaV1 interface {
	Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error)
	Recalculate(ctx context.Context, req *v1.RecalculateReq) (res *v1.ActionRes, err error)
}
