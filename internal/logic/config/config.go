// Package config implements the IConfig service with Redis read-through caching
// and database-backed configuration storage with tenant-level override support.
//
// All Get* methods automatically resolve the tenant from BizCtx.
// Reading order: tenant cache → global cache → DB tenant → DB global → default.
package config

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/entity"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/crypto"
	credis "multi-tenant-saas/utility/redis"
)

const (
	cacheKeyFmtGlobal = "config:global:key:%s"   // config:global:key:auth.session.absoluteTTL
	cacheKeyFmtTenant = "config:tenant:%s:key:%s" // config:tenant:<tid>:key:rateLimit.rps
	cacheTTL          = 3600 * time.Second        // 1 hour
)

func init() {
	service.RegisterConfig(&sConfig{})
}

type sConfig struct{}

// resolveTenantID extracts the tenant ID from BizCtx. Returns "" if none.
func resolveTenantID(ctx context.Context) string {
	return service.BizCtx().GetTenantID(ctx)
}

// =============================================================================
// Public typed getters — tenant resolved from BizCtx automatically
// =============================================================================

func (s *sConfig) GetString(ctx context.Context, key string, defaultVal string) string {
	val, err := s.read(ctx, key)
	if err != nil || val == "" {
		return defaultVal
	}
	return val
}

func (s *sConfig) GetInt(ctx context.Context, key string, defaultVal int) int {
	val, err := s.read(ctx, key)
	if err != nil {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		g.Log().Warningf(ctx, "config[%s] parse int failed: %v, using default %d", key, err, defaultVal)
		return defaultVal
	}
	return n
}

func (s *sConfig) GetFloat(ctx context.Context, key string, defaultVal float64) float64 {
	val, err := s.read(ctx, key)
	if err != nil {
		return defaultVal
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		g.Log().Warningf(ctx, "config[%s] parse float failed: %v, using default %f", key, err, defaultVal)
		return defaultVal
	}
	return f
}

func (s *sConfig) GetBool(ctx context.Context, key string, defaultVal bool) bool {
	val, err := s.read(ctx, key)
	if err != nil {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		g.Log().Warningf(ctx, "config[%s] parse bool failed: %v, using default %v", key, err, defaultVal)
		return defaultVal
	}
	return b
}

func (s *sConfig) GetDuration(ctx context.Context, key string, defaultVal time.Duration) time.Duration {
	val, err := s.read(ctx, key)
	if err != nil {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		g.Log().Warningf(ctx, "config[%s] parse duration failed: %v, using default %v", key, err, defaultVal)
		return defaultVal
	}
	return d
}

// =============================================================================
// Core read chain: tenant Redis → global Redis → DB tenant → DB global → error
// =============================================================================

func (s *sConfig) read(ctx context.Context, key string) (string, error) {
	tenantID := resolveTenantID(ctx)
	return s.readWithTenant(ctx, key, tenantID)
}

// readWithTenant is the core read logic with an explicit tenantID parameter.
func (s *sConfig) readWithTenant(ctx context.Context, key string, tenantID string) (string, error) {
	adapter := credis.GetCacheManager().GetAdapter("default")

	// ① Try tenant Redis cache
	if tenantID != "" && adapter != nil {
		tenantKey := fmt.Sprintf(cacheKeyFmtTenant, tenantID, key)
		if val, err := adapter.Get(ctx, tenantKey); err == nil && val != "" {
			return val, nil
		}
	}

	// ② Try global Redis cache
	if adapter != nil {
		globalKey := fmt.Sprintf(cacheKeyFmtGlobal, key)
		if val, err := adapter.Get(ctx, globalKey); err == nil && val != "" {
			// Backfill tenant cache so next read needs only 1 Redis round-trip
			if tenantID != "" {
				tenantKey := fmt.Sprintf(cacheKeyFmtTenant, tenantID, key)
				if err := adapter.Set(ctx, tenantKey, val, cacheTTL); err != nil {
					g.Log().Warningf(ctx, "config tenant-cache backfill[%s] failed: %v", tenantKey, err)
				}
			}
			return val, nil
		}
	}

	// ③ Try DB tenant-level
	if tenantID != "" {
		config, err := s.queryDB(ctx, key, tenantID)
		if err == nil && config != nil {
			val, decErr := s.decryptIfNeeded(ctx, config)
			if decErr != nil {
				return "", decErr
			}
			s.fillCache(ctx, adapter, key, tenantID, val)
			return val, nil
		}
	}

	// ④ Try DB global level
	config, err := s.queryDB(ctx, key, "")
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", gerror.NewCode(gcode.CodeNotFound, fmt.Sprintf("config not found: %s", key))
	}
	val, decErr := s.decryptIfNeeded(ctx, config)
	if decErr != nil {
		return "", decErr
	}
	// Fill global cache + also tenant cache (avoid 2-step lookup next time)
	s.fillCache(ctx, adapter, key, "", val)
	if tenantID != "" {
		s.fillCache(ctx, adapter, key, tenantID, val)
	}
	return val, nil
}

func (s *sConfig) queryDB(ctx context.Context, key string, tenantID string) (*entity.SystemConfig, error) {
	var config *entity.SystemConfig
	m := dao.SystemConfig.Ctx(ctx).Where(dao.SystemConfig.Columns().Key, key)
	if tenantID != "" {
		m = m.Where(dao.SystemConfig.Columns().TenantId, tenantID)
	} else {
		m = m.WhereNull(dao.SystemConfig.Columns().TenantId)
	}
	err := m.Scan(&config)
	return config, err
}

