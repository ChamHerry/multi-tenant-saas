// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// AnalysisSnapshots is the golang structure for table analysis_snapshots.
type AnalysisSnapshots struct {
	Id               uuid.UUID   `json:"id"                orm:"id"                description:""` //
	RepoId           uuid.UUID   `json:"repo_id"           orm:"repo_id"           description:""` //
	CommitHash       string      `json:"commit_hash"       orm:"commit_hash"       description:""` //
	AnalyzeJobId     uuid.UUID   `json:"analyze_job_id"    orm:"analyze_job_id"    description:""` //
	FilesCount       int         `json:"files_count"       orm:"files_count"       description:""` //
	SymbolsCount     int         `json:"symbols_count"     orm:"symbols_count"     description:""` //
	RelationsCount   int         `json:"relations_count"   orm:"relations_count"   description:""` //
	ProcessesCount   int         `json:"processes_count"   orm:"processes_count"   description:""` //
	CommunitiesCount int         `json:"communities_count" orm:"communities_count" description:""` //
	Stats            *gjson.Json `json:"stats"             orm:"stats"             description:""` //
	CreatedAt        time.Time   `json:"created_at"        orm:"created_at"        description:""` //
}
