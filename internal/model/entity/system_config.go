// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SystemConfig is the golang structure for table system_config.
type SystemConfig struct {
	Key         string    `json:"key"          orm:"key"          description:""` //
	Value       string    `json:"value"        orm:"value"        description:""` //
	ValueType   string    `json:"valueType"    orm:"value_type"   description:""` //
	Description string    `json:"description"  orm:"description"  description:""` //
	IsEncrypted bool      `json:"isEncrypted"  orm:"is_encrypted" description:""` //
	CreatedAt   time.Time `json:"createdAt"    orm:"created_at"   description:""` //
	UpdatedAt   time.Time `json:"updatedAt"    orm:"updated_at"   description:""` //
}
