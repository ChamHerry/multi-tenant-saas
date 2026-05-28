// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// Files is the golang structure for table files.
type Files struct {
	Id          uuid.UUID `json:"id"           orm:"id"           description:""` //
	RepoId      uuid.UUID `json:"repo_id"      orm:"repo_id"      description:""` //
	CommitHash  string    `json:"commit_hash"  orm:"commit_hash"  description:""` //
	Path        string    `json:"path"         orm:"path"         description:""` //
	Language    string    `json:"language"     orm:"language"     description:""` //
	SizeBytes   int64     `json:"size_bytes"   orm:"size_bytes"   description:""` //
	ContentHash string    `json:"content_hash" orm:"content_hash" description:""` //
	IsDeleted   bool      `json:"is_deleted"   orm:"is_deleted"   description:""` //
	CreatedAt   time.Time `json:"created_at"   orm:"created_at"   description:""` //
	UpdatedAt   time.Time `json:"updated_at"   orm:"updated_at"   description:""` //
}
