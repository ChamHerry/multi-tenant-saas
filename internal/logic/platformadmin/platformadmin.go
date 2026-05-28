package platformadmin

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

var rolePermissions = map[string][]service.PlatformPermission{
	"super_admin": {
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionTenantManage,
		service.PlatformPermissionUserRead,
		service.PlatformPermissionUserManage,
		service.PlatformPermissionAuditRead,
		service.PlatformPermissionBillingManage,
		service.PlatformPermissionAdminManage,
	},
	"support": {
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionUserRead,
		service.PlatformPermissionAuditRead,
	},
	"billing_admin": {
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionBillingManage,
	},
	"auditor": {
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionAuditRead,
	},
}

type sPlatformAdmin struct{}

func init() {
	service.RegisterPlatformAdmin(&sPlatformAdmin{})
}

func (s *sPlatformAdmin) Resolve(ctx context.Context, userID string) (*service.PlatformAdminContext, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	pac, err := s.resolveActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	if pac == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "active platform admin not found")
	}
	return pac, nil
}

func (s *sPlatformAdmin) ResolveOptional(ctx context.Context, userID string) (*service.PlatformAdminContext, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	return s.resolveActive(ctx, userID)
}

func (s *sPlatformAdmin) PermissionsForRole(role string) []service.PlatformPermission {
	permissions, ok := rolePermissions[role]
	if !ok {
		return []service.PlatformPermission{}
	}
	return append([]service.PlatformPermission(nil), permissions...)
}

func (s *sPlatformAdmin) resolveActive(ctx context.Context, userID string) (*service.PlatformAdminContext, error) {
	record, err := g.DB().GetOne(ctx, `
SELECT user_id, role, status
FROM public.platform_admins
WHERE user_id=? AND status='active'
LIMIT 1`, userID)
	if err != nil {
		return nil, gerror.Wrap(err, "select platform admin")
	}
	if record.IsEmpty() {
		return nil, nil
	}
	role := record["role"].String()
	permissions := s.PermissionsForRole(role)
	if len(permissions) == 0 {
		return nil, gerror.NewCodef(gcode.CodeNotAuthorized, "unknown platform admin role %q", role)
	}
	return &service.PlatformAdminContext{
		UserID:      record["user_id"].String(),
		Role:        role,
		Status:      record["status"].String(),
		Permissions: permissions,
	}, nil
}

func (s *sPlatformAdmin) Require(ctx context.Context, permission service.PlatformPermission) error {
	pac, err := service.MustPlatformAdminContext(ctx)
	if err != nil {
		return err
	}
	for _, allowed := range pac.Permissions {
		if allowed == permission {
			return nil
		}
	}
	return gerror.NewCodef(gcode.CodeNotAuthorized, "platform role %s is not allowed to perform %s", pac.Role, permission)
}

