// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// Symbols is the golang structure of table symbols for DAO operations like Where/Data.
type Symbols struct {
	g.Meta         `orm:"table:symbols, do:true"`
	RepoId         any         //
	Uid            any         //
	FileId         any         //
	Name           any         //
	Kind           any         //
	FilePath       any         //
	StartLine      any         //
	EndLine        any         //
	Content        any         //
	ParameterCount any         //
	ReturnType     any         //
	Properties     *gjson.Json //
	ContentHash    any         //
	CreatedAt      any         //
	UpdatedAt      any         //
}
