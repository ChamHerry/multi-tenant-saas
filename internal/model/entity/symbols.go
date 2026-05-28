// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// Symbols is the golang structure for table symbols.
type Symbols struct {
	RepoId         uuid.UUID   `json:"repo_id"         orm:"repo_id"         description:""` //
	Uid            string      `json:"uid"             orm:"uid"             description:""` //
	FileId         uuid.UUID   `json:"file_id"         orm:"file_id"         description:""` //
	Name           string      `json:"name"            orm:"name"            description:""` //
	Kind           string      `json:"kind"            orm:"kind"            description:""` //
	FilePath       string      `json:"file_path"       orm:"file_path"       description:""` //
	StartLine      int         `json:"start_line"      orm:"start_line"      description:""` //
	EndLine        int         `json:"end_line"        orm:"end_line"        description:""` //
	Content        string      `json:"content"         orm:"content"         description:""` //
	ParameterCount int         `json:"parameter_count" orm:"parameter_count" description:""` //
	ReturnType     string      `json:"return_type"     orm:"return_type"     description:""` //
	Properties     *gjson.Json `json:"properties"      orm:"properties"      description:""` //
	ContentHash    string      `json:"content_hash"    orm:"content_hash"    description:""` //
	CreatedAt      time.Time   `json:"created_at"      orm:"created_at"      description:""` //
	UpdatedAt      time.Time   `json:"updated_at"      orm:"updated_at"      description:""` //
}
