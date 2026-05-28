// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// ApiKeyTenantGrants is the golang structure for table api_key_tenant_grants.
type ApiKeyTenantGrants struct {
	Id              uuid.UUID `json:"id"                 orm:"id"                 description:""` //
	ApiKeyId        uuid.UUID `json:"api_key_id"         orm:"api_key_id"         description:""` //
	TenantId        uuid.UUID `json:"tenant_id"          orm:"tenant_id"          description:""` //
	Scopes          []string  `json:"scopes"             orm:"scopes"             description:""` //
	Status          string    `json:"status"             orm:"status"             description:""` //
	GrantedByUserId uuid.UUID `json:"granted_by_user_id" orm:"granted_by_user_id" description:""` //
	RevokedByUserId uuid.UUID `json:"revoked_by_user_id" orm:"revoked_by_user_id" description:""` //
	CreatedAt       time.Time `json:"created_at"         orm:"created_at"         description:""` //
	UpdatedAt       time.Time `json:"updated_at"         orm:"updated_at"         description:""` //
	RevokedAt       time.Time `json:"revoked_at"         orm:"revoked_at"         description:""` //
}
