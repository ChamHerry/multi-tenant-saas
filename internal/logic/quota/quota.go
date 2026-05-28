package quota

import (
	"context"
	"regexp"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
	"repomind-temp/utility/uuid"
)

var (
	internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	codeQuotaExceeded = gcode.New(429001, "QUOTA_EXCEEDED", nil)
)

type sQuota struct{}

func init() {
	service.RegisterQuota(&sQuota{})
}

func (s *sQuota) GetTenantQuota(ctx context.Context, tenantID string) (*service.TenantQuotaView, error) {
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	plan, err := tenantPlan(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	metrics := managedMetrics()
	items := make([]service.QuotaUsage, 0, len(metrics))
	for _, metric := range metrics {
		used, err := currentUsage(ctx, tenantID, metric)
		if err != nil {
			return nil, err
		}
		reserved, _ := currentReserved(ctx, tenantID, metric)
		limit, err := metricLimit(ctx, tenantID, plan, metric)
		if err != nil {
			return nil, err
		}
		var remaining *int64
		if limit != nil {
			value := *limit - used - reserved
			if value < 0 {
				value = 0
			}
			remaining = &value
		}
		items = append(items, service.QuotaUsage{Metric: metric, Used: used, Reserved: reserved, Limit: limit, Remaining: remaining})
	}
	return &service.TenantQuotaView{TenantID: tenantID, Plan: plan, Items: items}, nil
}

func (s *sQuota) Require(ctx context.Context, tenantID string, metric service.QuotaMetric, delta int64) error {
	if delta <= 0 {
		return nil
	}
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	plan, err := tenantPlan(ctx, tenantID)
	if err != nil {
		return err
	}
	limit, err := metricLimit(ctx, tenantID, plan, metric)
	if err != nil || limit == nil {
		return err
	}
	used, err := currentUsage(ctx, tenantID, metric)
	if err != nil {
		return err
	}
	reserved, _ := currentReserved(ctx, tenantID, metric)
	if used+reserved+delta > *limit {
		return quotaExceeded(metric, *limit, used, reserved, delta)
	}
	return nil
}

func (s *sQuota) RequireAndReserve(ctx context.Context, in service.QuotaReservationInput) (*service.QuotaReservation, error) {
	if err := validateReservationInput(in); err != nil {
		return nil, err
	}
	var reservation *service.QuotaReservation
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		plan, err := tenantPlanTx(ctx, tx, in.TenantID)
		if err != nil {
			return err
		}
		limit, err := metricLimitTx(ctx, tx, in.TenantID, plan, in.Metric)
		if err != nil {
			return err
		}
		used, err := currentUsageTx(ctx, tx, in.TenantID, in.Metric)
		if err != nil {
			return err
		}
		counter, err := lockCounter(ctx, tx, in.TenantID, in.Metric)
		if err != nil {
			return err
		}
		reserved := counter["reserved"].Int64()
		if limit != nil && used+reserved+in.Delta > *limit {
			return quotaExceeded(in.Metric, *limit, used, reserved, in.Delta)
		}
		reservationID := uuid.GenerateV4()
		record, err := tx.Ctx(ctx).GetOne(`
INSERT INTO public.tenant_usage_reservations(id, tenant_id, metric, delta, status, resource_type, resource_id, expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, 'reserved', ?, ?, ?, now(), now())
RETURNING id, tenant_id, metric, delta, status, resource_type, resource_id, expires_at, created_at, updated_at`,
			reservationID, in.TenantID, string(in.Metric), in.Delta, in.ResourceType, in.ResourceID, in.ExpiresAt)
		if err != nil {
			return gerror.Wrap(err, "insert quota reservation")
		}
		_, err = tx.Ctx(ctx).Exec(`
UPDATE public.tenant_usage_counters
SET reserved=reserved+?, limit_snapshot=?, updated_at=now()
WHERE tenant_id=? AND metric=? AND period_start='1970-01-01 00:00:00+00'`, in.Delta, limitValue(limit), in.TenantID, string(in.Metric))
		if err != nil {
			return gerror.Wrap(err, "update quota reserved")
		}
		reservation = mapReservation(record)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

func (s *sQuota) CommitReservation(ctx context.Context, reservationID string) error {
	return finishReservation(ctx, reservationID, "committed")
}

func (s *sQuota) ReleaseReservation(ctx context.Context, reservationID string) error {
	return finishReservation(ctx, reservationID, "released")
}

func (s *sQuota) Recalculate(ctx context.Context, tenantID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	metrics := managedMetrics()
	for _, metric := range metrics {
		used, err := currentUsage(ctx, tenantID, metric)
		if err != nil {
			return err
		}
		_, err = g.DB().Exec(ctx, `
INSERT INTO public.tenant_usage_counters(tenant_id, metric, used, reserved, updated_at)
VALUES (?, ?, ?, 0, now())
ON CONFLICT (tenant_id, metric, period_start) DO UPDATE
SET used=EXCLUDED.used, updated_at=now()`, tenantID, string(metric), used)
		if err != nil {
			return gerror.Wrapf(err, "recalculate quota metric %s", metric)
		}
	}
	return service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, Action: "quota.recalculate", ResourceType: "tenant", ResourceID: tenantID})
}

