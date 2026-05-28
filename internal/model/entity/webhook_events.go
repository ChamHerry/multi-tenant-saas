// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// WebhookEvents is the golang structure for table webhook_events.
type WebhookEvents struct {
	Id                 uuid.UUID   `json:"id"                    orm:"id"                    description:""` //
	ReceivedByTenantId uuid.UUID   `json:"received_by_tenant_id" orm:"received_by_tenant_id" description:""` //
	RepoId             uuid.UUID   `json:"repo_id"               orm:"repo_id"               description:""` //
	Provider           string      `json:"provider"              orm:"provider"              description:""` //
	EventType          string      `json:"event_type"            orm:"event_type"            description:""` //
	DeliveryId         string      `json:"delivery_id"           orm:"delivery_id"           description:""` //
	PayloadHash        string      `json:"payload_hash"          orm:"payload_hash"          description:""` //
	Payload            *gjson.Json `json:"payload"               orm:"payload"               description:""` //
	Status             string      `json:"status"                orm:"status"                description:""` //
	AnalyzeJobId       uuid.UUID   `json:"analyze_job_id"        orm:"analyze_job_id"        description:""` //
	ReceivedAt         time.Time   `json:"received_at"           orm:"received_at"           description:""` //
	ProcessedAt        time.Time   `json:"processed_at"          orm:"processed_at"          description:""` //
	ErrorMsg           string      `json:"error_msg"             orm:"error_msg"             description:""` //
}
