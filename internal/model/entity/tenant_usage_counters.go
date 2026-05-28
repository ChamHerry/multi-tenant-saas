// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// TenantUsageCounters is the golang structure for table tenant_usage_counters.
type TenantUsageCounters struct {
	TenantId      uuid.UUID `json:"tenant_id"      orm:"tenant_id"      description:""` //
	Metric        string    `json:"metric"         orm:"metric"         description:""` //
	PeriodStart   time.Time `json:"period_start"   orm:"period_start"   description:""` //
	PeriodEnd     time.Time `json:"period_end"     orm:"period_end"     description:""` //
	Used          int64     `json:"used"           orm:"used"           description:""` //
	Reserved      int64     `json:"reserved"       orm:"reserved"       description:""` //
	LimitSnapshot int64     `json:"limit_snapshot" orm:"limit_snapshot" description:""` //
	UpdatedAt     time.Time `json:"updated_at"     orm:"updated_at"     description:""` //
}
