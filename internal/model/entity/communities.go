// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// Communities is the golang structure for table communities.
type Communities struct {
	RepoId      uuid.UUID   `json:"repo_id"      orm:"repo_id"      description:""` //
	Id          string      `json:"id"           orm:"id"           description:""` //
	Label       string      `json:"label"        orm:"label"        description:""` //
	Cohesion    float64     `json:"cohesion"     orm:"cohesion"     description:""` //
	SymbolCount int         `json:"symbol_count" orm:"symbol_count" description:""` //
	Keywords    []string    `json:"keywords"     orm:"keywords"     description:""` //
	Description string      `json:"description"  orm:"description"  description:""` //
	Metadata    *gjson.Json `json:"metadata"     orm:"metadata"     description:""` //
	CreatedAt   time.Time   `json:"created_at"   orm:"created_at"   description:""` //
}
