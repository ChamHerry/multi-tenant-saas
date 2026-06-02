// Package service defines the IConfig interface for system configuration management.
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
	TenantID    *string
	Description string
}

// IConfig provides unified access to system configuration stored in the database
// with Redis caching and tenant-level override support.
//
// All Get* methods automatically resolve tenant from BizCtx.
// Reading order: tenant Redis → global Redis → DB tenant → DB global → default value.
type IConfig interface {
	// Single-key typed getters. Tenant resolved from bizctx; nil tenant → global only.
	GetString(ctx context.Context, key string, defaultVal string) string
	GetInt(ctx context.Context, key string, defaultVal int) int
	GetFloat(ctx context.Context, key string, defaultVal float64) float64
	GetBool(ctx context.Context, key string, defaultVal bool) bool
	GetDuration(ctx context.Context, key string, defaultVal time.Duration) time.Duration

	// GetStrings batch-fetches multiple config keys using Redis MGET (single round trip).
	GetStrings(ctx context.Context, keys []string) (map[string]string, error)

	// Admin operations. TenantID nil = global.
	Set(ctx context.Context, params *ConfigSetParams) error
	Delete(ctx context.Context, key string, tenantID *string) error
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