func (s *sPlatformAdmin) ListTenants(ctx context.Context, filter service.PlatformTenantFilter) (*service.PlatformTenantList, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	if status := strings.TrimSpace(filter.Status); status != "" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	if query := strings.TrimSpace(filter.Query); query != "" {
		where = append(where, "(name ILIKE ? OR slug ILIKE ? OR id::text = ?)")
		like := "%" + query + "%"
		args = append(args, like, like, query)
	}
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL := strings.Join(where, " AND ")
	countRecord, err := g.DB().GetOne(ctx, "SELECT count(*) AS total FROM public.tenants WHERE "+whereSQL, args...)
	if err != nil {
		return nil, gerror.Wrap(err, "count platform tenants")
	}
	rowsArgs := append(append([]any{}, args...), limit, offset)
	rows, err := g.DB().GetAll(ctx, `
SELECT id, name, slug, plan, status, created_at, updated_at
FROM public.tenants
WHERE `+whereSQL+`
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, rowsArgs...)
	if err != nil {
		return nil, gerror.Wrap(err, "list platform tenants")
	}
	items := make([]service.Tenant, 0, len(rows))
	for _, row := range rows {
		item, err := mapTenant(row)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return &service.PlatformTenantList{Items: items, Total: countRecord["total"].Int()}, nil
}

func (s *sPlatformAdmin) ListUsers(ctx context.Context, filter service.PlatformUserFilter) (*service.PlatformUserList, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	if status := strings.TrimSpace(filter.Status); status != "" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	if query := strings.TrimSpace(filter.Query); query != "" {
		where = append(where, "(email ILIKE ? OR display_name ILIKE ? OR id::text = ?)")
		like := "%" + query + "%"
		args = append(args, like, like, query)
	}
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL := strings.Join(where, " AND ")
	countRecord, err := g.DB().GetOne(ctx, "SELECT count(*) AS total FROM public.users WHERE "+whereSQL, args...)
	if err != nil {
		return nil, gerror.Wrap(err, "count platform users")
	}
	rowsArgs := append(append([]any{}, args...), limit, offset)
	rows, err := g.DB().GetAll(ctx, `
SELECT id, email, display_name, avatar_url, status, last_login_at, metadata, created_at, updated_at
FROM public.users
WHERE `+whereSQL+`
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, rowsArgs...)
	if err != nil {
		return nil, gerror.Wrap(err, "list platform users")
	}
	items := make([]service.User, 0, len(rows))
	for _, row := range rows {
		item, err := mapUser(row)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return &service.PlatformUserList{Items: items, Total: countRecord["total"].Int()}, nil
}

func (s *sPlatformAdmin) UpdateUserStatus(ctx context.Context, actorUserID, targetUserID, status string) error {
	if err := validateInternalID(actorUserID); err != nil {
		return err
	}
	if err := validateInternalID(targetUserID); err != nil {
		return err
	}
	status = strings.TrimSpace(status)
	if status != "active" && status != "disabled" {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "invalid user status %q", status)
	}
	result, err := g.DB().Exec(ctx, `
UPDATE public.users
SET status=?, updated_at=now()
WHERE id=? AND deleted_at IS NULL`, status, targetUserID)
	if err != nil {
		return gerror.Wrap(err, "update platform user status")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.NewCode(gcode.CodeNotFound, "user not found")
	}
	return service.Audit().Write(ctx, service.AuditLogInput{UserID: actorUserID, Action: "platform_user.status.update", ResourceType: "user", ResourceID: targetUserID, Metadata: map[string]any{"status": status}})
}

func (s *sPlatformAdmin) ListPlatformAdmins(ctx context.Context, filter service.PlatformAdminFilter) (*service.PlatformAdminList, error) {
	where := []string{"1=1"}
	args := []any{}
	if status := strings.TrimSpace(filter.Status); status != "" {
		where = append(where, "pa.status=?")
		args = append(args, status)
	}
	if role := strings.TrimSpace(filter.Role); role != "" {
		where = append(where, "pa.role=?")
		args = append(args, role)
	}
	if query := strings.TrimSpace(filter.Query); query != "" {
		where = append(where, "(u.email ILIKE ? OR u.display_name ILIKE ? OR pa.user_id::text = ?)")
		like := "%" + query + "%"
		args = append(args, like, like, query)
	}
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL := strings.Join(where, " AND ")
	countRecord, err := g.DB().GetOne(ctx, `
SELECT count(*) AS total
FROM public.platform_admins pa
JOIN public.users u ON u.id=pa.user_id
WHERE `+whereSQL, args...)
	if err != nil {
		return nil, gerror.Wrap(err, "count platform admins")
	}
	rowsArgs := append(append([]any{}, args...), limit, offset)
	rows, err := g.DB().GetAll(ctx, `
SELECT pa.user_id, pa.role, pa.status, pa.created_by_user_id, pa.created_at, pa.updated_at
FROM public.platform_admins pa
JOIN public.users u ON u.id=pa.user_id
WHERE `+whereSQL+`
ORDER BY pa.created_at DESC
LIMIT ? OFFSET ?`, rowsArgs...)
	if err != nil {
		return nil, gerror.Wrap(err, "list platform admins")
	}
	items := make([]service.PlatformAdmin, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPlatformAdmin(row))
	}
	return &service.PlatformAdminList{Items: items, Total: countRecord["total"].Int()}, nil
}

