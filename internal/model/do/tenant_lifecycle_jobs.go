// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantLifecycleJobs is the golang structure of table tenant_lifecycle_jobs for DAO operations like Where/Data.
type TenantLifecycleJobs struct {
	g.Meta            `orm:"table:tenant_lifecycle_jobs, do:true"`
	Id                any         //
	TenantId          any         //
	Type              any         //
	Status            any         //
	RequestedByUserId any         //
	ScheduledAt       any         //
	StartedAt         any         //
	FinishedAt        any         //
	ErrorMessage      any         //
	ArtifactUri       any         //
	Metadata          *gjson.Json //
	CreatedAt         any         //
	UpdatedAt         any         //
}
