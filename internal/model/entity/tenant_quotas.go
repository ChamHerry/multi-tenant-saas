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
	MaxRepos          int       `json:"max_repos"           orm:"max_repos"           description:""` //
	MaxSymbols        int       `json:"max_symbols"         orm:"max_symbols"         description:""` //
	MaxStorageMb      int       `json:"max_storage_mb"      orm:"max_storage_mb"      description:""` //
	MaxDailyRequests  int       `json:"max_daily_requests"  orm:"max_daily_requests"  description:""` //
	MaxConcurrentJobs int       `json:"max_concurrent_jobs" orm:"max_concurrent_jobs" description:""` //
	CurrentRepos      int       `json:"current_repos"       orm:"current_repos"       description:""` //
	CurrentSymbols    int64     `json:"current_symbols"     orm:"current_symbols"     description:""` //
	CurrentStorageMb  int64     `json:"current_storage_mb"  orm:"current_storage_mb"  description:""` //
	UpdatedAt         time.Time `json:"updated_at"          orm:"updated_at"          description:""` //
}
