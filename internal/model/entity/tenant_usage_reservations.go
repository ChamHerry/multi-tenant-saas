// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// TenantUsageReservations is the golang structure for table tenant_usage_reservations.
type TenantUsageReservations struct {
	Id           uuid.UUID `json:"id"            orm:"id"            description:""` //
	TenantId     uuid.UUID `json:"tenant_id"     orm:"tenant_id"     description:""` //
	Metric       string    `json:"metric"        orm:"metric"        description:""` //
	Delta        int64     `json:"delta"         orm:"delta"         description:""` //
	Status       string    `json:"status"        orm:"status"        description:""` //
	ResourceType string    `json:"resource_type" orm:"resource_type" description:""` //
	ResourceId   string    `json:"resource_id"   orm:"resource_id"   description:""` //
	ExpiresAt    time.Time `json:"expires_at"    orm:"expires_at"    description:""` //
	CreatedAt    time.Time `json:"created_at"    orm:"created_at"    description:""` //
	UpdatedAt    time.Time `json:"updated_at"    orm:"updated_at"    description:""` //
}
