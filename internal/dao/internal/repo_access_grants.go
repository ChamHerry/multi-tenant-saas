// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// RepoAccessGrantsDao is the data access object for the table repo_access_grants.
type RepoAccessGrantsDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  RepoAccessGrantsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// RepoAccessGrantsColumns defines and stores column names for the table repo_access_grants.
type RepoAccessGrantsColumns struct {
	RepoId            string //
	SubjectType       string //
	SubjectId         string //
	Permission        string //
	Source            string //
	ExternalAccountId string //
	SyncedAt          string //
	ExpiresAt         string //
	CreatedAt         string //
	UpdatedAt         string //
}

// repoAccessGrantsColumns holds the columns for the table repo_access_grants.
var repoAccessGrantsColumns = RepoAccessGrantsColumns{
	RepoId:            "repo_id",
	SubjectType:       "subject_type",
	SubjectId:         "subject_id",
	Permission:        "permission",
	Source:            "source",
	ExternalAccountId: "external_account_id",
	SyncedAt:          "synced_at",
	ExpiresAt:         "expires_at",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewRepoAccessGrantsDao creates and returns a new DAO object for table data access.
func NewRepoAccessGrantsDao(handlers ...gdb.ModelHandler) *RepoAccessGrantsDao {
	return &RepoAccessGrantsDao{
		group:    "default",
		table:    "repo_access_grants",
		columns:  repoAccessGrantsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *RepoAccessGrantsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *RepoAccessGrantsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *RepoAccessGrantsDao) Columns() RepoAccessGrantsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *RepoAccessGrantsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *RepoAccessGrantsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *RepoAccessGrantsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
