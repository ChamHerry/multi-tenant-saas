// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// AuditLogs is the golang structure for table audit_logs.
type AuditLogs struct {
	Id           uuid.UUID   `json:"id"            orm:"id"            description:""` //
	TenantId     uuid.UUID   `json:"tenant_id"     orm:"tenant_id"     description:""` //
	UserId       uuid.UUID   `json:"user_id"       orm:"user_id"       description:""` //
	Action       string      `json:"action"        orm:"action"        description:""` //
	ResourceType string      `json:"resource_type" orm:"resource_type" description:""` //
	ResourceId   string      `json:"resource_id"   orm:"resource_id"   description:""` //
	Ip           string      `json:"ip"            orm:"ip"            description:""` //
	UserAgent    string      `json:"user_agent"    orm:"user_agent"    description:""` //
	Metadata     *gjson.Json `json:"metadata"      orm:"metadata"      description:""` //
	CreatedAt    time.Time   `json:"created_at"    orm:"created_at"    description:""` //
}
