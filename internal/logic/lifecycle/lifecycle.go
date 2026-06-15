package lifecycle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/do"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type sTenantLifecycle struct{}

func init() {
	service.RegisterTenantLifecycle(&sTenantLifecycle{})
}

func (s *sTenantLifecycle) RequestExport(ctx context.Context, tenantID, actorUserID string) (*service.TenantLifecycleJob, error) {
	return s.createJob(ctx, tenantID, actorUserID, "export", time.Now(), nil)
}

func (s *sTenantLifecycle) RequestPurge(ctx context.Context, tenantID, actorUserID string, after time.Time) (*service.TenantLifecycleJob, error) {
	if after.IsZero() {
		after = time.Now()
	}
	return s.createJob(ctx, tenantID, actorUserID, "purge", after, nil)
}

func (s *sTenantLifecycle) RunPendingJobs(ctx context.Context, limit int) error {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cols := dao.TenantLifecycleJobs.Columns()
	rows, err := dao.TenantLifecycleJobs.Ctx(ctx).
		Where(cols.Status, "pending").
		Where(cols.ScheduledAt + " <= NOW()").
		OrderAsc(cols.ScheduledAt).
		Limit(limit).
		All()
	if err != nil {
		return gerror.Wrap(err, "list pending tenant lifecycle jobs")
	}
	for _, row := range rows {
		jobID := row["id"].String()
		jobType := row["type"].String()
		if err := s.runOne(ctx, jobID, row["tenant_id"].String(), jobType); err != nil {
			g.Log().Errorf(ctx, "tenant lifecycle job %s failed: %v", jobID, err)
		}
	}
	return nil
}

func (s *sTenantLifecycle) CancelJob(ctx context.Context, jobID, actorUserID string) error {
	if err := validateInternalID(jobID); err != nil {
		return err
	}
	if actorUserID != "" {
		if err := validateInternalID(actorUserID); err != nil {
			return err
		}
	}
	cols := dao.TenantLifecycleJobs.Columns()
	result, err := dao.TenantLifecycleJobs.Ctx(ctx).
		Where(cols.Id, jobID).
		Where(cols.Status, "pending").
		Data(do.TenantLifecycleJobs{Status: "cancelled"}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "cancel tenant lifecycle job")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.NewCode(gcode.CodeNotFound, "pending lifecycle job not found")
	}
	return service.Audit().Write(ctx, service.AuditLogInput{UserID: actorUserID, Action: "tenant.lifecycle.cancel", ResourceType: "tenant_lifecycle_job", ResourceID: jobID})
}

func (s *sTenantLifecycle) createJob(ctx context.Context, tenantID, actorUserID, jobType string, scheduledAt time.Time, metadata map[string]any) (*service.TenantLifecycleJob, error) {
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	if actorUserID != "" {
		if err := validateInternalID(actorUserID); err != nil {
			return nil, err
		}
	}
	payload, err := json.Marshal(defaultMetadata(metadata))
	if err != nil {
		return nil, gerror.Wrap(err, "marshal lifecycle metadata")
	}
	jobID := uuid.GenerateV4()
	var requestedByVal any
	if actorUserID != "" {
		requestedByVal = actorUserID
	}
	jsonMeta, err := gjson.LoadJson(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "parse lifecycle metadata")
	}
	_, err = dao.TenantLifecycleJobs.Ctx(ctx).Data(do.TenantLifecycleJobs{
		Id:                jobID,
		TenantId:          tenantID,
		Type:              jobType,
		RequestedByUserId: requestedByVal,
		ScheduledAt:       scheduledAt,
		Metadata:          jsonMeta,
	}).Insert()
	if err != nil {
		return nil, gerror.Wrap(err, "insert tenant lifecycle job")
	}
	cols := dao.TenantLifecycleJobs.Columns()
	record, err := dao.TenantLifecycleJobs.Ctx(ctx).Where(cols.Id, jobID).One()
	if err != nil {
		return nil, gerror.Wrap(err, "select created lifecycle job")
	}
	job, err := mapJob(record)
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.lifecycle." + jobType + ".request", ResourceType: "tenant_lifecycle_job", ResourceID: jobID})
	return job, nil
}

func (s *sTenantLifecycle) runOne(ctx context.Context, jobID, tenantID, jobType string) error {
	cols := dao.TenantLifecycleJobs.Columns()
	_, err := dao.TenantLifecycleJobs.Ctx(ctx).
		Where(cols.Id, jobID).
		Where(cols.Status, "pending").
		Data(do.TenantLifecycleJobs{Status: "running", StartedAt: gdb.Raw("NOW()")}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "mark lifecycle job running")
	}
	var runErr error
	var artifactURI string
	switch jobType {
	case "purge":
		runErr = purgeTenant(ctx, tenantID)
	case "export":
		artifactURI, runErr = exportTenant(ctx, tenantID)
	default:
		runErr = gerror.NewCodef(gcode.CodeInvalidParameter, "unknown lifecycle job type %s", jobType)
	}
	if runErr != nil {
		_, _ = dao.TenantLifecycleJobs.Ctx(ctx).
			Where(cols.Id, jobID).
			Data(do.TenantLifecycleJobs{Status: "failed", ErrorMessage: runErr.Error(), FinishedAt: gdb.Raw("NOW()")}).
			Update()
		return runErr
	}
	_, err = dao.TenantLifecycleJobs.Ctx(ctx).
		Where(cols.Id, jobID).
		Data(do.TenantLifecycleJobs{Status: "succeeded", ArtifactUri: artifactURI, FinishedAt: gdb.Raw("NOW()")}).
		Update()
	return gerror.Wrap(err, "mark lifecycle job succeeded")
}

