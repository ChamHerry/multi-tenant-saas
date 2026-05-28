// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// SymbolSearch is the golang structure of table symbol_search for DAO operations like Where/Data.
type SymbolSearch struct {
	g.Meta             `orm:"table:symbol_search, do:true"`
	Id                 any         //
	RepoId             any         //
	SymbolUid          any         //
	Name               any         //
	Kind               any         //
	FilePath           any         //
	Content            any         //
	Summary            any         //
	CommunityId        any         //
	ProcessId          any         //
	NameTsv            any         //
	ContentTsv         any         //
	Embedding          any         //
	EmbeddingModel     any         //
	EmbeddingDimension any         //
	SearchBoost        any         //
	Metadata           *gjson.Json //
	CreatedAt          any         //
	UpdatedAt          any         //
}
