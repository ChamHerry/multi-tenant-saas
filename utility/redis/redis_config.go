// Package redis — go-redis adapter for GoFrame
// Provides Redis connection configuration and management.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// RedisConfig holds connection parameters for a Redis instance.
type RedisConfig struct {
	Address  string
	DB       int
	Password string
	// Connection pool
	PoolSize     int
	MinIdleConns int
	// Timeouts
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

// LoadRedisConfig reads Redis configuration from GoFrame config (YAML).
// Config path: redis.<group>.*
func LoadRedisConfig(ctx context.Context, group string) (*RedisConfig, error) {
	prefix := fmt.Sprintf("redis.%s", group)

	cfg := &RedisConfig{
		Address:         g.Cfg().MustGet(ctx, prefix+".address", "127.0.0.1:6379").String(),
		DB:              g.Cfg().MustGet(ctx, prefix+".db", 0).Int(),
		Password:        g.Cfg().MustGet(ctx, prefix+".pass", "").String(),
		PoolSize:        g.Cfg().MustGet(ctx, prefix+".maxActive", 100).Int(),
		MinIdleConns:    g.Cfg().MustGet(ctx, prefix+".maxIdle", 10).Int(),
		ConnMaxLifetime: parseDuration(ctx, prefix+".maxConnLifetime", 60*time.Second),
		ConnMaxIdleTime: parseDuration(ctx, prefix+".idleTimeout", 60*time.Second),
		DialTimeout:     parseDuration(ctx, prefix+".dialTimeout", 5*time.Second),
		ReadTimeout:     parseDuration(ctx, prefix+".readTimeout", 3*time.Second),
		WriteTimeout:    parseDuration(ctx, prefix+".writeTimeout", 3*time.Second),
	}

	if cfg.Address == "" {
		return nil, gerror.Newf("redis.%s.address is required", group)
	}

	return cfg, nil
}

func parseDuration(ctx context.Context, key string, defaultVal time.Duration) time.Duration {
	val := g.Cfg().MustGet(ctx, key, "").String()
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}
