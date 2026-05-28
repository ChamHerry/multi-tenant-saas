// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuthSessions is the golang structure for table auth_sessions.
type AuthSessions struct {
	Id            uuid.UUID `json:"id"              orm:"id"              description:""` //
	UserId        uuid.UUID `json:"user_id"         orm:"user_id"         description:""` //
	SecretHash    string    `json:"secret_hash"     orm:"secret_hash"     description:""` //
	CsrfHash      string    `json:"csrf_hash"       orm:"csrf_hash"       description:""` //
	UserAgent     string    `json:"user_agent"      orm:"user_agent"      description:""` //
	Ip            string    `json:"ip"              orm:"ip"              description:""` //
	LastUsedAt    time.Time `json:"last_used_at"    orm:"last_used_at"    description:""` //
	ExpiresAt     time.Time `json:"expires_at"      orm:"expires_at"      description:""` //
	IdleExpiresAt time.Time `json:"idle_expires_at" orm:"idle_expires_at" description:""` //
	RevokedAt     time.Time `json:"revoked_at"      orm:"revoked_at"      description:""` //
	RevokeReason  string    `json:"revoke_reason"   orm:"revoke_reason"   description:""` //
	CreatedAt     time.Time `json:"created_at"      orm:"created_at"      description:""` //
	UpdatedAt     time.Time `json:"updated_at"      orm:"updated_at"      description:""` //
}
