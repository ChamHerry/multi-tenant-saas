package audit

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/do"
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

	// Auto-fill IP, UserAgent and trace metadata from BizContext.
	bc := service.BizCtx().Get(ctx)
	if in.IP == "" {
		in.IP = bc.ClientIP
	}
	if in.UserAgent == "" {
		in.UserAgent = bc.UserAgent
	}

	metadata := in.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	if bc.RequestID != "" {
		metadata["request_id"] = bc.RequestID
	}
	if bc.TraceID != "" {
		metadata["trace_id"] = bc.TraceID
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
	_, err = dao.AuditLogs.Ctx(ctx).Data(do.AuditLogs{
		TenantId:     nilIfEmpty(in.TenantID),
		UserId:       nilIfEmpty(in.UserID),
		Action:       in.Action,
		ResourceType: in.ResourceType,
		ResourceId:   nilIfEmpty(in.ResourceID),
		Ip:           nilIfEmpty(in.IP),
		UserAgent:    in.UserAgent,
		Metadata:     jsonMetadata,
	}).Insert()
	return gerror.Wrap(err, "insert audit log")
}

func nilIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
