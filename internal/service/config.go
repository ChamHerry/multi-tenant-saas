// Package service defines the IConfig interface for global system configuration management.
package service

import (
	"context"
	"time"
)

// ConfigSetParams holds parameters for setting a configuration value.
type ConfigSetParams struct {
	Key         string
	Value       string
	ValueType   string // string | number | bool | json | secret
	Description string
}

// IConfig provides unified access to global system configuration stored in the database
// with Redis read-through caching.
//
// Reading order: Redis → DB → default value.
type IConfig interface {
	GetString(ctx context.Context, key string, defaultVal string) string
	GetInt(ctx context.Context, key string, defaultVal int) int
	GetFloat(ctx context.Context, key string, defaultVal float64) float64
	GetBool(ctx context.Context, key string, defaultVal bool) bool
	GetDuration(ctx context.Context, key string, defaultVal time.Duration) time.Duration

	// GetStrings batch-fetches multiple config keys using Redis MGET (single round trip).
	GetStrings(ctx context.Context, keys []string) (map[string]string, error)

	// Admin operations.
	Set(ctx context.Context, params *ConfigSetParams) error
	Delete(ctx context.Context, key string) error
}

var localConfig IConfig

// Config returns the registered IConfig implementation.
func Config() IConfig {
	if localConfig == nil {
		panic("implement not registered for IConfig")
	}
	return localConfig
}

// RegisterConfig registers an IConfig implementation.
func RegisterConfig(i IConfig) {
	localConfig = i
}
