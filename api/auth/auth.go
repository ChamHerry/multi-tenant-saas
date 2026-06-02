package auth

import (
	"context"

	"multi-tenant-saas/api/auth/v1"
)

type IAuthPublicV1 interface {
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.AuthUserRes, err error)
	Register(ctx context.Context, req *v1.RegisterReq) (res *v1.AuthUserRes, err error)
	VerifyEmail(ctx context.Context, req *v1.VerifyEmailReq) (res *v1.VerifyEmailRes, err error)
	ResendVerification(ctx context.Context, req *v1.ResendVerificationReq) (res *v1.ResendVerificationRes, err error)
	ForgotPassword(ctx context.Context, req *v1.ForgotPasswordReq) (res *v1.ForgotPasswordRes, err error)
	ResetPassword(ctx context.Context, req *v1.ResetPasswordReq) (res *v1.ResetPasswordRes, err error)
}

type IAuthV1 interface {
	Session(ctx context.Context, req *v1.SessionReq) (res *v1.SessionRes, err error)
	Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.ActionRes, err error)
	ChangePassword(ctx context.Context, req *v1.ChangePasswordReq) (res *v1.ActionRes, err error)
}

type IAuthOAuthV1 interface {
	OAuthProviders(ctx context.Context, req *v1.OAuthProvidersReq) (res *v1.OAuthProvidersRes, err error)
	OAuthLogin(ctx context.Context, req *v1.OAuthLoginReq) (res *v1.OAuthLoginRes, err error)
	OAuthCallback(ctx context.Context, req *v1.OAuthCallbackReq) (res *v1.OAuthCallbackRes, err error)
}
