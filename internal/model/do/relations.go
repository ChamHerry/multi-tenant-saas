// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// Relations is the golang structure of table relations for DAO operations like Where/Data.
type Relations struct {
	g.Meta       `orm:"table:relations, do:true"`
	Id           any         //
	RepoId       any         //
	SourceUid    any         //
	TargetUid    any         //
	RelationType any         //
	Confidence   any         //
	Reason       any         //
	Properties   *gjson.Json //
	CreatedAt    any         //
}
