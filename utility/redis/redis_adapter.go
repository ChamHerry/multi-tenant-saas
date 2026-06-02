package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAdapter wraps go-redis client providing a simple interface
// for common Redis operations used by the application.
type RedisAdapter struct {
	client redis.UniversalClient
	name   string
}

// NewRedisAdapter creates a Redis adapter for standalone mode.
func NewRedisAdapter(ctx context.Context, group string) (*RedisAdapter, error) {
	cfg, err := LoadRedisConfig(ctx, group)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Address,
		DB:              cfg.DB,
		Password:        cfg.Password,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		DisableIdentity: true,
	})

	adapter := &RedisAdapter{client: client, name: group}

	// Verify connection
	if err := adapter.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis[%s] ping failed: %w", group, err)
	}

	return adapter, nil
}

// Get retrieves a string value by key.
func (a *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.client.Get(ctx, key).Result()
}

// GetInt64 retrieves an integer value by key. Missing keys return 0.
func (a *RedisAdapter) GetInt64(ctx context.Context, key string) (int64, error) {
	value, err := a.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return value, err
}

// Set stores a string value with TTL.
func (a *RedisAdapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return a.client.Set(ctx, key, value, ttl).Err()
}

// Del removes one or more keys.
func (a *RedisAdapter) Del(ctx context.Context, keys ...string) (int64, error) {
	return a.client.Del(ctx, keys...).Result()
}

// MGet retrieves multiple string values in a single round trip.
func (a *RedisAdapter) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return a.client.MGet(ctx, keys...).Result()
}

// Exists checks if keys exist.
func (a *RedisAdapter) Exists(ctx context.Context, keys ...string) (int64, error) {
	return a.client.Exists(ctx, keys...).Result()
}

// Expire sets a TTL on a key.
func (a *RedisAdapter) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return a.client.Expire(ctx, key, ttl).Result()
}

// IncrementWithTTL increments a counter and applies ttl when the counter is created.
func (a *RedisAdapter) IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	count, err := a.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err = a.client.Expire(ctx, key, ttl).Err(); err != nil {
			return count, err
		}
	}
	return count, nil
}

// TTL returns the remaining time to live for a key.
func (a *RedisAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return a.client.TTL(ctx, key).Result()
}

// Ping tests the connection.
func (a *RedisAdapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx).Err()
}

// Close shuts down the client.
func (a *RedisAdapter) Close() error {
	return a.client.Close()
}

// Name returns the adapter's config group name.
func (a *RedisAdapter) Name() string {
	return a.name
}
