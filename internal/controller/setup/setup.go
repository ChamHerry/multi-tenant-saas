package setup

import (
	"context"

	setupapi "multi-tenant-saas/api/setup"
	"multi-tenant-saas/api/setup/v1"
	"multi-tenant-saas/internal/service"
)

type ControllerV1 struct{}

func NewV1() setupapi.ISetupV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) State(ctx context.Context, req *v1.StateReq) (res *v1.StateRes, err error) {
	return service.SystemSetup().State(ctx)
}

func (c *ControllerV1) Complete(ctx context.Context, req *v1.CompleteReq) (res *v1.CompleteRes, err error) {
	result, err := service.SystemSetup().Complete(ctx, service.CompleteSetupInput{
		Admin: service.SetupAdminInput{
			Email:       req.Admin.Email,
			Password:    req.Admin.Password,
			DisplayName: req.Admin.DisplayName,
		},
		Runtime: service.SetupRuntimeInput{
			WebBaseURL:            req.Runtime.WebBaseURL,
			GenerateSessionSecret: req.Runtime.GenerateSessionSecret,
			GenerateAPIKeySecret:  req.Runtime.GenerateAPIKeySecret,
		},
		IP:        service.BizCtx().GetClientIP(ctx),
		UserAgent: service.BizCtx().GetUserAgent(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CompleteRes{Initialized: result.Initialized, User: result.User}, nil
}
