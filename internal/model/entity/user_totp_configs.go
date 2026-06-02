// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserTotpConfigs is the golang structure for table user_totp_configs.
type UserTotpConfigs struct {
	Id              uuid.UUID `json:"id"               orm:"id"               description:""` //
	UserId          uuid.UUID `json:"user_id"          orm:"user_id"          description:""` //
	SecretEncrypted string    `json:"secret_encrypted" orm:"secret_encrypted" description:""` //
	Enabled         bool      `json:"enabled"          orm:"enabled"          description:""` //
	EnabledAt       time.Time `json:"enabled_at"       orm:"enabled_at"       description:""` //
	CreatedAt       time.Time `json:"created_at"       orm:"created_at"       description:""` //
	UpdatedAt       time.Time `json:"updated_at"       orm:"updated_at"       description:""` //
}
