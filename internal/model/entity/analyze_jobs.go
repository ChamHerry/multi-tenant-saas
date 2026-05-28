// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// AnalyzeJobs is the golang structure for table analyze_jobs.
type AnalyzeJobs struct {
	Id                  uuid.UUID   `json:"id"                     orm:"id"                     description:""` //
	RequestedByTenantId uuid.UUID   `json:"requested_by_tenant_id" orm:"requested_by_tenant_id" description:""` //
	RepoId              uuid.UUID   `json:"repo_id"                orm:"repo_id"                description:""` //
	Trigger             string      `json:"trigger"                orm:"trigger"                description:""` //
	Mode                string      `json:"mode"                   orm:"mode"                   description:""` //
	Status              string      `json:"status"                 orm:"status"                 description:""` //
	Priority            int         `json:"priority"               orm:"priority"               description:""` //
	Progress            int         `json:"progress"               orm:"progress"               description:""` //
	WorkerId            string      `json:"worker_id"              orm:"worker_id"              description:""` //
	FromCommit          string      `json:"from_commit"            orm:"from_commit"            description:""` //
	ToCommit            string      `json:"to_commit"              orm:"to_commit"              description:""` //
	ChangedFiles        *gjson.Json `json:"changed_files"          orm:"changed_files"          description:""` //
	Stats               *gjson.Json `json:"stats"                  orm:"stats"                  description:""` //
	ErrorCode           string      `json:"error_code"             orm:"error_code"             description:""` //
	ErrorMsg            string      `json:"error_msg"              orm:"error_msg"              description:""` //
	RetryCount          int         `json:"retry_count"            orm:"retry_count"            description:""` //
	MaxRetries          int         `json:"max_retries"            orm:"max_retries"            description:""` //
	LockedAt            time.Time   `json:"locked_at"              orm:"locked_at"              description:""` //
	StartedAt           time.Time   `json:"started_at"             orm:"started_at"             description:""` //
	FinishedAt          time.Time   `json:"finished_at"            orm:"finished_at"            description:""` //
	CreatedAt           time.Time   `json:"created_at"             orm:"created_at"             description:""` //
	UpdatedAt           time.Time   `json:"updated_at"             orm:"updated_at"             description:""` //
}
