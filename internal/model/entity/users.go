// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// Users is the golang structure for table users.
type Users struct {
	Id          uuid.UUID   `json:"id"            orm:"id"            description:""` //
	Email       string      `json:"email"         orm:"email"         description:""` //
	DisplayName string      `json:"display_name"  orm:"display_name"  description:""` //
	Status      string      `json:"status"        orm:"status"        description:""` //
	LastLoginAt time.Time   `json:"last_login_at" orm:"last_login_at" description:""` //
	CreatedAt   time.Time   `json:"created_at"    orm:"created_at"    description:""` //
	UpdatedAt   time.Time   `json:"updated_at"    orm:"updated_at"    description:""` //
	DeletedAt   time.Time   `json:"deleted_at"    orm:"deleted_at"    description:""` //
	AvatarUrl   string      `json:"avatar_url"    orm:"avatar_url"    description:""` //
	Metadata    *gjson.Json `json:"metadata"      orm:"metadata"      description:""` //
}
