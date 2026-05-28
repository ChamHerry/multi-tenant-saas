// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
)

// PlanEntitlements is the golang structure for table plan_entitlements.
type PlanEntitlements struct {
	Plan       string      `json:"plan"        orm:"plan"        description:""` //
	FeatureKey string      `json:"feature_key" orm:"feature_key" description:""` //
	Enabled    bool        `json:"enabled"     orm:"enabled"     description:""` //
	LimitValue int64       `json:"limit_value" orm:"limit_value" description:""` //
	Metadata   *gjson.Json `json:"metadata"    orm:"metadata"    description:""` //
	CreatedAt  time.Time   `json:"created_at"  orm:"created_at"  description:""` //
	UpdatedAt  time.Time   `json:"updated_at"  orm:"updated_at"  description:""` //
}