func purgeTenant(ctx context.Context, tenantID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	cols := dao.Tenants.Columns()
	record, err := dao.Tenants.Ctx(ctx).
		Unscoped().
		Where(cols.Id, tenantID).
		Where(cols.Status, "deleted").
		One()
	if err != nil {
		return gerror.Wrap(err, "select tenant for purge")
	}
	if record.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "tenant is not in deleted status")
	}
	_, err = dao.Tenants.Ctx(ctx).
		Unscoped().
		Where(cols.Id, tenantID).
		Where(cols.Status, "deleted").
		Data(do.Tenants{Status: "purged"}).
		Update()
	return gerror.Wrap(err, "mark public-schema tenant purge complete")
}

// exportTenant queries core tenant data and writes it to a JSONL file.
// Returns the artifact URI (file path) on success.
func exportTenant(ctx context.Context, tenantID string) (string, error) {
	if err := validateInternalID(tenantID); err != nil {
		return "", err
	}
	exportDir := service.Config().GetString(ctx, "tenant.lifecycle.exportDir", "/tmp/tenant-exports")
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return "", gerror.Wrap(err, "create export directory")
	}
	fileName := fmt.Sprintf("%s_%s.jsonl", tenantID, time.Now().Format("20060102T150405"))
	filePath := filepath.Join(exportDir, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		return "", gerror.Wrap(err, "create export file")
	}
	defer f.Close()
	// Per-table DAO queries
	daoQueries := []struct {
		label string
		query func(ctx context.Context, tid string) ([]gdb.Record, error)
	}{
		{"tenants", func(ctx context.Context, tid string) ([]gdb.Record, error) {
			return dao.Tenants.Ctx(ctx).Where(dao.Tenants.Columns().Id, tid).All()
		}},
		{"tenant_memberships", func(ctx context.Context, tid string) ([]gdb.Record, error) {
			return dao.TenantMemberships.Ctx(ctx).Where(dao.TenantMemberships.Columns().TenantId, tid).All()
		}},
		{"tenant_invitations", func(ctx context.Context, tid string) ([]gdb.Record, error) {
			return dao.TenantInvitations.Ctx(ctx).Where(dao.TenantInvitations.Columns().TenantId, tid).All()
		}},
		{"api_keys", func(ctx context.Context, tid string) ([]gdb.Record, error) {
			// api_keys.tenant_id is nullable, export by user_id via grants
			return dao.ApiKeys.Ctx(ctx).
				LeftJoin("api_key_tenant_grants akg", "akg.api_key_id = api_keys.id").
				Where("akg.tenant_id", tid).
				All()
		}},
		{"api_key_tenant_grants", func(ctx context.Context, tid string) ([]gdb.Record, error) {
			return dao.ApiKeyTenantGrants.Ctx(ctx).Where(dao.ApiKeyTenantGrants.Columns().TenantId, tid).All()
		}},
	}
	total := 0
	for _, dq := range daoQueries {
		rows, err := dq.query(ctx, tenantID)
		if err != nil {
			return "", gerror.Wrapf(err, "export table %s", dq.label)
		}
		for _, row := range rows {
			line, err := json.Marshal(row.Map())
			if err != nil {
				continue
			}
			if _, err := f.Write(append(line, '\n')); err != nil {
				return "", gerror.Wrap(err, "write export line")
			}
			total++
		}
	}
	g.Log().Infof(ctx, "[lifecycle] exported tenant %s: %d rows -> %s", tenantID, total, filePath)
	return filePath, nil
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "invalid internal uuid %q", id)
	}
	return nil
}

func mapJob(record gdb.Record) (*service.TenantLifecycleJob, error) {
	metadata := map[string]any{}
	if raw := record["metadata"].String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &metadata)
	}
	return &service.TenantLifecycleJob{
		ID:                record["id"].String(),
		TenantID:          record["tenant_id"].String(),
		Type:              record["type"].String(),
		Status:            record["status"].String(),
		RequestedByUserID: record["requested_by_user_id"].String(),
		ScheduledAt:       record["scheduled_at"].Time(),
		StartedAt:         nullableTime(record["started_at"]),
		FinishedAt:        nullableTime(record["finished_at"]),
		ErrorMessage:      record["error_message"].String(),
		ArtifactURI:       record["artifact_uri"].String(),
		Metadata:          metadata,
		CreatedAt:         record["created_at"].Time(),
		UpdatedAt:         record["updated_at"].Time(),
	}, nil
}

func nullableTime(value any) *time.Time {
	v, ok := value.(interface {
		IsNil() bool
		Time(...string) time.Time
	})
	if !ok || v.IsNil() {
		return nil
	}
	t := v.Time()
	return &t
}

func defaultMetadata(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	return in
}
