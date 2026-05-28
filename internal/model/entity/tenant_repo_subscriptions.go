// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// TenantRepoSubscriptions is the golang structure for table tenant_repo_subscriptions.
type TenantRepoSubscriptions struct {
	TenantId      uuid.UUID `json:"tenant_id"        orm:"tenant_id"        description:""` //
	RepoId        uuid.UUID `json:"repo_id"          orm:"repo_id"          description:""` //
	AddedByUserId uuid.UUID `json:"added_by_user_id" orm:"added_by_user_id" description:""` //
	CredentialId  uuid.UUID `json:"credential_id"    orm:"credential_id"    description:""` //
	Purpose       string    `json:"purpose"          orm:"purpose"          description:""` //
	Status        string    `json:"status"           orm:"status"           description:""` //
	CreatedAt     time.Time `json:"created_at"       orm:"created_at"       description:""` //
	UpdatedAt     time.Time `json:"updated_at"       orm:"updated_at"       description:""` //
	DeletedAt     time.Time `json:"deleted_at"       orm:"deleted_at"       description:""` //
}
