package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
)

type StateReq struct {
	g.Meta `path:"/setup/state" tags:"Setup" method:"get" summary:"Get initial setup state"`
}

type SetupAdminPayload struct {
	Email       string `json:"email" v:"required|email"`
	Password    string `json:"password" v:"required"`
	DisplayName string `json:"display_name"`
}

type SetupRuntimePayload struct {
	WebBaseURL            string `json:"web_base_url" v:"required"`
	GenerateSessionSecret bool   `json:"generate_session_secret"`
	GenerateAPIKeySecret  bool   `json:"generate_api_key_secret"`
}

type CompleteReq struct {
	g.Meta  `path:"/setup/complete" tags:"Setup" method:"post" summary:"Complete initial system setup"`
	Admin   SetupAdminPayload   `json:"admin" v:"required"`
	Runtime SetupRuntimePayload `json:"runtime" v:"required"`
}

type StateRes = service.SetupStatus

type CompleteRes struct {
	Initialized bool          `json:"initialized"`
	User        *service.User `json:"user"`
}
