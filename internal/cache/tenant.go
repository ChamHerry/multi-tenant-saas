package cache

import (
	"context"
	"time"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
)

var tenantCache = New[service.Tenant]()

const tenantTTL = 60 * time.Second

// GetTenant returns a cached tenant or fetches from DB via DAO.
func GetTenant(ctx context.Context, tenantID string) (*service.Tenant, error) {
	if v, ok := tenantCache.Get("tenant:" + tenantID); ok {
		return &v, nil
	}
	cols := dao.Tenants.Columns()
	record, err := dao.Tenants.Ctx(ctx).
		Where(cols.Id, tenantID).
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, nil
	}
	t := service.Tenant{
		ID:        record[cols.Id].String(),
		Name:      record[cols.Name].String(),
		Slug:      record[cols.Slug].String(),
		Status:    record[cols.Status].String(),
		CreatedAt: record[cols.CreatedAt].Time(),
		UpdatedAt: record[cols.UpdatedAt].Time(),
	}
	tenantCache.Set("tenant:"+tenantID, t, tenantTTL)
	return &t, nil
}

// InvalidateTenant removes cached tenant entries.
func InvalidateTenant(tenantID string) {
	tenantCache.Invalidate("tenant:" + tenantID)
}
