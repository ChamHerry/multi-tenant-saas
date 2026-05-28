// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RepoBranches is the golang structure of table repo_branches for DAO operations like Where/Data.
type RepoBranches struct {
	g.Meta     `orm:"table:repo_branches, do:true"`
	Id         any //
	RepoId     any //
	BranchName any //
	CommitHash any //
	IsDefault  any //
	LastSeenAt any //
}
