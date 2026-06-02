package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
)

type LoginReq struct {
	g.Meta   `path:"/auth/login" tags:"Auth" method:"post" summary:"Login with email and password"`
	Email    string `json:"email" v:"required"`
	Password string `json:"password" v:"required"`
}

type RegisterReq struct {
	g.Meta      `path:"/auth/register" tags:"Auth" method:"post" summary:"Register password user"`
	Email       string `json:"email" v:"required"`
	Password    string `json:"password" v:"required"`
	DisplayName string `json:"display_name"`
}

type SessionReq struct {
	g.Meta `path:"/auth/session" tags:"Auth" method:"get" summary:"Get current auth session"`
}

type LogoutReq struct {
	g.Meta `path:"/auth/logout" tags:"Auth" method:"post" summary:"Logout current session"`
}

type ChangePasswordReq struct {
	g.Meta      `path:"/auth/password/change" tags:"Auth" method:"post" summary:"Change password"`
	OldPassword string `json:"old_password" v:"required"`
	NewPassword string `json:"new_password" v:"required"`
}

type AuthUserRes struct {
	User *service.User `json:"user"`
}

type SessionRes struct {
	User    *service.User        `json:"user"`
	Session *service.AuthSession `json:"session,omitempty"`
}

type VerifyEmailReq struct {
	g.Meta `path:"/auth/verify-email" tags:"Auth" method:"post" summary:"Verify email address"`
	Token  string `json:"token" v:"required"`
}

type VerifyEmailRes struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type ResendVerificationReq struct {
	g.Meta `path:"/auth/resend-verification" tags:"Auth" method:"post" summary:"Resend email verification"`
	Email  string `json:"email" v:"required|email"`
}

type ResendVerificationRes struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type ForgotPasswordReq struct {
	g.Meta `path:"/auth/forgot-password" tags:"Auth" method:"post" summary:"Request password reset email"`
	Email  string `json:"email" v:"required|email"`
}

type ForgotPasswordRes struct {
	OK bool `json:"ok"`
}

type ResetPasswordReq struct {
	g.Meta      `path:"/auth/reset-password" tags:"Auth" method:"post" summary:"Reset password with token"`
	Token       string `json:"token" v:"required|length:64,64"`
	NewPassword string `json:"new_password" v:"required"`
}

type ResetPasswordRes struct {
	OK bool `json:"ok"`
}

type ActionRes struct {
	OK bool `json:"ok"`
}
