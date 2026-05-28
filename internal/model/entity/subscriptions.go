// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// Subscriptions is the golang structure for table subscriptions.
type Subscriptions struct {
	Id        uuid.UUID   `json:"id"         orm:"id"         description:""` //
	TenantId  uuid.UUID   `json:"tenant_id"  orm:"tenant_id"  description:""` //
	Plan      string      `json:"plan"       orm:"plan"       description:""` //
	Status    string      `json:"status"     orm:"status"     description:""` //
	StartedAt time.Time   `json:"started_at" orm:"started_at" description:""` //
	EndsAt    time.Time   `json:"ends_at"    orm:"ends_at"    description:""` //
	Metadata  *gjson.Json `json:"metadata"   orm:"metadata"   description:""` //
}
