package apikey

import (
	"context"
	"time"

	apiapikey "repomind-temp/api/apikey"
	"repomind-temp/api/apikey/v1"
	"repomind-temp/internal/service"
)

type ControllerV1 struct{}

func NewV1() apiapikey.IAPIKeyV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionAPIKeyManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.Quota().Require(ctx, tc.TenantID, service.MetricAPIKeyCount, 1); err != nil {
		return nil, err
	}
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, err
		}
		expiresAt = &parsed
	}
	created, err := service.APIKeyService().Create(ctx, service.CreateAPIKeyInput{
		TenantID:  tc.TenantID,
		UserID:    tc.UserID,
		Name:      req.Name,
		Scopes:    req.Scopes,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "api_key.create", ResourceType: "api_key", ResourceID: created.ID})
	return &v1.CreateRes{APIKey: &created.APIKey, RawKey: created.RawKey}, nil
}

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionAPIKeyManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := service.APIKeyService().List(ctx, tc.TenantID)
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{APIKeys: items}, nil
}

func (c *ControllerV1) Revoke(ctx context.Context, req *v1.RevokeReq) (res *v1.ActionRes, err error) {
	if err = service.RBAC().Require(ctx, service.PermissionAPIKeyManage); err != nil {
		return nil, err
	}
	tc, err := service.MustTenantContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.APIKeyService().Revoke(ctx, tc.TenantID, req.APIKey); err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "api_key.revoke", ResourceType: "api_key", ResourceID: req.APIKey})
	return &v1.ActionRes{OK: true}, nil
}
