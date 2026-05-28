// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// Subscriptions is the golang structure of table subscriptions for DAO operations like Where/Data.
type Subscriptions struct {
	g.Meta    `orm:"table:subscriptions, do:true"`
	Id        any         //
	TenantId  any         //
	Plan      any         //
	Status    any         //
	StartedAt any         //
	EndsAt    any         //
	Metadata  *gjson.Json //
}
