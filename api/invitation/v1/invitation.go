package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

type ListTenantReq struct {
	g.Meta `path:"/tenants/{tenant}/invitations" tags:"Invitation" method:"get" summary:"List tenant invitations"`
	Tenant string `v:"required"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type CreateReq struct {
	g.Meta       `path:"/tenants/{tenant}/invitations" tags:"Invitation" method:"post" summary:"Create tenant invitation"`
	Tenant       string `v:"required"`
	InviteeEmail string `json:"invitee_email" v:"required"`
	Role         string `json:"role" v:"required"`
	Message      string `json:"message"`
	ExpiresHours int    `json:"expires_hours"`
}

type ResendReq struct {
	g.Meta     `path:"/tenants/{tenant}/invitations/{invitation}/resend" tags:"Invitation" method:"post" summary:"Resend tenant invitation"`
	Tenant     string `v:"required"`
	Invitation string `v:"required"`
}

type RevokeReq struct {
	g.Meta     `path:"/tenants/{tenant}/invitations/{invitation}/revoke" tags:"Invitation" method:"post" summary:"Revoke tenant invitation"`
	Tenant     string `v:"required"`
	Invitation string `v:"required"`
}

type ListMineReq struct {
	g.Meta `path:"/me/invitations" tags:"Invitation" method:"get" summary:"List my invitations"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type AcceptReq struct {
	g.Meta `path:"/invitations/accept" tags:"Invitation" method:"post" summary:"Accept tenant invitation"`
	Token  string `json:"token" v:"required"`
}

type DeclineReq struct {
	g.Meta     `path:"/me/invitations/{invitation}/decline" tags:"Invitation" method:"post" summary:"Decline tenant invitation"`
	Invitation string `v:"required"`
}

type ListRes struct {
	Invitations []service.TenantInvitation `json:"invitations"`
	Total       int                        `json:"total"`
}

type CreateRes struct {
	Invitation service.TenantInvitation `json:"invitation"`
	Token      string                   `json:"token,omitempty"`
	AcceptURL  string                   `json:"accept_url,omitempty"`
}

type AcceptRes struct {
	Member *service.TenantMembership `json:"member"`
}

type ActionRes struct {
	OK bool `json:"ok"`
}
