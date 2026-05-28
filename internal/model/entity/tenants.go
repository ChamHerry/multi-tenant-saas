// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// Tenants is the golang structure for table tenants.
type Tenants struct {
	Id        uuid.UUID   `json:"id"             orm:"id"             description:""` //
	Name      string      `json:"name"           orm:"name"           description:""` //
	Slug      string      `json:"slug"           orm:"slug"           description:""` //
	Plan      string      `json:"plan"           orm:"plan"           description:""` //
	Status    string      `json:"status"         orm:"status"         description:""` //
	Metadata  *gjson.Json `json:"metadata"       orm:"metadata"       description:""` //
	CreatedAt time.Time   `json:"created_at"     orm:"created_at"     description:""` //
	UpdatedAt time.Time   `json:"updated_at"     orm:"updated_at"     description:""` //
	DeletedAt time.Time   `json:"deleted_at"     orm:"deleted_at"     description:""` //
}
