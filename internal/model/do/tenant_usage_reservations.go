// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TenantUsageReservations is the golang structure of table tenant_usage_reservations for DAO operations like Where/Data.
type TenantUsageReservations struct {
	g.Meta       `orm:"table:tenant_usage_reservations, do:true"`
	Id           any //
	TenantId     any //
	Metric       any //
	Delta        any //
	Status       any //
	ResourceType any //
	ResourceId   any //
	ExpiresAt    any //
	CreatedAt    any //
	UpdatedAt    any //
}
