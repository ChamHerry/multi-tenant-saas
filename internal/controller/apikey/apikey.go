package apikey

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	apiapikey "repomind-temp/api/apikey"
	"repomind-temp/api/apikey/v1"
	"repomind-temp/internal/service"
)

type ControllerV1 struct{}

func NewV1() apiapikey.IAPIKeyV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) CreatePersonal(ctx context.Context, req *v1.CreatePersonalReq) (res *v1.CreateRes, err error) {
	identity, err := requireInteractiveIdentity(ctx)
	if err != nil {
		return nil, err
	}
	expiresAt, err := parseExpiresAt(req.ExpiresAt)
	if err != nil {
		return nil, err
	}
	created, err := service.APIKeyService().CreatePersonal(ctx, service.CreatePersonalAPIKeyInput{
		UserID:    identity.UserID,
		Name:      req.Name,
		Scopes:    req.Scopes,
		Grants:    req.Grants,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{
		Action:       "api_key.create_personal",
		ResourceType: "api_key",
		ResourceID:   created.ID,
		Metadata:     map[string]any{"scopes": req.Scopes, "grant_count": len(req.Grants)},
	})
	return &v1.CreateRes{APIKey: &created.APIKey, RawKey: created.RawKey}, nil
}

func (c *ControllerV1) ListPersonal(ctx context.Context, req *v1.ListPersonalReq) (res *v1.ListRes, err error) {
	identity, err := requireInteractiveIdentity(ctx)
	if err != nil {
		return nil, err
	}
	items, err := service.APIKeyService().ListPersonal(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{APIKeys: items}, nil
}

func (c *ControllerV1) RevokePersonal(ctx context.Context, req *v1.RevokePersonalReq) (res *v1.ActionRes, err error) {
	identity, err := requireInteractiveIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err = service.APIKeyService().RevokePersonal(ctx, identity.UserID, req.APIKey); err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{Action: "api_key.revoke_personal", ResourceType: "api_key", ResourceID: req.APIKey})
	return &v1.ActionRes{OK: true}, nil
}

func parseExpiresAt(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func requireInteractiveIdentity(ctx context.Context) (*service.AuthIdentity, error) {
	identity, err := service.MustAuthIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if identity.Type == "api_key" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api keys cannot manage personal api keys")
	}
	return identity, nil
}
