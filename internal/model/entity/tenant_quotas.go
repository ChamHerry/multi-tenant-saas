// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// TenantQuotas is the golang structure for table tenant_quotas.
type TenantQuotas struct {
	TenantId          uuid.UUID `json:"tenant_id"           orm:"tenant_id"           description:""` //
	MaxDailyRequests  int       `json:"max_daily_requests"  orm:"max_daily_requests"  description:""` //
	MaxConcurrentJobs int       `json:"max_concurrent_jobs" orm:"max_concurrent_jobs" description:""` //
	MaxMembers        int       `json:"max_members"         orm:"max_members"         description:""` //
	MaxApiKeys        int       `json:"max_api_keys"        orm:"max_api_keys"        description:""` //
	UpdatedAt         time.Time `json:"updated_at"          orm:"updated_at"          description:""` //
}
