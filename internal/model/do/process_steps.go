// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProcessSteps is the golang structure of table process_steps for DAO operations like Where/Data.
type ProcessSteps struct {
	g.Meta         `orm:"table:process_steps, do:true"`
	RepoId         any //
	ProcessId      any //
	StepIndex      any //
	SymbolUid      any //
	RelationToNext any //
}
