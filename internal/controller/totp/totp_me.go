package totp

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	totpapi "multi-tenant-saas/api/totp"
	"multi-tenant-saas/api/totp/v1"
	"multi-tenant-saas/internal/service"
)

type MeTotpControllerV1 struct{}

func NewMeTotpV1() totpapi.IMeTotpV1 {
	return &MeTotpControllerV1{}
}

func (c *MeTotpControllerV1) Setup(ctx context.Context, req *v1.SetupReq) (res *v1.SetupRes, err error) {
	identity, err := requireInteractiveSecurityIdentity(ctx)
	if err != nil {
		return nil, err
	}
	setup, err := service.TOTP().Setup(ctx, identity.UserID, req.Password)
	if err != nil {
		return nil, err
	}
	return &v1.SetupRes{Secret: setup.Secret, URL: setup.URL}, nil
}

func (c *MeTotpControllerV1) Enable(ctx context.Context, req *v1.EnableReq) (res *v1.EnableRes, err error) {
	identity, err := requireInteractiveSecurityIdentity(ctx)
	if err != nil {
		return nil, err
	}
	codes, err := service.TOTP().Enable(ctx, identity.UserID, req.Code)
	if err != nil {
		return nil, err
	}
	return &v1.EnableRes{OK: true, BackupCodes: codes}, nil
}

func (c *MeTotpControllerV1) Disable(ctx context.Context, req *v1.DisableReq) (res *v1.ActionRes, err error) {
	identity, err := requireInteractiveSecurityIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.TOTP().Disable(ctx, identity.UserID, req.Password, req.Code); err != nil {
		return nil, err
	}
	return &v1.ActionRes{OK: true}, nil
}

func (c *MeTotpControllerV1) RegenerateBackupCodes(ctx context.Context, req *v1.RegenerateBackupCodesReq) (res *v1.RegenerateBackupCodesRes, err error) {
	identity, err := requireInteractiveSecurityIdentity(ctx)
	if err != nil {
		return nil, err
	}
	codes, err := service.TOTP().RegenerateBackupCodes(ctx, identity.UserID, req.Password)
	if err != nil {
		return nil, err
	}
	return &v1.RegenerateBackupCodesRes{BackupCodes: codes}, nil
}

func (c *MeTotpControllerV1) Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error) {
	identity, err := requireSecurityReadIdentity(ctx)
	if err != nil {
		return nil, err
	}
	status, err := service.TOTP().GetStatus(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.StatusRes{
		Enabled:              status.Enabled,
		SetupInitiated:       status.SetupInitiated,
		EnabledAt:            status.EnabledAt,
		BackupCodesRemaining: status.BackupCodesRemaining,
		CreatedAt:            status.CreatedAt,
	}, nil
}

func requireSecurityReadIdentity(ctx context.Context) (*service.AuthIdentity, error) {
	if err := service.RBAC().RequireAuthScope(ctx, service.PermissionUserSecurityRead); err != nil {
		return nil, err
	}
	return service.MustAuthIdentity(ctx)
}

func requireInteractiveSecurityIdentity(ctx context.Context) (*service.AuthIdentity, error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if identity.Type != "session" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "interactive session is required to manage TOTP settings")
	}
	return identity, nil
}
