// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// Repos is the golang structure of table repos for DAO operations like Where/Data.
type Repos struct {
	g.Meta              `orm:"table:repos, do:true"`
	Id                  any         //
	Provider            any         //
	RemoteUrl           any         //
	CloneUrl            any         //
	WebUrl              any         //
	Namespace           any         //
	Name                any         //
	DefaultBranch       any         //
	LastCommit          any         //
	AnalyzeStatus       any         //
	WebhookId           any         //
	Stats               *gjson.Json //
	LastAnalyzedAt      any         //
	CreatedAt           any         //
	UpdatedAt           any         //
	DeletedAt           any         //
	CodeHostUrl         any         //
	ExternalId          any         //
	ExternalNodeId      any         //
	OwnerName           any         //
	FullName            any         //
	NormalizedRemoteKey any         //
	Visibility          any         //
	IsFork              any         //
	IsArchived          any         //
	IndexedCommitHash   any         //
	PermissionSyncedAt  any         //
	Metadata            *gjson.Json //
}
