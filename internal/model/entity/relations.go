// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// Relations is the golang structure for table relations.
type Relations struct {
	Id           uuid.UUID   `json:"id"            orm:"id"            description:""` //
	RepoId       uuid.UUID   `json:"repo_id"       orm:"repo_id"       description:""` //
	SourceUid    string      `json:"source_uid"    orm:"source_uid"    description:""` //
	TargetUid    string      `json:"target_uid"    orm:"target_uid"    description:""` //
	RelationType string      `json:"relation_type" orm:"relation_type" description:""` //
	Confidence   float64     `json:"confidence"    orm:"confidence"    description:""` //
	Reason       string      `json:"reason"        orm:"reason"        description:""` //
	Properties   *gjson.Json `json:"properties"    orm:"properties"    description:""` //
	CreatedAt    time.Time   `json:"created_at"    orm:"created_at"    description:""` //
}
