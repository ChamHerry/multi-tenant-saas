// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// RepoBranches is the golang structure for table repo_branches.
type RepoBranches struct {
	Id         uuid.UUID `json:"id"           orm:"id"           description:""` //
	RepoId     uuid.UUID `json:"repo_id"      orm:"repo_id"      description:""` //
	BranchName string    `json:"branch_name"  orm:"branch_name"  description:""` //
	CommitHash string    `json:"commit_hash"  orm:"commit_hash"  description:""` //
	IsDefault  bool      `json:"is_default"   orm:"is_default"   description:""` //
	LastSeenAt time.Time `json:"last_seen_at" orm:"last_seen_at" description:""` //
}