func (s *sConfig) fillCache(ctx context.Context, adapter *credis.RedisAdapter,
	key string, tenantID string, value string) {
	if adapter == nil || value == "" {
		return
	}
	var cacheKey string
	if tenantID != "" {
		cacheKey = fmt.Sprintf(cacheKeyFmtTenant, tenantID, key)
	} else {
		cacheKey = fmt.Sprintf(cacheKeyFmtGlobal, key)
	}
	if err := adapter.Set(ctx, cacheKey, value, cacheTTL); err != nil {
		g.Log().Warningf(ctx, "config cache set[%s] failed: %v", cacheKey, err)
	}
}

func (s *sConfig) decryptIfNeeded(ctx context.Context, config *entity.SystemConfig) (string, error) {
	if config.ValueType != "secret" || config.Value == "" {
		return config.Value, nil
	}
	decrypted, err := crypto.Decrypt(config.Value)
	if err != nil {
		g.Log().Errorf(ctx, "config[%s] decrypt failed: %v", config.Key, err)
		return "", gerror.Wrap(err, "decrypt config failed")
	}
	return decrypted, nil
}

// =============================================================================
// Batch read with MGET
// =============================================================================

func (s *sConfig) GetStrings(ctx context.Context, keys []string) (map[string]string, error) {
	tenantID := resolveTenantID(ctx)
	result := make(map[string]string, len(keys))
	adapter := credis.GetCacheManager().GetAdapter("default")

	var missing []string

	if adapter != nil {
		// Build cache keys with tenant prefix when available
		cacheKeys := make([]string, len(keys))
		for i, k := range keys {
			if tenantID != "" {
				cacheKeys[i] = fmt.Sprintf(cacheKeyFmtTenant, tenantID, k)
			} else {
				cacheKeys[i] = fmt.Sprintf(cacheKeyFmtGlobal, k)
			}
		}

		vals, err := adapter.MGet(ctx, cacheKeys...)
		if err == nil {
			for i, v := range vals {
				if v != nil {
					if s, ok := v.(string); ok && s != "" {
						result[keys[i]] = s
						continue
					}
				}
				// Tenant miss — try global cache and backfill
				if tenantID != "" {
					globalKey := fmt.Sprintf(cacheKeyFmtGlobal, keys[i])
					if gv, gerr := adapter.Get(ctx, globalKey); gerr == nil && gv != "" {
						result[keys[i]] = gv
						// Backfill tenant cache
						tenantKey := fmt.Sprintf(cacheKeyFmtTenant, tenantID, keys[i])
						if serr := adapter.Set(ctx, tenantKey, gv, cacheTTL); serr != nil {
							g.Log().Warningf(ctx, "config tenant-cache backfill[%s] failed: %v", tenantKey, serr)
						}
						continue
					}
				}
				missing = append(missing, keys[i])
			}
		} else {
			missing = keys
		}
	} else {
		missing = keys
	}

	for _, k := range missing {
		val, err := s.readWithTenant(ctx, k, tenantID)
		if err == nil {
			result[k] = val
		}
	}

	return result, nil
}

// =============================================================================
// Admin operations
// =============================================================================

func (s *sConfig) Set(ctx context.Context, params *service.ConfigSetParams) error {
	value := params.Value
	isEncrypted := false

	if params.ValueType == "secret" && value != "" {
		encrypted, err := crypto.Encrypt(value)
		if err != nil {
			return fmt.Errorf("encrypt config value: %w", err)
		}
		value = encrypted
		isEncrypted = true
	}

	var tenantID interface{}
	if params.TenantID != nil && *params.TenantID != "" {
		tenantID = *params.TenantID
	}

	_, err := dao.SystemConfig.Ctx(ctx).
		Data(g.Map{
			dao.SystemConfig.Columns().Key:         params.Key,
			dao.SystemConfig.Columns().Value:       value,
			dao.SystemConfig.Columns().ValueType:   params.ValueType,
			dao.SystemConfig.Columns().TenantId:    tenantID,
			dao.SystemConfig.Columns().Description: params.Description,
			dao.SystemConfig.Columns().IsEncrypted: isEncrypted,
		}).
		OnConflict(dao.SystemConfig.Columns().Key, dao.SystemConfig.Columns().TenantId).
		Save()

	if err != nil {
		return gerror.Wrap(err, "save config failed")
	}

	s.invalidateCache(ctx, params.Key, params.TenantID)
	return nil
}

func (s *sConfig) Delete(ctx context.Context, key string, tenantID *string) error {
	m := dao.SystemConfig.Ctx(ctx).Where(dao.SystemConfig.Columns().Key, key)
	if tenantID != nil && *tenantID != "" {
		m = m.Where(dao.SystemConfig.Columns().TenantId, *tenantID)
	} else {
		m = m.WhereNull(dao.SystemConfig.Columns().TenantId)
	}
	_, err := m.Delete()
	if err != nil {
		return gerror.Wrap(err, "delete config failed")
	}

	s.invalidateCache(ctx, key, tenantID)
	return nil
}

func (s *sConfig) invalidateCache(ctx context.Context, key string, tenantID *string) {
	adapter := credis.GetCacheManager().GetAdapter("default")
	if adapter == nil {
		return
	}
	if tenantID != nil && *tenantID != "" {
		if _, err := adapter.Del(ctx, fmt.Sprintf(cacheKeyFmtTenant, *tenantID, key)); err != nil {
			g.Log().Warningf(ctx, "config cache del[tenant:%s] failed: %v", key, err)
		}
	}
	if _, err := adapter.Del(ctx, fmt.Sprintf(cacheKeyFmtGlobal, key)); err != nil {
		g.Log().Warningf(ctx, "config cache del[global:%s] failed: %v", key, err)
	}
}

