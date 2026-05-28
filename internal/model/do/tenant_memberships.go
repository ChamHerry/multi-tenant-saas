// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TenantMemberships is the golang structure of table tenant_memberships for DAO operations like Where/Data.
type TenantMemberships struct {
	g.Meta          `orm:"table:tenant_memberships, do:true"`
	Id              any //
	TenantId        any //
	UserId          any //
	Role            any //
	Status          any //
	InvitedByUserId any //
	JoinedAt        any //
	CreatedAt       any //
	UpdatedAt       any //
	DeletedAt       any //
}
