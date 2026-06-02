package platformadmin

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

var rolePermissions = map[string][]service.PlatformPermission{
	"super_admin": {
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionTenantManage,
		service.PlatformPermissionUserRead,
		service.PlatformPermissionUserManage,
		service.PlatformPermissionAuditRead,
		service.PlatformPermissionAdminManage,
		service.PlatformPermissionConfigRead,
		service.PlatformPermissionConfigManage,
	},
	"support": {
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionUserRead,
		service.PlatformPermissionAuditRead,
		service.PlatformPermissionConfigRead,
	},
	"auditor": {
		service.PlatformPermissionTenantRead,
		service.PlatformPermissionAuditRead,
		service.PlatformPermissionConfigRead,
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
	cols := dao.PlatformAdmins.Columns()
	record, err := dao.PlatformAdmins.Ctx(ctx).
		Where(cols.UserId, userID).
		Where(cols.Status, "active").
		One()
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

	cols := dao.Tenants.Columns()
	total, err := dao.Tenants.Ctx(ctx).Where(whereSQL, args...).Count()
	if err != nil {
		return nil, gerror.Wrap(err, "count platform tenants")
	}

	rows, err := dao.Tenants.Ctx(ctx).
		Fields(cols.Id, cols.Name, cols.Slug, cols.Status, cols.CreatedAt, cols.UpdatedAt).
		Where(whereSQL, args...).
		OrderDesc(cols.CreatedAt).
		Limit(limit).
		Offset(offset).
		All()
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
	return &service.PlatformTenantList{Items: items, Total: total}, nil
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

	cols := dao.Users.Columns()
	total, err := dao.Users.Ctx(ctx).Where(whereSQL, args...).Count()
	if err != nil {
		return nil, gerror.Wrap(err, "count platform users")
	}

	rows, err := dao.Users.Ctx(ctx).
		Fields(cols.Id, cols.Email, cols.DisplayName, cols.AvatarUrl, cols.Status, cols.LastLoginAt, cols.Metadata, cols.CreatedAt, cols.UpdatedAt).
		Where(whereSQL, args...).
		OrderDesc(cols.CreatedAt).
		Limit(limit).
		Offset(offset).
		All()
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
	return &service.PlatformUserList{Items: items, Total: total}, nil
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

	userCols := dao.Users.Columns()
	result, err := dao.Users.Ctx(ctx).
		Where(userCols.Id, targetUserID).
		Where("deleted_at IS NULL").
		Data(g.Map{userCols.Status: status, userCols.UpdatedAt: "now()"}).
		Update()
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
		where = append(where, "platform_admins.status=?")
		args = append(args, status)
	}
	if role := strings.TrimSpace(filter.Role); role != "" {
		where = append(where, "platform_admins.role=?")
		args = append(args, role)
	}
	if query := strings.TrimSpace(filter.Query); query != "" {
		where = append(where, "(u.email ILIKE ? OR u.display_name ILIKE ? OR platform_admins.user_id::text = ?)")
		like := "%" + query + "%"
		args = append(args, like, like, query)
	}
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL := strings.Join(where, " AND ")

	cols := dao.PlatformAdmins.Columns()
	total, err := dao.PlatformAdmins.Ctx(ctx).
		LeftJoin("users u", "u.id = platform_admins.user_id").
		Where(whereSQL, args...).
		Count()
	if err != nil {
		return nil, gerror.Wrap(err, "count platform admins")
	}

	rows, err := dao.PlatformAdmins.Ctx(ctx).
		LeftJoin("users u", "u.id = platform_admins.user_id").
		Fields("platform_admins.*").
		Where(whereSQL, args...).
		OrderDesc(cols.CreatedAt).
		Limit(limit).
		Offset(offset).
		All()
	if err != nil {
		return nil, gerror.Wrap(err, "list platform admins")
	}
	items := make([]service.PlatformAdmin, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPlatformAdmin(row))
	}
	return &service.PlatformAdminList{Items: items, Total: total}, nil
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

	cols := dao.PlatformAdmins.Columns()
	now := time.Now()

	// Check existing admin record, then Insert or Update
	existing, err := dao.PlatformAdmins.Ctx(ctx).Where(cols.UserId, targetUserID).One()
	if err != nil {
		return gerror.Wrap(err, "select platform admin")
	}

	if existing.IsEmpty() {
		_, err = dao.PlatformAdmins.Ctx(ctx).Data(g.Map{
			cols.UserId:          targetUserID,
			cols.Role:            role,
			cols.Status:          "active",
			cols.CreatedByUserId: actorUserID,
			cols.CreatedAt:       now,
			cols.UpdatedAt:       now,
		}).Insert()
	} else {
		_, err = dao.PlatformAdmins.Ctx(ctx).
			Where(cols.UserId, targetUserID).
			Data(g.Map{
				cols.Role:      role,
				cols.Status:    "active",
				cols.UpdatedAt: now,
			}).Update()
	}
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

	cols := dao.PlatformAdmins.Columns()
	result, err := dao.PlatformAdmins.Ctx(ctx).
		Where(cols.UserId, targetUserID).
		Where(cols.Status, "active").
		Data(g.Map{cols.Status: "suspended", cols.UpdatedAt: "now()"}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "revoke platform admin")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.NewCode(gcode.CodeNotFound, "platform admin not found")
	}
	return service.Audit().Write(ctx, service.AuditLogInput{UserID: actorUserID, Action: "platform_admin.revoke", ResourceType: "platform_admin", ResourceID: targetUserID})
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
