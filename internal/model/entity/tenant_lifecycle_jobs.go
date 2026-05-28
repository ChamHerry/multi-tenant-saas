// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// TenantLifecycleJobs is the golang structure for table tenant_lifecycle_jobs.
type TenantLifecycleJobs struct {
	Id                uuid.UUID   `json:"id"                   orm:"id"                   description:""` //
	TenantId          uuid.UUID   `json:"tenant_id"            orm:"tenant_id"            description:""` //
	Type              string      `json:"type"                 orm:"type"                 description:""` //
	Status            string      `json:"status"               orm:"status"               description:""` //
	RequestedByUserId uuid.UUID   `json:"requested_by_user_id" orm:"requested_by_user_id" description:""` //
	ScheduledAt       time.Time   `json:"scheduled_at"         orm:"scheduled_at"         description:""` //
	StartedAt         time.Time   `json:"started_at"           orm:"started_at"           description:""` //
	FinishedAt        time.Time   `json:"finished_at"          orm:"finished_at"          description:""` //
	ErrorMessage      string      `json:"error_message"        orm:"error_message"        description:""` //
	ArtifactUri       string      `json:"artifact_uri"         orm:"artifact_uri"         description:""` //
	Metadata          *gjson.Json `json:"metadata"             orm:"metadata"             description:""` //
	CreatedAt         time.Time   `json:"created_at"           orm:"created_at"           description:""` //
	UpdatedAt         time.Time   `json:"updated_at"           orm:"updated_at"           description:""` //
}
