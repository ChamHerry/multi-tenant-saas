package tenant

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

var (
	slugPattern       = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,78}[a-z0-9]$`)
	internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

const tenantSlugRuleMessage = "tenant slug must be 3-80 lowercase letters, numbers, or hyphens and must start/end with a letter or number"

type sTenant struct{}

func init() {
	s := &sTenant{}
	service.RegisterTenantProvision(s)
	service.RegisterTenantAdmin(s)
}

func (s *sTenant) CreateTenant(ctx context.Context, in service.CreateTenantInput) (*service.Tenant, error) {
	if err := validateCreateTenantInput(in); err != nil {
		return nil, err
	}
	if in.OwnerUserID != "" {
		owner, err := service.UserService().GetUser(ctx, in.OwnerUserID)
		if err != nil {
			return nil, err
		}
		if owner.Status != "active" {
			return nil, gerror.Newf("owner user %s is not active", in.OwnerUserID)
		}
	}
	tenantID := uuid.GenerateV4()
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(defaultMetadata(in.Metadata))
	if err != nil {
		return nil, gerror.Wrap(err, "marshal tenant metadata")
	}
	now := time.Now()

	cols := dao.Tenants.Columns()
	_, err = dao.Tenants.Ctx(ctx).Data(g.Map{
		cols.Id:        tenantID,
		cols.Name:      in.Name,
		cols.Slug:      in.Slug,
		cols.Status:    "active",
		cols.Metadata:  string(metadata),
		cols.CreatedAt: "now()",
		cols.UpdatedAt: "now()",
	}).Insert()
	if err != nil {
		return nil, gerror.Wrap(err, "insert public tenant metadata")
	}

	created := &service.Tenant{
		ID:          tenantID,
		Name:        in.Name,
		Slug:        in.Slug,
		Status:      "active",
		OwnerUserID: in.OwnerUserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if in.OwnerUserID != "" {
		if _, err = service.TenantMembershipService().AddMember(ctx, service.AddTenantMemberInput{
			TenantID: tenantID,
			UserID:   in.OwnerUserID,
			Role:     "owner",
			Status:   "active",
		}); err != nil {
			return created, err
		}
	}
	if err = service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: in.OwnerUserID, Action: "tenant.create", ResourceType: "tenant", ResourceID: tenantID, Metadata: map[string]any{
		"template": "multi_tenant_saas",
	}}); err != nil {
		return created, err
	}

	return created, nil
}

func (s *sTenant) GetTenant(ctx context.Context, tenantSelector string) (*service.Tenant, error) {
	where, arg, err := tenantSelectorWhere(tenantSelector)
	if err != nil {
		return nil, err
	}
	return fetchTenantByWhere(ctx, where, arg)
}

func (s *sTenant) UpdateTenant(ctx context.Context, tenantID string, in service.UpdateTenantInput) (*service.Tenant, error) {
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	if in.Slug != "" && !slugPattern.MatchString(in.Slug) {
		return nil, invalidTenantSlugError(in.Slug)
	}
	if in.Name == "" && in.Slug == "" && in.Metadata == nil {
		return fetchTenant(ctx, tenantID)
	}
	metadata := ""
	if in.Metadata != nil {
		payload, err := json.Marshal(defaultMetadata(in.Metadata))
		if err != nil {
			return nil, gerror.Wrap(err, "marshal tenant metadata")
		}
		metadata = string(payload)
	}

	cols := dao.Tenants.Columns()
	data := g.Map{cols.UpdatedAt: "now()"}
	if in.Name != "" {
		data[cols.Name] = in.Name
	}
	if in.Slug != "" {
		data[cols.Slug] = in.Slug
	}
	if metadata != "" {
		data[cols.Metadata] = metadata
	}

	_, err := dao.Tenants.Ctx(ctx).
		Where(cols.Id, tenantID).
		Where("deleted_at IS NULL").
		Data(data).
		Update()
	if err != nil {
		return nil, gerror.Wrap(err, "update tenant")
	}
	return fetchTenant(ctx, tenantID)
}

func (s *sTenant) SuspendTenant(ctx context.Context, tenantID, actorUserID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if actorUserID != "" {
		if err := validateInternalID(actorUserID); err != nil {
			return err
		}
	}

	cols := dao.Tenants.Columns()
	result, err := dao.Tenants.Ctx(ctx).
		Where(cols.Id, tenantID).
		Where("deleted_at IS NULL").
		Where(cols.Status+" <> ?", "deleted").
		Data(g.Map{cols.Status: "suspended", cols.UpdatedAt: "now()"}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "suspend tenant")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.Newf("tenant %s not found", tenantID)
	}
	return service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.suspend", ResourceType: "tenant", ResourceID: tenantID})
}

func (s *sTenant) RestoreTenant(ctx context.Context, tenantID, actorUserID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if actorUserID != "" {
		if err := validateInternalID(actorUserID); err != nil {
			return err
		}
	}

	cols := dao.Tenants.Columns()
	result, err := dao.Tenants.Ctx(ctx).
		Where(cols.Id, tenantID).
		Where("deleted_at IS NULL").
		Where(cols.Status+" <> ?", "deleted").
		Data(g.Map{cols.Status: "active", cols.UpdatedAt: "now()"}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "restore tenant")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.Newf("tenant %s not found", tenantID)
	}
	return service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.restore", ResourceType: "tenant", ResourceID: tenantID})
}

func (s *sTenant) DeleteTenant(ctx context.Context, tenantID, actorUserID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if actorUserID != "" {
		if err := validateInternalID(actorUserID); err != nil {
			return err
		}
	}

	cols := dao.Tenants.Columns()
	result, err := dao.Tenants.Ctx(ctx).
		Where(cols.Id, tenantID).
		Where("deleted_at IS NULL").
		Data(g.Map{cols.Status: "deleted", cols.DeletedAt: "now()", cols.UpdatedAt: "now()"}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "delete tenant")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.Newf("tenant %s not found", tenantID)
	}
	if err = service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.delete", ResourceType: "tenant", ResourceID: tenantID}); err != nil {
		return err
	}
	purgeDelayHours := g.Cfg().MustGet(ctx, "tenant.lifecycle.purgeDelayHours", 720).Int()
	if purgeDelayHours > 0 {
		_, _ = service.TenantLifecycle().RequestPurge(ctx, tenantID, actorUserID, time.Now().Add(time.Duration(purgeDelayHours)*time.Hour))
	}
	return nil
}

func validateCreateTenantInput(in service.CreateTenantInput) error {
	if in.Name == "" {
		return gerror.New("tenant name is required")
	}
	if !slugPattern.MatchString(in.Slug) {
		return invalidTenantSlugError(in.Slug)
	}
	if in.OwnerUserID == "" && !in.SystemOwnerless {
		return gerror.New("owner user id is required; pass SystemOwnerless only for explicit ops/system tenants")
	}
	if in.OwnerUserID != "" {
		return validateInternalID(in.OwnerUserID)
	}
	return nil
}

func invalidTenantSlugError(slug string) error {
	return gerror.NewCodef(gcode.CodeInvalidParameter, "%s: %q", tenantSlugRuleMessage, slug)
}

func validateInternalID(tenantID string) error {
	if !internalIDPattern.MatchString(tenantID) {
		return gerror.Newf("invalid internal uuid %q", tenantID)
	}
	return nil
}

func fetchTenant(ctx context.Context, tenantID string) (*service.Tenant, error) {
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	return fetchTenantByWhere(ctx, "id=?", tenantID)
}

func fetchTenantByWhere(ctx context.Context, where string, arg any) (*service.Tenant, error) {
	cols := dao.Tenants.Columns()
	record, err := dao.Tenants.Ctx(ctx).
		Fields(cols.Id, cols.Name, cols.Slug, cols.Status, cols.CreatedAt, cols.UpdatedAt).
		Where(where, arg).
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select tenant")
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("tenant not found")
	}
	id := record["id"].String()
	if !internalIDPattern.MatchString(id) {
		return nil, gerror.Newf("invalid tenant id %q", id)
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

func tenantSelectorWhere(selector string) (where string, arg string, err error) {
	if internalIDPattern.MatchString(selector) {
		return "id=?", selector, nil
	}
	if slugPattern.MatchString(selector) {
		return "slug=?", selector, nil
	}
	return "", "", gerror.Newf("invalid tenant selector %q", selector)
}

func defaultMetadata(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	return in
}
