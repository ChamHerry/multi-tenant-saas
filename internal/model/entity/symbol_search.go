// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// SymbolSearch is the golang structure for table symbol_search.
type SymbolSearch struct {
	Id                 uuid.UUID   `json:"id"                  orm:"id"                  description:""` //
	RepoId             uuid.UUID   `json:"repo_id"             orm:"repo_id"             description:""` //
	SymbolUid          string      `json:"symbol_uid"          orm:"symbol_uid"          description:""` //
	Name               string      `json:"name"                orm:"name"                description:""` //
	Kind               string      `json:"kind"                orm:"kind"                description:""` //
	FilePath           string      `json:"file_path"           orm:"file_path"           description:""` //
	Content            string      `json:"content"             orm:"content"             description:""` //
	Summary            string      `json:"summary"             orm:"summary"             description:""` //
	CommunityId        string      `json:"community_id"        orm:"community_id"        description:""` //
	ProcessId          string      `json:"process_id"          orm:"process_id"          description:""` //
	NameTsv            string      `json:"name_tsv"            orm:"name_tsv"            description:""` //
	ContentTsv         string      `json:"content_tsv"         orm:"content_tsv"         description:""` //
	Embedding          string      `json:"embedding"           orm:"embedding"           description:""` //
	EmbeddingModel     string      `json:"embedding_model"     orm:"embedding_model"     description:""` //
	EmbeddingDimension int         `json:"embedding_dimension" orm:"embedding_dimension" description:""` //
	SearchBoost        float64     `json:"search_boost"        orm:"search_boost"        description:""` //
	Metadata           *gjson.Json `json:"metadata"            orm:"metadata"            description:""` //
	CreatedAt          time.Time   `json:"created_at"          orm:"created_at"          description:""` //
	UpdatedAt          time.Time   `json:"updated_at"          orm:"updated_at"          description:""` //
}
