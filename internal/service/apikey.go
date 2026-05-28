package service

import (
	"context"
	"time"
)

type CreateAPIKeyInput struct {
	TenantID  string
	UserID    string
	Name      string
	Scopes    []string
	ExpiresAt *time.Time
}

type APIKey struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type CreatedAPIKey struct {
	APIKey
	RawKey string `json:"raw_key"`
}

type IAPIKey interface {
	Create(ctx context.Context, in CreateAPIKeyInput) (*CreatedAPIKey, error)
	List(ctx context.Context, tenantID string) ([]APIKey, error)
	Revoke(ctx context.Context, tenantID, apiKeyID string) error
	Authenticate(ctx context.Context, rawKey string) (*AuthIdentity, error)
}

var localAPIKey IAPIKey

func APIKeyService() IAPIKey {
	if localAPIKey == nil {
		panic("implement not found for interface IAPIKey")
	}
	return localAPIKey
}

func RegisterAPIKey(i IAPIKey) {
	localAPIKey = i
}