func finishReservation(ctx context.Context, reservationID string, status string) error {
	if err := validateInternalID(reservationID); err != nil {
		return err
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record, err := tx.Ctx(ctx).GetOne(`
SELECT id, tenant_id, metric, delta, status
FROM public.tenant_usage_reservations
WHERE id=?
FOR UPDATE`, reservationID)
		if err != nil {
			return gerror.Wrap(err, "select quota reservation")
		}
		if record.IsEmpty() {
			return gerror.NewCode(gcode.CodeNotFound, "quota reservation not found")
		}
		if record["status"].String() != "reserved" {
			return nil
		}
		delta := record["delta"].Int64()
		tenantID := record["tenant_id"].String()
		metric := record["metric"].String()
		if status == "committed" {
			_, err = tx.Ctx(ctx).Exec(`
UPDATE public.tenant_usage_counters
SET reserved=GREATEST(0, reserved-?), used=used+?, updated_at=now()
WHERE tenant_id=? AND metric=? AND period_start='1970-01-01 00:00:00+00'`, delta, delta, tenantID, metric)
		} else {
			_, err = tx.Ctx(ctx).Exec(`
UPDATE public.tenant_usage_counters
SET reserved=GREATEST(0, reserved-?), updated_at=now()
WHERE tenant_id=? AND metric=? AND period_start='1970-01-01 00:00:00+00'`, delta, tenantID, metric)
		}
		if err != nil {
			return gerror.Wrap(err, "finish quota counter")
		}
		_, err = tx.Ctx(ctx).Exec(`UPDATE public.tenant_usage_reservations SET status=?, updated_at=now() WHERE id=?`, status, reservationID)
		return gerror.Wrap(err, "finish quota reservation")
	})
}

func validateReservationInput(in service.QuotaReservationInput) error {
	if err := validateInternalID(in.TenantID); err != nil {
		return err
	}
	if in.Delta <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "quota reservation delta must be positive")
	}
	if in.Metric == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "quota metric is required")
	}
	if in.ResourceType == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "quota resource type is required")
	}
	return nil
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "invalid internal uuid %q", id)
	}
	return nil
}

func tenantPlan(ctx context.Context, tenantID string) (string, error) {
	record, err := g.DB().GetOne(ctx, `SELECT plan FROM public.tenants WHERE id=? AND deleted_at IS NULL`, tenantID)
	if err != nil {
		return "", gerror.Wrap(err, "select tenant plan")
	}
	if record.IsEmpty() {
		return "", gerror.NewCode(gcode.CodeNotFound, "tenant not found")
	}
	return record["plan"].String(), nil
}

