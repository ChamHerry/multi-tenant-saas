package admin

import (
	"context"

	"multi-tenant-saas/api/admin/v1"
	"multi-tenant-saas/internal/service"
)

func (c *ControllerV1) ListSystemConfigs(ctx context.Context, req *v1.ListSystemConfigsReq) (res *v1.SystemConfigListRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionConfigRead); err != nil {
		return nil, err
	}
	items, err := service.SystemConfigAdmin().List(ctx, service.SystemConfigAdminListFilter{
		Query:    req.Query,
		Category: req.Category,
		Limit:    req.Limit,
		Offset:   req.Offset,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SystemConfigListRes{Items: items.Items, Total: items.Total}, nil
}

func (c *ControllerV1) GetSystemConfig(ctx context.Context, req *v1.GetSystemConfigReq) (res *v1.SystemConfigRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionConfigRead); err != nil {
		return nil, err
	}
	item, err := service.SystemConfigAdmin().Get(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	return &v1.SystemConfigRes{Config: item}, nil
}

func (c *ControllerV1) UpsertSystemConfig(ctx context.Context, req *v1.UpsertSystemConfigReq) (res *v1.SystemConfigRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionConfigManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	actorUserID := ""
	if pac != nil {
		actorUserID = pac.UserID
	}
	item, err := service.SystemConfigAdmin().Upsert(ctx, service.SystemConfigAdminUpsertInput{
		Key:           req.Key,
		Value:         req.Value,
		ValueProvided: req.ValueProvided,
		ValueType:     req.ValueType,
		Description:   req.Description,
		ActorUserID:   actorUserID,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SystemConfigRes{Config: item}, nil
}

func (c *ControllerV1) DeleteSystemConfig(ctx context.Context, req *v1.DeleteSystemConfigReq) (res *v1.ActionRes, err error) {
	if err = service.PlatformAdminService().Require(ctx, service.PlatformPermissionConfigManage); err != nil {
		return nil, err
	}
	pac, _ := service.PlatformAdminContextFromCtx(ctx)
	actorUserID := ""
	if pac != nil {
		actorUserID = pac.UserID
	}
	if err = service.SystemConfigAdmin().Delete(ctx, service.SystemConfigAdminDeleteInput{Key: req.Key, ActorUserID: actorUserID}); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}
