// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// ApiKeys is the golang structure for table api_keys.
type ApiKeys struct {
	Id              uuid.UUID `json:"id"                 orm:"id"                 description:""` //
	TenantId        uuid.UUID `json:"tenant_id"          orm:"tenant_id"          description:""` //
	UserId          uuid.UUID `json:"user_id"            orm:"user_id"            description:""` //
	Name            string    `json:"name"               orm:"name"               description:""` //
	KeyHash         string    `json:"key_hash"           orm:"key_hash"           description:""` //
	Scopes          []string  `json:"scopes"             orm:"scopes"             description:""` //
	LastUsedAt      time.Time `json:"last_used_at"       orm:"last_used_at"       description:""` //
	ExpiresAt       time.Time `json:"expires_at"         orm:"expires_at"         description:""` //
	CreatedAt       time.Time `json:"created_at"         orm:"created_at"         description:""` //
	RevokedAt       time.Time `json:"revoked_at"         orm:"revoked_at"         description:""` //
	KeyPrefix       string    `json:"key_prefix"         orm:"key_prefix"         description:""` //
	CreatedByUserId uuid.UUID `json:"created_by_user_id" orm:"created_by_user_id" description:""` //
}
