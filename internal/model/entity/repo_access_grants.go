// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// RepoAccessGrants is the golang structure for table repo_access_grants.
type RepoAccessGrants struct {
	RepoId            uuid.UUID `json:"repo_id"             orm:"repo_id"             description:""` //
	SubjectType       string    `json:"subject_type"        orm:"subject_type"        description:""` //
	SubjectId         uuid.UUID `json:"subject_id"          orm:"subject_id"          description:""` //
	Permission        string    `json:"permission"          orm:"permission"          description:""` //
	Source            string    `json:"source"              orm:"source"              description:""` //
	ExternalAccountId string    `json:"external_account_id" orm:"external_account_id" description:""` //
	SyncedAt          time.Time `json:"synced_at"           orm:"synced_at"           description:""` //
	ExpiresAt         time.Time `json:"expires_at"          orm:"expires_at"          description:""` //
	CreatedAt         time.Time `json:"created_at"          orm:"created_at"          description:""` //
	UpdatedAt         time.Time `json:"updated_at"          orm:"updated_at"          description:""` //
}