func (s *sPlatformAdmin) Grant(ctx context.Context, actorUserID, targetUserID, role string) error {
	if err := validateInternalID(actorUserID); err != nil {
		return err
	}
	if err := validateInternalID(targetUserID); err != nil {
		return err
	}
	if _, ok := rolePermissions[role]; !ok {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "invalid platform admin role %q", role)
	}
	_, err := g.DB().Exec(ctx, `
INSERT INTO public.platform_admins(user_id, role, status, created_by_user_id, created_at, updated_at)
VALUES (?, ?, 'active', ?, now(), now())
ON CONFLICT (user_id) DO UPDATE
SET role=EXCLUDED.role, status='active', updated_at=now()`, targetUserID, role, actorUserID)
	if err != nil {
		return gerror.Wrap(err, "grant platform admin")
	}
	return service.Audit().Write(ctx, service.AuditLogInput{UserID: actorUserID, Action: "platform_admin.grant", ResourceType: "platform_admin", ResourceID: targetUserID, Metadata: map[string]any{"role": role}})
}

func (s *sPlatformAdmin) Revoke(ctx context.Context, actorUserID, targetUserID string) error {
	if err := validateInternalID(actorUserID); err != nil {
		return err
	}
	if err := validateInternalID(targetUserID); err != nil {
		return err
	}
	result, err := g.DB().Exec(ctx, `UPDATE public.platform_admins SET status='suspended', updated_at=now() WHERE user_id=?`, targetUserID)
	if err != nil {
		return gerror.Wrap(err, "revoke platform admin")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.NewCode(gcode.CodeNotFound, "platform admin not found")
	}
	return service.Audit().Write(ctx, service.AuditLogInput{UserID: actorUserID, Action: "platform_admin.revoke", ResourceType: "platform_admin", ResourceID: targetUserID})
}

func (s *sPlatformAdmin) ListPlans(ctx context.Context, filter service.PlanEntitlementFilter) (*service.PlanEntitlementList, error) {
	where := []string{"1=1"}
	args := []any{}
	if plan := strings.TrimSpace(filter.Plan); plan != "" {
		where = append(where, "plan=?")
		args = append(args, plan)
	}
	if feature := strings.TrimSpace(filter.FeatureKey); feature != "" {
		where = append(where, "feature_key=?")
		args = append(args, feature)
	}
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL := strings.Join(where, " AND ")
	countRecord, err := g.DB().GetOne(ctx, "SELECT count(*) AS total FROM public.plan_entitlements WHERE "+whereSQL, args...)
	if err != nil {
		return nil, gerror.Wrap(err, "count plan entitlements")
	}
	rowsArgs := append(append([]any{}, args...), limit, offset)
	rows, err := g.DB().GetAll(ctx, `
SELECT plan, feature_key, enabled, limit_value, metadata, created_at, updated_at
FROM public.plan_entitlements
WHERE `+whereSQL+`
ORDER BY plan ASC, feature_key ASC
LIMIT ? OFFSET ?`, rowsArgs...)
	if err != nil {
		return nil, gerror.Wrap(err, "list plan entitlements")
	}
	items := make([]service.PlanEntitlement, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPlanEntitlement(row))
	}
	return &service.PlanEntitlementList{Items: items, Total: countRecord["total"].Int()}, nil
}

