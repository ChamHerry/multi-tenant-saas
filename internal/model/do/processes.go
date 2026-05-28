// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// Processes is the golang structure of table processes for DAO operations like Where/Data.
type Processes struct {
	g.Meta        `orm:"table:processes, do:true"`
	RepoId        any         //
	Id            any         //
	Label         any         //
	ProcessType   any         //
	StepCount     any         //
	Communities   []string    //
	EntryPointUid any         //
	TerminalUid   any         //
	Summary       any         //
	Metadata      *gjson.Json //
	CreatedAt     any         //
	UpdatedAt     any         //
}
