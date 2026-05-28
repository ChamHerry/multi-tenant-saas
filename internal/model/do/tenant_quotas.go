// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TenantQuotas is the golang structure of table tenant_quotas for DAO operations like Where/Data.
type TenantQuotas struct {
	g.Meta            `orm:"table:tenant_quotas, do:true"`
	TenantId          any //
	MaxDailyRequests  any //
	MaxConcurrentJobs any //
	MaxMembers        any //
	MaxApiKeys        any //
	UpdatedAt         any //
}