func tenantPlanTx(ctx context.Context, tx gdb.TX, tenantID string) (string, error) {
	record, err := tx.Ctx(ctx).GetOne(`SELECT plan FROM public.tenants WHERE id=? AND deleted_at IS NULL`, tenantID)
	if err != nil {
		return "", gerror.Wrap(err, "select tenant plan")
	}
	if record.IsEmpty() {
		return "", gerror.NewCode(gcode.CodeNotFound, "tenant not found")
	}
	return record["plan"].String(), nil
}

func managedMetrics() []service.QuotaMetric {
	return []service.QuotaMetric{service.MetricAPIKeyCount, service.MetricMemberCount}
}

func metricLimit(ctx context.Context, tenantID, plan string, metric service.QuotaMetric) (*int64, error) {
	column, err := quotaOverrideColumn(metric)
	if err != nil {
		return nil, err
	}
	if column != "" {
		record, err := g.DB().GetOne(ctx, "SELECT "+column+" AS limit_value FROM public.tenant_quotas WHERE tenant_id=?", tenantID)
		if err != nil {
			return nil, gerror.Wrap(err, "select tenant quota")
		}
		if !record.IsEmpty() && !record["limit_value"].IsNil() {
			value := record["limit_value"].Int64()
			return &value, nil
		}
	}
	return planMetricLimit(ctx, plan, metric)
}

func metricLimitTx(ctx context.Context, tx gdb.TX, tenantID, plan string, metric service.QuotaMetric) (*int64, error) {
	column, err := quotaOverrideColumn(metric)
	if err != nil {
		return nil, err
	}
	if column != "" {
		record, err := tx.Ctx(ctx).GetOne("SELECT "+column+" AS limit_value FROM public.tenant_quotas WHERE tenant_id=?", tenantID)
		if err != nil {
			return nil, gerror.Wrap(err, "select tenant quota")
		}
		if !record.IsEmpty() && !record["limit_value"].IsNil() {
			value := record["limit_value"].Int64()
			return &value, nil
		}
	}
	feature := featureKey(metric)
	if feature == "" {
		return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "unknown quota metric %s", metric)
	}
	record, err := tx.Ctx(ctx).GetOne(`
SELECT limit_value
FROM public.plan_entitlements
WHERE plan=? AND feature_key=? AND enabled=true`, plan, feature)
	if err != nil {
		return nil, gerror.Wrap(err, "select plan entitlement")
	}
	if record.IsEmpty() || record["limit_value"].IsNil() {
		return nil, nil
	}
	value := record["limit_value"].Int64()
	return &value, nil
}

func planMetricLimit(ctx context.Context, plan string, metric service.QuotaMetric) (*int64, error) {
	feature := featureKey(metric)
	if feature == "" {
		return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "unknown quota metric %s", metric)
	}
	record, err := g.DB().GetOne(ctx, `
SELECT limit_value
FROM public.plan_entitlements
WHERE plan=? AND feature_key=? AND enabled=true`, plan, feature)
	if err != nil {
		return nil, gerror.Wrap(err, "select plan entitlement")
	}
	if record.IsEmpty() || record["limit_value"].IsNil() {
		return nil, nil
	}
	value := record["limit_value"].Int64()
	return &value, nil
}

func featureKey(metric service.QuotaMetric) string {
	switch metric {
	case service.MetricAPIKeyCount:
		return "api_key.max_count"
	case service.MetricMemberCount:
		return "member.max_count"
	default:
		return ""
	}
}

func quotaOverrideColumn(metric service.QuotaMetric) (string, error) {
	switch metric {
	case service.MetricAPIKeyCount:
		return "max_api_keys", nil
	case service.MetricMemberCount:
		return "max_members", nil
	default:
		return "", gerror.NewCodef(gcode.CodeInvalidParameter, "unknown quota metric %s", metric)
	}
}