func (s *sPlatformAdmin) UpdateTenantPlan(ctx context.Context, actorUserID, tenantID, plan string) error {
	if err := validateInternalID(actorUserID); err != nil {
		return err
	}
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	plan = strings.TrimSpace(plan)
	if plan == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "plan is required")
	}
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		exists, err := tx.Ctx(ctx).GetOne(`SELECT 1 FROM public.plan_entitlements WHERE plan=? LIMIT 1`, plan)
		if err != nil {
			return gerror.Wrap(err, "select plan entitlement")
		}
		if exists.IsEmpty() {
			return gerror.NewCodef(gcode.CodeInvalidParameter, "unknown plan %q", plan)
		}
		result, err := tx.Ctx(ctx).Exec(`
UPDATE public.tenants
SET plan=?, updated_at=now()
WHERE id=? AND deleted_at IS NULL`, plan, tenantID)
		if err != nil {
			return gerror.Wrap(err, "update tenant plan")
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return gerror.NewCode(gcode.CodeNotFound, "tenant not found")
		}
		subResult, err := tx.Ctx(ctx).Exec(`UPDATE public.subscriptions SET plan=?, started_at=now(), ends_at=NULL WHERE tenant_id=? AND status='active'`, plan, tenantID)
		if err != nil {
			return gerror.Wrap(err, "update active subscription")
		}
		subRows, _ := subResult.RowsAffected()
		if subRows == 0 {
			_, err = tx.Ctx(ctx).Exec(`INSERT INTO public.subscriptions(tenant_id, plan, status, started_at) VALUES (?, ?, 'active', now())`, tenantID, plan)
			if err != nil {
				return gerror.Wrap(err, "insert active subscription")
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.plan.update", ResourceType: "tenant", ResourceID: tenantID, Metadata: map[string]any{"plan": plan}})
}

func (s *sPlatformAdmin) UpdateTenantQuota(ctx context.Context, actorUserID, tenantID string, in service.UpdateTenantQuotaInput) error {
	if err := validateInternalID(actorUserID); err != nil {
		return err
	}
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if in.MaxMembers <= 0 {
		return gerror.NewCode(gcode.CodeMissingParameter, "at least one positive quota field is required")
	}
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		exists, err := tx.Ctx(ctx).GetOne(`SELECT 1 FROM public.tenants WHERE id=? AND deleted_at IS NULL`, tenantID)
		if err != nil {
			return gerror.Wrap(err, "select tenant")
		}
		if exists.IsEmpty() {
			return gerror.NewCode(gcode.CodeNotFound, "tenant not found")
		}
		_, err = tx.Ctx(ctx).Exec(`
INSERT INTO public.tenant_quotas(tenant_id, max_members)
VALUES (
    ?,
    CAST(CASE WHEN ? > 0 THEN ? ELSE NULL END AS INT)
)
ON CONFLICT (tenant_id) DO UPDATE
SET max_members=CAST(CASE WHEN ? > 0 THEN ? ELSE public.tenant_quotas.max_members END AS INT),
    updated_at=now()`,
			tenantID,
			in.MaxMembers, in.MaxMembers,
			in.MaxMembers, in.MaxMembers)
		return gerror.Wrap(err, "upsert tenant quota row")
	})
	if err != nil {
		return err
	}
	return service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.quota.update", ResourceType: "tenant", ResourceID: tenantID, Metadata: map[string]any{
		"max_members": in.MaxMembers,
	}})
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "invalid internal uuid %q", id)
	}
	return nil
}

func normalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func mapTenant(record gdb.Record) (*service.Tenant, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	return &service.Tenant{
		ID:        id,
		Name:      record["name"].String(),
		Slug:      record["slug"].String(),
		Plan:      record["plan"].String(),
		Status:    record["status"].String(),
		CreatedAt: record["created_at"].Time(),
		UpdatedAt: record["updated_at"].Time(),
	}, nil
}

func mapUser(record gdb.Record) (*service.User, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	return &service.User{
		ID:          id,
		Email:       record["email"].String(),
		DisplayName: record["display_name"].String(),
		AvatarURL:   record["avatar_url"].String(),
		Status:      record["status"].String(),
		LastLoginAt: nullableTime(record["last_login_at"]),
		Metadata:    map[string]any{},
		CreatedAt:   record["created_at"].Time(),
		UpdatedAt:   record["updated_at"].Time(),
	}, nil
}

func mapPlatformAdmin(record gdb.Record) service.PlatformAdmin {
	return service.PlatformAdmin{
		UserID:          record["user_id"].String(),
		Role:            record["role"].String(),
		Status:          record["status"].String(),
		CreatedByUserID: record["created_by_user_id"].String(),
		CreatedAt:       record["created_at"].Time(),
		UpdatedAt:       record["updated_at"].Time(),
	}
}

func mapPlanEntitlement(record gdb.Record) service.PlanEntitlement {
	metadata := map[string]any{}
	if raw := record["metadata"].String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &metadata)
	}
	return service.PlanEntitlement{
		Plan:       record["plan"].String(),
		FeatureKey: record["feature_key"].String(),
		Enabled:    record["enabled"].Bool(),
		LimitValue: nullableInt64(record["limit_value"]),
		Metadata:   metadata,
		CreatedAt:  record["created_at"].Time(),
		UpdatedAt:  record["updated_at"].Time(),
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

func nullableInt64(value any) *int64 {
	v, ok := value.(interface {
		IsNil() bool
		Int64() int64
	})
	if !ok || v.IsNil() {
		return nil
	}
	i := v.Int64()
	return &i
}
