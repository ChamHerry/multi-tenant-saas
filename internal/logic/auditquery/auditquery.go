package auditquery

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type sAuditQuery struct{}

func init() {
	service.RegisterAuditQuery(&sAuditQuery{})
}

func (s *sAuditQuery) List(ctx context.Context, filter service.AuditLogFilter) (*service.AuditLogList, error) {
	where, args, err := buildWhere(filter)
	if err != nil {
		return nil, err
	}
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL := strings.Join(where, " AND ")
	total, err := dao.AuditLogs.Ctx(ctx).Where(whereSQL, args...).Count()
	if err != nil {
		return nil, gerror.Wrap(err, "count audit logs")
	}
	rows, err := dao.AuditLogs.Ctx(ctx).
		Where(whereSQL, args...).
		OrderDesc(dao.AuditLogs.Columns().CreatedAt).
		Limit(limit).
		Offset(offset).
		All()
	if err != nil {
		return nil, gerror.Wrap(err, "list audit logs")
	}
	items := make([]service.AuditLog, 0, len(rows))
	for _, row := range rows {
		item, err := mapAuditLog(row)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return &service.AuditLogList{Logs: items, Total: total}, nil
}

func (s *sAuditQuery) Export(ctx context.Context, filter service.AuditLogFilter) (*service.AuditExportJob, error) {
	where, args, err := buildWhere(filter)
	if err != nil {
		return nil, err
	}
	whereSQL := strings.Join(where, " AND ")
	rows, err := dao.AuditLogs.Ctx(ctx).Where(whereSQL, args...).All()
	if err != nil {
		return nil, gerror.Wrap(err, "export audit logs")
	}
	var buf bytes.Buffer
	for _, row := range rows {
		item, err := mapAuditLog(row)
		if err != nil {
			return nil, err
		}
		line, err := json.Marshal(item)
		if err != nil {
			return nil, gerror.Wrap(err, "marshal audit export line")
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}
	jobID := uuid.GenerateV4()
	fileName := "audit_logs_" + jobID + ".jsonl"
	return &service.AuditExportJob{ID: jobID, Status: "succeeded", Format: "jsonl", ContentType: "application/x-ndjson", FileName: fileName, Content: buf.String()}, nil
}

func buildWhere(filter service.AuditLogFilter) ([]string, []any, error) {
	where := []string{"1=1"}
	args := []any{}
	if filter.TenantID != "" {
		if err := validateInternalID(filter.TenantID); err != nil {
			return nil, nil, err
		}
		where = append(where, "tenant_id=?")
		args = append(args, filter.TenantID)
	}
	if filter.UserID != "" {
		if err := validateInternalID(filter.UserID); err != nil {
			return nil, nil, err
		}
		where = append(where, "user_id=?")
		args = append(args, filter.UserID)
	}
	if filter.Action != "" {
		where = append(where, "action=?")
		args = append(args, filter.Action)
	}
	if filter.ResourceType != "" {
		where = append(where, "resource_type=?")
		args = append(args, filter.ResourceType)
	}
	if filter.ResourceID != "" {
		where = append(where, "resource_id=?")
		args = append(args, filter.ResourceID)
	}
	if filter.From != nil {
		where = append(where, "created_at>=?")
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		where = append(where, "created_at<=?")
		args = append(args, *filter.To)
	}
	return where, args, nil
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "invalid internal uuid %q", id)
	}
	return nil
}

func normalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func mapAuditLog(record gdb.Record) (*service.AuditLog, error) {
	metadata := map[string]any{}
	if raw := record["metadata"].String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &metadata)
	}
	redactMetadata(metadata)
	return &service.AuditLog{
		ID:           record["id"].String(),
		TenantID:     record["tenant_id"].String(),
		UserID:       record["user_id"].String(),
		Action:       record["action"].String(),
		ResourceType: record["resource_type"].String(),
		ResourceID:   record["resource_id"].String(),
		IP:           record["ip"].String(),
		UserAgent:    record["user_agent"].String(),
		Metadata:     metadata,
		CreatedAt:    record["created_at"].Time(),
	}, nil
}

func redactMetadata(metadata map[string]any) {
	for key := range metadata {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "key") {
			metadata[key] = "[REDACTED]"
		}
	}
}
