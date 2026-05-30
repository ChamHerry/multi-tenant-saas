package auth

import (
	"context"

	"multi-tenant-saas/api/auth/v1"
)

type IAuthPublicV1 interface {
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.AuthUserRes, err error)
	Register(ctx context.Context, req *v1.RegisterReq) (res *v1.AuthUserRes, err error)
}

type IAuthV1 interface {
	Session(ctx context.Context, req *v1.SessionReq) (res *v1.SessionRes, err error)
	Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.ActionRes, err error)
	ChangePassword(ctx context.Context, req *v1.ChangePasswordReq) (res *v1.ActionRes, err error)
}
