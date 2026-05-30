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

type ActionRes struct {
	OK bool `json:"ok"`
}
