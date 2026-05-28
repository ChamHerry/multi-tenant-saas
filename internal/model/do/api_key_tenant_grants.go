// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ApiKeyTenantGrants is the golang structure of table api_key_tenant_grants for DAO operations like Where/Data.
type ApiKeyTenantGrants struct {
	g.Meta          `orm:"table:api_key_tenant_grants, do:true"`
	Id              any      //
	ApiKeyId        any      //
	TenantId        any      //
	Scopes          []string //
	Status          any      //
	GrantedByUserId any      //
	RevokedByUserId any      //
	CreatedAt       any      //
	UpdatedAt       any      //
	RevokedAt       any      //
}
