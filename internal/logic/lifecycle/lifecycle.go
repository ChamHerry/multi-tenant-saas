package lifecycle

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
	"repomind-temp/utility/uuid"
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
	rows, err := g.DB().GetAll(ctx, `
SELECT id, tenant_id, type
FROM public.tenant_lifecycle_jobs
WHERE status='pending' AND scheduled_at <= now()
ORDER BY scheduled_at ASC
LIMIT ?`, limit)
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
	result, err := g.DB().Exec(ctx, `
UPDATE public.tenant_lifecycle_jobs
SET status='cancelled', updated_at=now(), finished_at=now()
WHERE id=? AND status='pending'`, jobID)
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
	record, err := g.DB().GetOne(ctx, `
INSERT INTO public.tenant_lifecycle_jobs(id, tenant_id, type, status, requested_by_user_id, scheduled_at, metadata, created_at, updated_at)
VALUES (?, ?, ?, 'pending', NULLIF(?, '')::uuid, ?, ?::jsonb, now(), now())
RETURNING id, tenant_id, type, status, requested_by_user_id, scheduled_at, started_at, finished_at,
          error_message, artifact_uri, metadata, created_at, updated_at`, jobID, tenantID, jobType, actorUserID, scheduledAt, string(payload))
	if err != nil {
		return nil, gerror.Wrap(err, "insert tenant lifecycle job")
	}
	job, err := mapJob(record)
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.lifecycle." + jobType + ".request", ResourceType: "tenant_lifecycle_job", ResourceID: jobID})
	return job, nil
}

func (s *sTenantLifecycle) runOne(ctx context.Context, jobID, tenantID, jobType string) error {
	_, err := g.DB().Exec(ctx, `UPDATE public.tenant_lifecycle_jobs SET status='running', started_at=now(), updated_at=now() WHERE id=? AND status='pending'`, jobID)
	if err != nil {
		return gerror.Wrap(err, "mark lifecycle job running")
	}
	var runErr error
	switch jobType {
	case "purge":
		runErr = purgeTenant(ctx, tenantID)
	case "export", "quota_recalculate":
		runErr = nil
	default:
		runErr = gerror.NewCodef(gcode.CodeInvalidParameter, "unknown lifecycle job type %s", jobType)
	}
	if runErr != nil {
		_, _ = g.DB().Exec(ctx, `UPDATE public.tenant_lifecycle_jobs SET status='failed', error_message=?, finished_at=now(), updated_at=now() WHERE id=?`, runErr.Error(), jobID)
		return runErr
	}
	_, err = g.DB().Exec(ctx, `UPDATE public.tenant_lifecycle_jobs SET status='succeeded', finished_at=now(), updated_at=now() WHERE id=?`, jobID)
	return gerror.Wrap(err, "mark lifecycle job succeeded")
}

func purgeTenant(ctx context.Context, tenantID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	record, err := g.DB().GetOne(ctx, `SELECT id FROM public.tenants WHERE id=? AND status='deleted'`, tenantID)
	if err != nil {
		return gerror.Wrap(err, "select tenant for purge")
	}
	if record.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "tenant is not in deleted status")
	}
	_, err = g.DB().Exec(ctx, `
UPDATE public.tenants
SET updated_at=now()
WHERE id=? AND status='deleted'`, tenantID)
	return gerror.Wrap(err, "mark public-schema tenant purge complete")
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
