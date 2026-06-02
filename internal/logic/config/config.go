// Package config implements the IConfig service with Redis read-through caching
// and database-backed global configuration storage.
//
// Reading order: Redis → DB → default value.
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
	cacheKeyFmt = "config:%s"       // config:auth.session.absoluteTTL
	cacheTTL    = 3600 * time.Second // 1 hour
)

func init() {
	service.RegisterConfig(&sConfig{})
}

type sConfig struct{}

// =============================================================================
// Public typed getters
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
// Core read: Redis → DB → error
// =============================================================================

func (s *sConfig) read(ctx context.Context, key string) (string, error) {
	adapter := credis.GetCacheManager().GetAdapter("default")

	// ① Try Redis cache
	if adapter != nil {
		cacheKey := fmt.Sprintf(cacheKeyFmt, key)
		if val, err := adapter.Get(ctx, cacheKey); err == nil && val != "" {
			return val, nil
		}
	}

	// ② Try DB
	config, err := s.queryDB(ctx, key)
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

	// ③ Fill cache
	s.fillCache(ctx, adapter, key, val)
	return val, nil
}

func (s *sConfig) queryDB(ctx context.Context, key string) (*entity.SystemConfig, error) {
	var config *entity.SystemConfig
	err := dao.SystemConfig.Ctx(ctx).
		Where(dao.SystemConfig.Columns().Key, key).
		Scan(&config)
	return config, err
}

func (s *sConfig) fillCache(ctx context.Context, adapter *credis.RedisAdapter, key string, value string) {
	if adapter == nil || value == "" {
		return
	}
	cacheKey := fmt.Sprintf(cacheKeyFmt, key)
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
	result := make(map[string]string, len(keys))
	adapter := credis.GetCacheManager().GetAdapter("default")

	var missing []string

	if adapter != nil {
		cacheKeys := make([]string, len(keys))
		for i, k := range keys {
			cacheKeys[i] = fmt.Sprintf(cacheKeyFmt, k)
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
				missing = append(missing, keys[i])
			}
		} else {
			missing = keys
		}
	} else {
		missing = keys
	}

	for _, k := range missing {
		val, err := s.read(ctx, k)
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

	_, err := dao.SystemConfig.Ctx(ctx).
		Data(g.Map{
			dao.SystemConfig.Columns().Key:         params.Key,
			dao.SystemConfig.Columns().Value:       value,
			dao.SystemConfig.Columns().ValueType:   params.ValueType,
			dao.SystemConfig.Columns().Description: params.Description,
			dao.SystemConfig.Columns().IsEncrypted: isEncrypted,
		}).
		OnConflict(dao.SystemConfig.Columns().Key).
		Save()

	if err != nil {
		return gerror.Wrap(err, "save config failed")
	}

	s.invalidateCache(ctx, params.Key)
	return nil
}

func (s *sConfig) Delete(ctx context.Context, key string) error {
	_, err := dao.SystemConfig.Ctx(ctx).
		Where(dao.SystemConfig.Columns().Key, key).
		Delete()
	if err != nil {
		return gerror.Wrap(err, "delete config failed")
	}

	s.invalidateCache(ctx, key)
	return nil
}

func (s *sConfig) invalidateCache(ctx context.Context, key string) {
	adapter := credis.GetCacheManager().GetAdapter("default")
	if adapter == nil {
		return
	}
	if _, err := adapter.Del(ctx, fmt.Sprintf(cacheKeyFmt, key)); err != nil {
		g.Log().Warningf(ctx, "config cache del[%s] failed: %v", key, err)
	}
}
