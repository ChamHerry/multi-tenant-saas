package service

import (
	"context"
	"time"
)

type CreateTenantInput struct {
	Name            string
	Slug            string
	OwnerUserID     string
	SystemOwnerless bool
	Plan            string
	MaxRepos        int
	MaxSymbols      int
	MaxStorageMB    int
	Metadata        map[string]any
}

type Tenant struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Plan         string    `json:"plan"`
	Status       string    `json:"status"`
	MaxRepos     int       `json:"max_repos"`
	MaxSymbols   int       `json:"max_symbols"`
	MaxStorageMB int       `json:"max_storage_mb"`
	OwnerUserID  string    `json:"owner_user_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdateTenantInput struct {
	Name         string
	Slug         string
	Plan         string
	MaxRepos     int
	MaxSymbols   int
	MaxStorageMB int
	Metadata     map[string]any
}

type ITenantAdmin interface {
	GetTenant(ctx context.Context, tenantSelector string) (*Tenant, error)
	UpdateTenant(ctx context.Context, tenantID string, in UpdateTenantInput) (*Tenant, error)
	SuspendTenant(ctx context.Context, tenantID, actorUserID string) error
	RestoreTenant(ctx context.Context, tenantID, actorUserID string) error
	DeleteTenant(ctx context.Context, tenantID, actorUserID string) error
}

type ITenantProvision interface {
	CreateTenant(ctx context.Context, in CreateTenantInput) (*Tenant, error)
}

var localTenantProvision ITenantProvision
var localTenantAdmin ITenantAdmin

func TenantProvision() ITenantProvision {
	if localTenantProvision == nil {
		panic("implement not found for interface ITenantProvision")
	}
	return localTenantProvision
}

func RegisterTenantProvision(i ITenantProvision) {
	localTenantProvision = i
}

func TenantAdmin() ITenantAdmin {
	if localTenantAdmin == nil {
		panic("implement not found for interface ITenantAdmin")
	}
	return localTenantAdmin
}

func RegisterTenantAdmin(i ITenantAdmin) {
	localTenantAdmin = i
}
