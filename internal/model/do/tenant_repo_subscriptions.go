// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TenantRepoSubscriptions is the golang structure of table tenant_repo_subscriptions for DAO operations like Where/Data.
type TenantRepoSubscriptions struct {
	g.Meta        `orm:"table:tenant_repo_subscriptions, do:true"`
	TenantId      any //
	RepoId        any //
	AddedByUserId any //
	CredentialId  any //
	Purpose       any //
	Status        any //
	CreatedAt     any //
	UpdatedAt     any //
	DeletedAt     any //
}
