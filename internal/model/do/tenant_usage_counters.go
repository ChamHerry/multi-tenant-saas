// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TenantUsageCounters is the golang structure of table tenant_usage_counters for DAO operations like Where/Data.
type TenantUsageCounters struct {
	g.Meta        `orm:"table:tenant_usage_counters, do:true"`
	TenantId      any //
	Metric        any //
	PeriodStart   any //
	PeriodEnd     any //
	Used          any //
	Reserved      any //
	LimitSnapshot any //
	UpdatedAt     any //
}
