// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserPasswordCredentials is the golang structure for table user_password_credentials.
type UserPasswordCredentials struct {
	UserId            uuid.UUID `json:"user_id"             orm:"user_id"             description:""` //
	PasswordHash      string    `json:"password_hash"       orm:"password_hash"       description:""` //
	HashAlg           string    `json:"hash_alg"            orm:"hash_alg"            description:""` //
	HashCost          int       `json:"hash_cost"           orm:"hash_cost"           description:""` //
	PasswordChangedAt time.Time `json:"password_changed_at" orm:"password_changed_at" description:""` //
	CreatedAt         time.Time `json:"created_at"          orm:"created_at"          description:""` //
	UpdatedAt         time.Time `json:"updated_at"          orm:"updated_at"          description:""` //
}
