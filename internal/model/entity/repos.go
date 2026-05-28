// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// Repos is the golang structure for table repos.
type Repos struct {
	Id                  uuid.UUID   `json:"id"                    orm:"id"                    description:""` //
	Provider            string      `json:"provider"              orm:"provider"              description:""` //
	RemoteUrl           string      `json:"remote_url"            orm:"remote_url"            description:""` //
	CloneUrl            string      `json:"clone_url"             orm:"clone_url"             description:""` //
	WebUrl              string      `json:"web_url"               orm:"web_url"               description:""` //
	Namespace           string      `json:"namespace"             orm:"namespace"             description:""` //
	Name                string      `json:"name"                  orm:"name"                  description:""` //
	DefaultBranch       string      `json:"default_branch"        orm:"default_branch"        description:""` //
	LastCommit          string      `json:"last_commit"           orm:"last_commit"           description:""` //
	AnalyzeStatus       string      `json:"analyze_status"        orm:"analyze_status"        description:""` //
	WebhookId           string      `json:"webhook_id"            orm:"webhook_id"            description:""` //
	Stats               *gjson.Json `json:"stats"                 orm:"stats"                 description:""` //
	LastAnalyzedAt      time.Time   `json:"last_analyzed_at"      orm:"last_analyzed_at"      description:""` //
	CreatedAt           time.Time   `json:"created_at"            orm:"created_at"            description:""` //
	UpdatedAt           time.Time   `json:"updated_at"            orm:"updated_at"            description:""` //
	DeletedAt           time.Time   `json:"deleted_at"            orm:"deleted_at"            description:""` //
	CodeHostUrl         string      `json:"code_host_url"         orm:"code_host_url"         description:""` //
	ExternalId          string      `json:"external_id"           orm:"external_id"           description:""` //
	ExternalNodeId      string      `json:"external_node_id"      orm:"external_node_id"      description:""` //
	OwnerName           string      `json:"owner_name"            orm:"owner_name"            description:""` //
	FullName            string      `json:"full_name"             orm:"full_name"             description:""` //
	NormalizedRemoteKey string      `json:"normalized_remote_key" orm:"normalized_remote_key" description:""` //
	Visibility          string      `json:"visibility"            orm:"visibility"            description:""` //
	IsFork              bool        `json:"is_fork"               orm:"is_fork"               description:""` //
	IsArchived          bool        `json:"is_archived"           orm:"is_archived"           description:""` //
	IndexedCommitHash   string      `json:"indexed_commit_hash"   orm:"indexed_commit_hash"   description:""` //
	PermissionSyncedAt  time.Time   `json:"permission_synced_at"  orm:"permission_synced_at"  description:""` //
	Metadata            *gjson.Json `json:"metadata"              orm:"metadata"              description:""` //
}
