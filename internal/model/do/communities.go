// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// Communities is the golang structure of table communities for DAO operations like Where/Data.
type Communities struct {
	g.Meta      `orm:"table:communities, do:true"`
	RepoId      any         //
	Id          any         //
	Label       any         //
	Cohesion    any         //
	SymbolCount any         //
	Keywords    []string    //
	Description any         //
	Metadata    *gjson.Json //
	CreatedAt   any         //
}
