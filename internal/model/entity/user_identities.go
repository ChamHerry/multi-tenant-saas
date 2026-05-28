// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// UserIdentities is the golang structure for table user_identities.
type UserIdentities struct {
	Id            uuid.UUID   `json:"id"             orm:"id"             description:""` //
	UserId        uuid.UUID   `json:"user_id"        orm:"user_id"        description:""` //
	Provider      string      `json:"provider"       orm:"provider"       description:""` //
	AuthId        string      `json:"auth_id"        orm:"auth_id"        description:""` //
	Email         string      `json:"email"          orm:"email"          description:""` //
	EmailVerified bool        `json:"email_verified" orm:"email_verified" description:""` //
	RawProfile    *gjson.Json `json:"raw_profile"    orm:"raw_profile"    description:""` //
	LastLoginAt   time.Time   `json:"last_login_at"  orm:"last_login_at"  description:""` //
	CreatedAt     time.Time   `json:"created_at"     orm:"created_at"     description:""` //
	UpdatedAt     time.Time   `json:"updated_at"     orm:"updated_at"     description:""` //
}
