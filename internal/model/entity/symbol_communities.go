// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/google/uuid"
)

// SymbolCommunities is the golang structure for table symbol_communities.
type SymbolCommunities struct {
	RepoId      uuid.UUID `json:"repo_id"      orm:"repo_id"      description:""` //
	SymbolUid   string    `json:"symbol_uid"   orm:"symbol_uid"   description:""` //
	CommunityId string    `json:"community_id" orm:"community_id" description:""` //
	Score       float64   `json:"score"        orm:"score"        description:""` //
}
