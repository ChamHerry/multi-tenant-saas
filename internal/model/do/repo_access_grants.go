// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RepoAccessGrants is the golang structure of table repo_access_grants for DAO operations like Where/Data.
type RepoAccessGrants struct {
	g.Meta            `orm:"table:repo_access_grants, do:true"`
	RepoId            any //
	SubjectType       any //
	SubjectId         any //
	Permission        any //
	Source            any //
	ExternalAccountId any //
	SyncedAt          any //
	ExpiresAt         any //
	CreatedAt         any //
	UpdatedAt         any //
}
