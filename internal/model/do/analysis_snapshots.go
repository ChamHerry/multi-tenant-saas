// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// AnalysisSnapshots is the golang structure of table analysis_snapshots for DAO operations like Where/Data.
type AnalysisSnapshots struct {
	g.Meta           `orm:"table:analysis_snapshots, do:true"`
	Id               any         //
	RepoId           any         //
	CommitHash       any         //
	AnalyzeJobId     any         //
	FilesCount       any         //
	SymbolsCount     any         //
	RelationsCount   any         //
	ProcessesCount   any         //
	CommunitiesCount any         //
	Stats            *gjson.Json //
	CreatedAt        any         //
}
