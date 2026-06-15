package membership

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/do"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

var (
	internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	slugPattern       = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,78}[a-z0-9]$`)
	nonSlugChars      = regexp.MustCompile(`[^a-z0-9]+`)
	duplicateHyphens  = regexp.MustCompile(`-+`)
	allowedRoles      = map[string]struct{}{
		"owner":  {},
		"admin":  {},
		"member": {},
		"viewer": {},
	}
	allowedStatuses = map[string]struct{}{
		"invited":   {},
		"active":    {},
		"suspended": {},
		"removed":   {},
	}
)

type sTenantMembership struct{}

func init() {
	service.RegisterTenantMembership(&sTenantMembership{})
}

func (s *sTenantMembership) AddMember(ctx context.Context, in service.AddTenantMemberInput) (*service.TenantMembership, error) {
	if err := validateAddMemberInput(in); err != nil {
		return nil, err
	}
	status := defaultString(in.Status, "active")
	if err := ensureActiveUser(ctx, in.UserID); err != nil {
		return nil, err
	}
	if in.InvitedByUserID != "" {
		if err := ensureActiveUser(ctx, in.InvitedByUserID); err != nil {
			return nil, err
		}
	}
	if err := ensureActiveTenant(ctx, in.TenantID); err != nil {
		return nil, err
	}
	if in.Role != "owner" {
		if err := ensureNotRemovingLastOwner(ctx, in.TenantID, in.UserID); err != nil {
			return nil, err
		}
	}

	membershipID := uuid.GenerateV4()
	if err := validateInternalID(membershipID); err != nil {
		return nil, err
	}

	cols := dao.TenantMemberships.Columns()
	now := time.Now()

	// Check for existing membership, then Insert or Update
	existing, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, in.TenantID).
		Where(cols.UserId, in.UserID).
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select existing membership")
	}

	if existing.IsEmpty() {
		var invitedByVal any
		if in.InvitedByUserID != "" {
			invitedByVal = in.InvitedByUserID
		}
		_, err = dao.TenantMemberships.Ctx(ctx).Data(do.TenantMemberships{
			Id:              membershipID,
			TenantId:        in.TenantID,
			UserId:          in.UserID,
			Role:            in.Role,
			Status:          status,
			InvitedByUserId: invitedByVal,
		}).Insert()
		if err != nil {
			return nil, gerror.Wrap(err, "insert tenant membership")
		}
	} else {
		data := do.TenantMemberships{
			Role:   in.Role,
			Status: status,
		}
		if in.InvitedByUserID != "" {
			data.InvitedByUserId = in.InvitedByUserID
		}
		// If status was 'invited' and new status is 'active', set joined_at
		if existing[cols.Status].String() == "invited" && status == "active" {
			data.JoinedAt = now
		}
		_, err = dao.TenantMemberships.Ctx(ctx).
			Where(cols.TenantId, in.TenantID).
			Where(cols.UserId, in.UserID).
			Data(data).
			Update()
		if err != nil {
			return nil, gerror.Wrap(err, "update tenant membership")
		}
	}

	// Fetch the result
	record, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, in.TenantID).
		Where(cols.UserId, in.UserID).
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select tenant membership")
	}
	return mapMembership(record)
}

func (s *sTenantMembership) RemoveMember(ctx context.Context, tenantID, userID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if err := validateInternalID(userID); err != nil {
		return err
	}
	if err := ensureNotRemovingLastOwner(ctx, tenantID, userID); err != nil {
		return err
	}
	cols := dao.TenantMemberships.Columns()
	return dao.TenantMemberships.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := dao.TenantMemberships.Ctx(ctx).TX(tx).
			Where(cols.TenantId, tenantID).
			Where(cols.UserId, userID).
			Data(do.TenantMemberships{Status: "removed"}).
			Update()
		if err != nil {
			return gerror.Wrap(err, "remove tenant membership")
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return gerror.Newf("tenant membership tenant=%s user=%s not found", tenantID, userID)
		}
		if _, err = dao.TenantMemberships.Ctx(ctx).TX(tx).
			Where(cols.TenantId, tenantID).
			Where(cols.UserId, userID).
			Delete(); err != nil {
			return gerror.Wrap(err, "remove tenant membership")
		}
		return nil
	})
}

func (s *sTenantMembership) ChangeRole(ctx context.Context, tenantID, userID, role string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if err := validateInternalID(userID); err != nil {
		return err
	}
	if err := validateRole(role); err != nil {
		return err
	}
	if role != "owner" {
		if err := ensureNotRemovingLastOwner(ctx, tenantID, userID); err != nil {
			return err
		}
	}
	cols := dao.TenantMemberships.Columns()
	result, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.UserId, userID).
		Data(do.TenantMemberships{Role: role}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "change tenant member role")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.Newf("tenant membership tenant=%s user=%s not found", tenantID, userID)
	}
	return nil
}

func (s *sTenantMembership) UpdateMember(ctx context.Context, tenantID, userID string, in service.UpdateTenantMemberInput) (*service.TenantMembership, error) {
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	if in.Role != "" {
		if err := validateRole(in.Role); err != nil {
			return nil, err
		}
	}
	if in.Status != "" {
		if err := validateStatus(in.Status); err != nil {
			return nil, err
		}
	}
	if in.Role == "" && in.Status == "" {
		return s.getMember(ctx, tenantID, userID)
	}
	if (in.Role != "" && in.Role != "owner") || in.Status == "removed" || in.Status == "suspended" {
		if err := ensureNotRemovingLastOwner(ctx, tenantID, userID); err != nil {
			return nil, err
		}
	}
	cols := dao.TenantMemberships.Columns()
	data := do.TenantMemberships{}
	if in.Role != "" {
		data.Role = in.Role
	}
	if in.Status != "" {
		data.Status = in.Status
	}
	record, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.UserId, userID).
		Data(data).
		Update()
	if err != nil {
		return nil, gerror.Wrap(err, "update tenant membership")
	}
	// For Update(), we need to fetch the updated record
	if record != nil {
		_ = record
	}
	return s.getMember(ctx, tenantID, userID)
}

func (s *sTenantMembership) getMember(ctx context.Context, tenantID, userID string) (*service.TenantMembership, error) {
	cols := dao.TenantMemberships.Columns()
	record, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.UserId, userID).
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select tenant membership")
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("tenant membership tenant=%s user=%s not found", tenantID, userID)
	}
	return mapMembership(record)
}

func (s *sTenantMembership) ListUserTenants(ctx context.Context, userID string) ([]service.TenantMembershipWithTenant, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	items, err := listUserTenantItems(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		if err = ensureDefaultTenantForUser(ctx, userID); err != nil {
			return nil, err
		}
		items, err = listUserTenantItems(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	return items, nil
}

func listUserTenantItems(ctx context.Context, userID string) ([]service.TenantMembershipWithTenant, error) {
	// JOIN query for user memberships with tenant info
	result, err := dao.TenantMemberships.Ctx(ctx).
		LeftJoin("tenants t", "t.id = tenant_memberships.tenant_id").
		Fields("tenant_memberships.*, t.name AS tenant_name, t.slug AS tenant_slug, t.status AS tenant_status").
		Where("tenant_memberships.user_id", userID).
		Where("tenant_memberships.status IN(?)", g.Slice{"active", "invited"}).
		OrderDesc("tenant_memberships.created_at").
		All()
	if err != nil {
		return nil, gerror.Wrap(err, "list user tenants")
	}
	items := make([]service.TenantMembershipWithTenant, 0, len(result))
	for _, record := range result {
		membership, err := mapMembership(record)
		if err != nil {
			return nil, err
		}
		items = append(items, service.TenantMembershipWithTenant{
			TenantMembership: *membership,
			TenantName:       record["tenant_name"].String(),
			TenantSlug:       record["tenant_slug"].String(),
			TenantStatus:     record["tenant_status"].String(),
		})
	}
	return items, nil
}

func (s *sTenantMembership) ListTenantMembers(ctx context.Context, tenantID string) ([]service.TenantMembershipWithUser, error) {
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	// JOIN query for tenant memberships with user info
	result, err := dao.TenantMemberships.Ctx(ctx).
		LeftJoin("users u", "u.id = tenant_memberships.user_id").
		Fields("tenant_memberships.*, u.email, u.display_name, u.status AS user_status").
		Where("tenant_memberships.tenant_id", tenantID).
		OrderDesc("tenant_memberships.created_at").
		All()
	if err != nil {
		return nil, gerror.Wrap(err, "list tenant members")
	}
	items := make([]service.TenantMembershipWithUser, 0, len(result))
	for _, record := range result {
		membership, err := mapMembership(record)
		if err != nil {
			return nil, err
		}
		items = append(items, service.TenantMembershipWithUser{
			TenantMembership: *membership,
			Email:            record["email"].String(),
			DisplayName:      record["display_name"].String(),
			UserStatus:       record["user_status"].String(),
		})
	}
	return items, nil
}

func (s *sTenantMembership) ResolveTenantContext(ctx context.Context, userID, tenantSelector string) (*service.TenantContext, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	if _, _, err := tenantSelectorWhere(tenantSelector); err != nil {
		return nil, err
	}

	ctxRecord, found, err := resolveTenantContextRecord(ctx, userID, tenantSelector)
	if err != nil {
		return nil, err
	}
	if found {
		return ctxRecord, nil
	}
	return nil, gerror.Newf("active tenant membership not found for user=%s selector=%s", userID, tenantSelector)
}

func resolveTenantContextRecord(ctx context.Context, userID, tenantSelector string) (*service.TenantContext, bool, error) {
	where, arg, err := tenantSelectorWhere(tenantSelector)
	if err != nil {
		return nil, false, err
	}
	record, err := dao.TenantMemberships.Ctx(ctx).
		LeftJoin("tenants t", "t.id = tenant_memberships.tenant_id").
		Fields("tenant_memberships.tenant_id, tenant_memberships.user_id, tenant_memberships.role, t.slug AS tenant_slug").
		Where("tenant_memberships.user_id", userID).
		Where(where, arg).
		Where("tenant_memberships.status", "active").
		Where("t.status", "active").
		One()
	if err != nil {
		return nil, false, gerror.Wrap(err, "resolve tenant context")
	}
	if record.IsEmpty() {
		return nil, false, nil
	}
	return &service.TenantContext{
		TenantID:   record["tenant_id"].String(),
		UserID:     record["user_id"].String(),
		Role:       record["role"].String(),
		TenantSlug: record["tenant_slug"].String(),
	}, true, nil
}

func validateAddMemberInput(in service.AddTenantMemberInput) error {
	if err := validateInternalID(in.TenantID); err != nil {
		return err
	}
	if err := validateInternalID(in.UserID); err != nil {
		return err
	}
	if in.InvitedByUserID != "" {
		if err := validateInternalID(in.InvitedByUserID); err != nil {
			return err
		}
	}
	if err := validateRole(in.Role); err != nil {
		return err
	}
	return validateStatus(defaultString(in.Status, "active"))
}

func validateRole(role string) error {
	if _, ok := allowedRoles[role]; !ok {
		return gerror.Newf("invalid tenant role %q", role)
	}
	return nil
}

func validateStatus(status string) error {
	if _, ok := allowedStatuses[status]; !ok {
		return gerror.Newf("invalid tenant membership status %q", status)
	}
	return nil
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.Newf("invalid internal uuid %q", id)
	}
	return nil
}

func tenantSelectorWhere(selector string) (where string, arg string, err error) {
	if internalIDPattern.MatchString(selector) {
		return "t.id=?", selector, nil
	}
	if slugPattern.MatchString(selector) {
		return "t.slug=?", selector, nil
	}
	return "", "", gerror.Newf("invalid tenant selector %q", selector)
}

func ensureDefaultTenantForUser(ctx context.Context, userID string) error {
	user, err := service.UserService().GetUser(ctx, userID)
	if err != nil {
		return err
	}
	if user.Status != "active" {
		return gerror.Newf("user %s is not active", userID)
	}
	_, err = service.TenantProvision().CreateTenant(ctx, service.CreateTenantInput{
		Name:        defaultTenantName(user),
		Slug:        defaultTenantSlug(user),
		OwnerUserID: userID,
		Metadata: map[string]any{
			"default_personal_tenant": true,
			"source":                  "registration_default",
		},
	})
	return err
}

func defaultTenantName(user *service.User) string {
	label := strings.TrimSpace(user.Email)
	if label == "" {
		label = strings.TrimSpace(user.DisplayName)
	}
	if label == "" {
		shortID := strings.ReplaceAll(user.ID, "-", "")
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		if shortID == "" {
			shortID = "unknown"
		}
		label = "user-" + shortID
	}
	return label + " 的组织"
}

func defaultTenantSlug(user *service.User) string {
	seed := strings.TrimSpace(user.Email)
	if seed == "" {
		seed = strings.TrimSpace(user.DisplayName)
	}
	seed = strings.ToLower(seed)
	base := nonSlugChars.ReplaceAllString(seed, "-")
	base = duplicateHyphens.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" || len(base) < 2 {
		base = "user"
	}
	suffix := strings.ReplaceAll(user.ID, "-", "")
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	maxBaseLen := 80 - len(suffix) - 1
	if len(base) > maxBaseLen {
		base = strings.Trim(base[:maxBaseLen], "-")
	}
	if len(base) < 2 {
		base = "user"
	}
	return base + "-" + suffix
}

func ensureActiveUser(ctx context.Context, userID string) error {
	cols := dao.Users.Columns()
	record, err := dao.Users.Ctx(ctx).
		Where(cols.Id, userID).
		Where(cols.Status, "active").
		One()
	if err != nil {
		return gerror.Wrap(err, "select active user")
	}
	if record.IsEmpty() {
		return gerror.Newf("active user %s not found", userID)
	}
	return nil
}

func ensureActiveTenant(ctx context.Context, tenantID string) error {
	cols := dao.Tenants.Columns()
	record, err := dao.Tenants.Ctx(ctx).
		Where(cols.Id, tenantID).
		Where(cols.Status, "active").
		One()
	if err != nil {
		return gerror.Wrap(err, "select active tenant")
	}
	if record.IsEmpty() {
		return gerror.Newf("active tenant %s not found", tenantID)
	}
	return nil
}

func ensureNotRemovingLastOwner(ctx context.Context, tenantID, userID string) error {
	cols := dao.TenantMemberships.Columns()
	record, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.UserId, userID).
		Where(cols.Role, "owner").
		Where(cols.Status, "active").
		One()
	if err != nil {
		return gerror.Wrap(err, "select existing tenant membership role")
	}
	if record.IsEmpty() || record[cols.Role].String() != "owner" {
		return nil
	}
	otherOwner, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.UserId+" <> ?", userID).
		Where(cols.Role, "owner").
		Where(cols.Status, "active").
		One()
	if err != nil {
		return gerror.Wrap(err, "select other tenant owner")
	}
	if otherOwner.IsEmpty() {
		return gerror.Newf("cannot remove or demote the last active owner of tenant %s", tenantID)
	}
	return nil
}

func mapMembership(record gdb.Record) (*service.TenantMembership, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	tenantID := record["tenant_id"].String()
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	userID := record["user_id"].String()
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	return &service.TenantMembership{
		ID:              id,
		TenantID:        tenantID,
		UserID:          userID,
		Role:            record["role"].String(),
		Status:          record["status"].String(),
		InvitedByUserID: record["invited_by_user_id"].String(),
		JoinedAt:        nullableTime(record["joined_at"]),
		CreatedAt:       record["created_at"].Time(),
		UpdatedAt:       record["updated_at"].Time(),
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

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
