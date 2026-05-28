// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// PlanEntitlements is the golang structure of table plan_entitlements for DAO operations like Where/Data.
type PlanEntitlements struct {
	g.Meta     `orm:"table:plan_entitlements, do:true"`
	Plan       any         //
	FeatureKey any         //
	Enabled    any         //
	LimitValue any         //
	Metadata   *gjson.Json //
	CreatedAt  any         //
	UpdatedAt  any         //
}
