// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReposDao is the data access object for the table repos.
type ReposDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ReposColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ReposColumns defines and stores column names for the table repos.
type ReposColumns struct {
	Id                  string //
	Provider            string //
	RemoteUrl           string //
	CloneUrl            string //
	WebUrl              string //
	Namespace           string //
	Name                string //
	DefaultBranch       string //
	LastCommit          string //
	AnalyzeStatus       string //
	WebhookId           string //
	Stats               string //
	LastAnalyzedAt      string //
	CreatedAt           string //
	UpdatedAt           string //
	DeletedAt           string //
	CodeHostUrl         string //
	ExternalId          string //
	ExternalNodeId      string //
	OwnerName           string //
	FullName            string //
	NormalizedRemoteKey string //
	Visibility          string //
	IsFork              string //
	IsArchived          string //
	IndexedCommitHash   string //
	PermissionSyncedAt  string //
	Metadata            string //
}

// reposColumns holds the columns for the table repos.
var reposColumns = ReposColumns{
	Id:                  "id",
	Provider:            "provider",
	RemoteUrl:           "remote_url",
	CloneUrl:            "clone_url",
	WebUrl:              "web_url",
	Namespace:           "namespace",
	Name:                "name",
	DefaultBranch:       "default_branch",
	LastCommit:          "last_commit",
	AnalyzeStatus:       "analyze_status",
	WebhookId:           "webhook_id",
	Stats:               "stats",
	LastAnalyzedAt:      "last_analyzed_at",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	DeletedAt:           "deleted_at",
	CodeHostUrl:         "code_host_url",
	ExternalId:          "external_id",
	ExternalNodeId:      "external_node_id",
	OwnerName:           "owner_name",
	FullName:            "full_name",
	NormalizedRemoteKey: "normalized_remote_key",
	Visibility:          "visibility",
	IsFork:              "is_fork",
	IsArchived:          "is_archived",
	IndexedCommitHash:   "indexed_commit_hash",
	PermissionSyncedAt:  "permission_synced_at",
	Metadata:            "metadata",
}

// NewReposDao creates and returns a new DAO object for table data access.
func NewReposDao(handlers ...gdb.ModelHandler) *ReposDao {
	return &ReposDao{
		group:    "default",
		table:    "repos",
		columns:  reposColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReposDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReposDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReposDao) Columns() ReposColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReposDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReposDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *ReposDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
