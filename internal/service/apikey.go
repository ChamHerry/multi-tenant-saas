package service

import (
	"context"
	"time"
)

const (
	APIKeyTypePersonal = "personal"
)

type CreatePersonalAPIKeyInput struct {
	UserID    string
	Name      string
	Scopes    []string
	Grants    []CreateAPIKeyTenantGrantInput
	ExpiresAt *time.Time
}

type CreateAPIKeyTenantGrantInput struct {
	TenantID string   `json:"tenant_id"`
	Scopes   []string `json:"scopes"`
}

type APIKey struct {
	ID              string              `json:"id"`
	TenantID        string              `json:"tenant_id,omitempty"`
	UserID          string              `json:"user_id"`
	Name            string              `json:"name"`
	KeyType         string              `json:"key_type"`
	KeyPrefix       string              `json:"key_prefix"`
	Scopes          []string            `json:"scopes"`
	LastUsedAt      *time.Time          `json:"last_used_at,omitempty"`
	ExpiresAt       *time.Time          `json:"expires_at,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	RevokedAt       *time.Time          `json:"revoked_at,omitempty"`
	CreatedByUserID string              `json:"created_by_user_id,omitempty"`
	TenantGrants    []APIKeyTenantGrant `json:"tenant_grants,omitempty"`
}

type APIKeyTenantGrant struct {
	ID              string     `json:"id"`
	APIKeyID        string     `json:"api_key_id"`
	TenantID        string     `json:"tenant_id"`
	TenantSlug      string     `json:"tenant_slug,omitempty"`
	TenantName      string     `json:"tenant_name,omitempty"`
	UserID          string     `json:"user_id,omitempty"`
	KeyName         string     `json:"key_name,omitempty"`
	KeyPrefix       string     `json:"key_prefix,omitempty"`
	Scopes          []string   `json:"scopes"`
	Status          string     `json:"status"`
	GrantedByUserID string     `json:"granted_by_user_id,omitempty"`
	RevokedByUserID string     `json:"revoked_by_user_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
}

type CreatedAPIKey struct {
	APIKey
	RawKey string `json:"raw_key"`
}

type IAPIKey interface {
	CreatePersonal(ctx context.Context, in CreatePersonalAPIKeyInput) (*CreatedAPIKey, error)
	ListPersonal(ctx context.Context, userID string) ([]APIKey, error)
	RevokePersonal(ctx context.Context, userID, apiKeyID string) error
	ResolveTenantGrant(ctx context.Context, apiKeyID, tenantID string) (*APIKeyTenantGrant, error)

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
