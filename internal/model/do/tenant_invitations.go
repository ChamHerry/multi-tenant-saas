// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantInvitations is the golang structure of table tenant_invitations for DAO operations like Where/Data.
type TenantInvitations struct {
	g.Meta           `orm:"table:tenant_invitations, do:true"`
	Id               any         //
	TenantId         any         //
	InviteeEmail     any         //
	InviteeUserId    any         //
	Role             any         //
	Status           any         //
	TokenHash        any         //
	InvitedByUserId  any         //
	AcceptedByUserId any         //
	Message          any         //
	ExpiresAt        any         //
	AcceptedAt       any         //
	DeclinedAt       any         //
	RevokedAt        any         //
	ResentAt         any         //
	Metadata         *gjson.Json //
	CreatedAt        any         //
	UpdatedAt        any         //
}
