// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// AuditLogs is the golang structure of table audit_logs for DAO operations like Where/Data.
type AuditLogs struct {
	g.Meta       `orm:"table:audit_logs, do:true"`
	Id           any         //
	TenantId     any         //
	UserId       any         //
	Action       any         //
	ResourceType any         //
	ResourceId   any         //
	Ip           any         //
	UserAgent    any         //
	Metadata     *gjson.Json //
	CreatedAt    any         //
}
