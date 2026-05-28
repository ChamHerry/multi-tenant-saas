package service

import (
	"context"
	"time"
)

type AddTenantMemberInput struct {
	TenantID        string
	UserID          string
	Role            string
	Status          string
	InvitedByUserID string
}

type UpdateTenantMemberInput struct {
	Role   string
	Status string
}

type TenantMembership struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	UserID          string     `json:"user_id"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	InvitedByUserID string     `json:"invited_by_user_id,omitempty"`
	JoinedAt        *time.Time `json:"joined_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TenantMembershipWithTenant struct {
	TenantMembership
	TenantName   string `json:"tenant_name"`
	TenantSlug   string `json:"tenant_slug"`
	TenantStatus string `json:"tenant_status"`
}

type TenantMembershipWithUser struct {
	TenantMembership
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	UserStatus  string `json:"user_status"`
}

type TenantContext struct {
	TenantID      string       `json:"tenant_id"`
	UserID        string       `json:"user_id"`
	Role          string       `json:"role"`
	TenantSlug    string       `json:"tenant_slug"`
	AuthType      string       `json:"auth_type"`
	Scopes        []string     `json:"scopes,omitempty"`
	Permissions   []Permission `json:"permissions,omitempty"`
	APIKeyID      string       `json:"api_key_id,omitempty"`
	APIKeyGrantID string       `json:"api_key_grant_id,omitempty"`
	RequestID     string       `json:"request_id"`
}

type ITenantMembership interface {
	AddMember(ctx context.Context, in AddTenantMemberInput) (*TenantMembership, error)
	RemoveMember(ctx context.Context, tenantID, userID string) error
	ChangeRole(ctx context.Context, tenantID, userID, role string) error
	UpdateMember(ctx context.Context, tenantID, userID string, in UpdateTenantMemberInput) (*TenantMembership, error)
	ListUserTenants(ctx context.Context, userID string) ([]TenantMembershipWithTenant, error)
	ListTenantMembers(ctx context.Context, tenantID string) ([]TenantMembershipWithUser, error)
	ResolveTenantContext(ctx context.Context, userID, tenantSelector string) (*TenantContext, error)
}

var localTenantMembership ITenantMembership

func TenantMembershipService() ITenantMembership {
	if localTenantMembership == nil {
		panic("implement not found for interface ITenantMembership")
	}
	return localTenantMembership
}

func RegisterTenantMembership(i ITenantMembership) {
	localTenantMembership = i
}
