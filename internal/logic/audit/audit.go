package audit

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
)

type sAudit struct{}

func init() {
	service.RegisterAudit(&sAudit{})
}

func (s *sAudit) Write(ctx context.Context, in service.AuditLogInput) error {
	if in.TenantID == "" {
		if tc, ok := service.TenantContextFromCtx(ctx); ok {
			in.TenantID = tc.TenantID
			if in.UserID == "" {
				in.UserID = tc.UserID
			}
		}
	}
	if in.UserID == "" {
		if identity, ok := service.AuthIdentityFromCtx(ctx); ok {
			in.UserID = identity.UserID
		}
	}
	if in.PlatformRole == "" {
		if pac, ok := service.PlatformAdminContextFromCtx(ctx); ok {
			in.PlatformRole = pac.Role
		}
	}
	metadata := in.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	if requestID := service.RequestIDFromCtx(ctx); requestID != "" {
		metadata["request_id"] = requestID
	}
	if in.PlatformRole != "" {
		metadata["platform_role"] = in.PlatformRole
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return gerror.Wrap(err, "marshal audit metadata")
	}
	jsonMetadata, err := gjson.LoadJson(payload)
	if err != nil {
		return gerror.Wrap(err, "parse audit metadata")
	}
	cols := dao.AuditLogs.Columns()
	_, err = dao.AuditLogs.Ctx(ctx).Data(map[string]any{
		cols.TenantId:     nilIfEmpty(in.TenantID),
		cols.UserId:       nilIfEmpty(in.UserID),
		cols.Action:       in.Action,
		cols.ResourceType: in.ResourceType,
		cols.ResourceId:   in.ResourceID,
		cols.Ip:           nilIfEmpty(in.IP),
		cols.UserAgent:    in.UserAgent,
		cols.Metadata:     jsonMetadata,
	}).Insert()
	return gerror.Wrap(err, "insert audit log")
}

func nilIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
