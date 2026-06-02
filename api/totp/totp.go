package totp

import (
	"context"

	"multi-tenant-saas/api/totp/v1"
)

type IAuthPublicV1 interface {
	VerifyTOTP(ctx context.Context, req *v1.VerifyTOTPReq) (res *v1.AuthUserRes, err error)
}

type IMeTotpV1 interface {
	Setup(ctx context.Context, req *v1.SetupReq) (res *v1.SetupRes, err error)
	Enable(ctx context.Context, req *v1.EnableReq) (res *v1.EnableRes, err error)
	Disable(ctx context.Context, req *v1.DisableReq) (res *v1.ActionRes, err error)
	RegenerateBackupCodes(ctx context.Context, req *v1.RegenerateBackupCodesReq) (res *v1.RegenerateBackupCodesRes, err error)
	Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error)
}