func currentUsage(ctx context.Context, tenantID string, metric service.QuotaMetric) (int64, error) {
	query, err := usageQuery(metric)
	if err != nil {
		return 0, err
	}
	record, err := g.DB().GetOne(ctx, query, tenantID)
	if err != nil {
		return 0, gerror.Wrap(err, "select quota usage")
	}
	if record.IsEmpty() {
		return 0, nil
	}
	return record["used"].Int64(), nil
}

func currentUsageTx(ctx context.Context, tx gdb.TX, tenantID string, metric service.QuotaMetric) (int64, error) {
	query, err := usageQuery(metric)
	if err != nil {
		return 0, err
	}
	record, err := tx.Ctx(ctx).GetOne(query, tenantID)
	if err != nil {
		return 0, gerror.Wrap(err, "select quota usage")
	}
	if record.IsEmpty() {
		return 0, nil
	}
	return record["used"].Int64(), nil
}

func usageQuery(metric service.QuotaMetric) (string, error) {
	switch metric {
	case service.MetricAPIKeyCount:
		return `SELECT count(*) AS used FROM public.api_keys WHERE tenant_id=? AND revoked_at IS NULL`, nil
	case service.MetricMemberCount:
		return `SELECT count(*) AS used FROM public.tenant_memberships WHERE tenant_id=? AND status='active' AND deleted_at IS NULL`, nil
	default:
		return "", gerror.NewCodef(gcode.CodeInvalidParameter, "unknown quota metric %s", metric)
	}
}

func currentReserved(ctx context.Context, tenantID string, metric service.QuotaMetric) (int64, error) {
	record, err := g.DB().GetOne(ctx, `
SELECT COALESCE(sum(delta), 0) AS reserved
FROM public.tenant_usage_reservations
WHERE tenant_id=? AND metric=? AND status='reserved'`, tenantID, string(metric))
	if err != nil {
		return 0, gerror.Wrap(err, "select quota reserved")
	}
	return record["reserved"].Int64(), nil
}

func lockCounter(ctx context.Context, tx gdb.TX, tenantID string, metric service.QuotaMetric) (gdb.Record, error) {
	_, err := tx.Ctx(ctx).Exec(`
INSERT INTO public.tenant_usage_counters(tenant_id, metric, used, reserved, updated_at)
VALUES (?, ?, 0, 0, now())
ON CONFLICT (tenant_id, metric, period_start) DO NOTHING`, tenantID, string(metric))
	if err != nil {
		return nil, gerror.Wrap(err, "ensure quota counter")
	}
	record, err := tx.Ctx(ctx).GetOne(`
SELECT used, reserved
FROM public.tenant_usage_counters
WHERE tenant_id=? AND metric=? AND period_start='1970-01-01 00:00:00+00'
FOR UPDATE`, tenantID, string(metric))
	if err != nil {
		return nil, gerror.Wrap(err, "lock quota counter")
	}
	return record, nil
}

func quotaExceeded(metric service.QuotaMetric, limit, used, reserved, delta int64) error {
	return gerror.NewCodef(codeQuotaExceeded, "QUOTA_EXCEEDED: metric=%s limit=%d used=%d reserved=%d requested=%d", metric, limit, used, reserved, delta)
}

func limitValue(limit *int64) any {
	if limit == nil {
		return nil
	}
	return *limit
}

func mapReservation(record gdb.Record) *service.QuotaReservation {
	return &service.QuotaReservation{
		ID:           record["id"].String(),
		TenantID:     record["tenant_id"].String(),
		Metric:       service.QuotaMetric(record["metric"].String()),
		Delta:        record["delta"].Int64(),
		Status:       record["status"].String(),
		ResourceType: record["resource_type"].String(),
		ResourceID:   record["resource_id"].String(),
		ExpiresAt:    nullableTime(record["expires_at"]),
		CreatedAt:    record["created_at"].Time(),
		UpdatedAt:    record["updated_at"].Time(),
	}
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
