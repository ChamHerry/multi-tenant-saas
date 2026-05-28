// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// GitCredentials is the golang structure for table git_credentials.
type GitCredentials struct {
	Id                 uuid.UUID `json:"id"                   orm:"id"                   description:""` //
	TenantId           uuid.UUID `json:"tenant_id"            orm:"tenant_id"            description:""` //
	OwnerUserId        uuid.UUID `json:"owner_user_id"        orm:"owner_user_id"        description:""` //
	Provider           string    `json:"provider"             orm:"provider"             description:""` //
	BaseUrl            string    `json:"base_url"             orm:"base_url"             description:""` //
	AccessTokenCipher  []byte    `json:"access_token_cipher"  orm:"access_token_cipher"  description:""` //
	RefreshTokenCipher []byte    `json:"refresh_token_cipher" orm:"refresh_token_cipher" description:""` //
	TokenExpiresAt     time.Time `json:"token_expires_at"     orm:"token_expires_at"     description:""` //
	Scopes             []string  `json:"scopes"               orm:"scopes"               description:""` //
	Status             string    `json:"status"               orm:"status"               description:""` //
	CreatedAt          time.Time `json:"created_at"           orm:"created_at"           description:""` //
	UpdatedAt          time.Time `json:"updated_at"           orm:"updated_at"           description:""` //
	DeletedAt          time.Time `json:"deleted_at"           orm:"deleted_at"           description:""` //
}
