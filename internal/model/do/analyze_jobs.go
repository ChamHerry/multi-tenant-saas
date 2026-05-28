// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// AnalyzeJobs is the golang structure of table analyze_jobs for DAO operations like Where/Data.
type AnalyzeJobs struct {
	g.Meta              `orm:"table:analyze_jobs, do:true"`
	Id                  any         //
	RequestedByTenantId any         //
	RepoId              any         //
	Trigger             any         //
	Mode                any         //
	Status              any         //
	Priority            any         //
	Progress            any         //
	WorkerId            any         //
	FromCommit          any         //
	ToCommit            any         //
	ChangedFiles        *gjson.Json //
	Stats               *gjson.Json //
	ErrorCode           any         //
	ErrorMsg            any         //
	RetryCount          any         //
	MaxRetries          any         //
	LockedAt            any         //
	StartedAt           any         //
	FinishedAt          any         //
	CreatedAt           any         //
	UpdatedAt           any         //
}
