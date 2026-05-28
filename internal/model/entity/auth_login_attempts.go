// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuthLoginAttempts is the golang structure for table auth_login_attempts.
type AuthLoginAttempts struct {
	Id            uuid.UUID `json:"id"              orm:"id"              description:""` //
	LoginKey      string    `json:"login_key"       orm:"login_key"       description:""` //
	Ip            string    `json:"ip"              orm:"ip"              description:""` //
	FailedCount   int       `json:"failed_count"    orm:"failed_count"    description:""` //
	LockedUntil   time.Time `json:"locked_until"    orm:"locked_until"    description:""` //
	LastFailedAt  time.Time `json:"last_failed_at"  orm:"last_failed_at"  description:""` //
	LastSuccessAt time.Time `json:"last_success_at" orm:"last_success_at" description:""` //
	CreatedAt     time.Time `json:"created_at"      orm:"created_at"      description:""` //
	UpdatedAt     time.Time `json:"updated_at"      orm:"updated_at"      description:""` //
}
