package cache

import (
	"context"
	"time"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
)

var userCache = New[service.User]()

const userTTL = 120 * time.Second

// GetUser returns a cached user or fetches from DB via DAO.
func GetUser(ctx context.Context, userID string) (*service.User, error) {
	if v, ok := userCache.Get("user:" + userID); ok {
		return &v, nil
	}
	cols := dao.Users.Columns()
	record, err := dao.Users.Ctx(ctx).
		Where(cols.Id, userID).
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, nil
	}
	u := service.User{
		ID:          record[cols.Id].String(),
		Email:       record[cols.Email].String(),
		DisplayName: record[cols.DisplayName].String(),
		AvatarURL:   record[cols.AvatarUrl].String(),
		Status:      record[cols.Status].String(),
		CreatedAt:   record[cols.CreatedAt].Time(),
		UpdatedAt:   record[cols.UpdatedAt].Time(),
	}
	userCache.Set("user:"+userID, u, userTTL)
	return &u, nil
}

// InvalidateUser removes cached user entries.
func InvalidateUser(userID string) {
	userCache.Invalidate("user:" + userID)
}
