package service

import (
	"context"
	"time"
)

type CreateTenantInvitationInput struct {
	TenantID        string
	InviteeEmail    string
	Role            string
	Message         string
	InvitedByUserID string
	ExpiresIn       time.Duration
}

type TenantInvitationFilter struct {
	Status string
	Limit  int
	Offset int
}

type TenantInvitation struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	TenantName       string     `json:"tenant_name,omitempty"`
	TenantSlug       string     `json:"tenant_slug,omitempty"`
	InviteeEmail     string     `json:"invitee_email"`
	InviteeUserID    string     `json:"invitee_user_id,omitempty"`
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	InvitedByUserID  string     `json:"invited_by_user_id,omitempty"`
	AcceptedByUserID string     `json:"accepted_by_user_id,omitempty"`
	Message          string     `json:"message"`
	ExpiresAt        time.Time  `json:"expires_at"`
	AcceptedAt       *time.Time `json:"accepted_at,omitempty"`
	DeclinedAt       *time.Time `json:"declined_at,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	ResentAt         *time.Time `json:"resent_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreatedTenantInvitation struct {
	Invitation TenantInvitation `json:"invitation"`
	Token      string           `json:"token,omitempty"`
	AcceptURL  string           `json:"accept_url,omitempty"`
}

type TenantInvitationList struct {
	Invitations []TenantInvitation `json:"invitations"`
	Total       int                `json:"total"`
}

type ITenantInvitation interface {
	Create(ctx context.Context, in CreateTenantInvitationInput) (*CreatedTenantInvitation, error)
	ListTenantInvitations(ctx context.Context, tenantID string, filter TenantInvitationFilter) (*TenantInvitationList, error)
	ListMyInvitations(ctx context.Context, userID string, email string, filter TenantInvitationFilter) (*TenantInvitationList, error)
	Accept(ctx context.Context, token string, userID string) (*TenantMembership, error)
	Decline(ctx context.Context, invitationID string, userID string) error
	Revoke(ctx context.Context, tenantID string, invitationID string, actorUserID string) error
	Resend(ctx context.Context, tenantID string, invitationID string, actorUserID string) (*CreatedTenantInvitation, error)
	ExpirePending(ctx context.Context, now time.Time, limit int) (int, error)
}

var localTenantInvitation ITenantInvitation

func TenantInvitationService() ITenantInvitation {
	if localTenantInvitation == nil {
		panic("implement not found for interface ITenantInvitation")
	}
	return localTenantInvitation
}

func RegisterTenantInvitation(i ITenantInvitation) {
	localTenantInvitation = i
}
