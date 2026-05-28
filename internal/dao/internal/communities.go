// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CommunitiesDao is the data access object for the table communities.
type CommunitiesDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CommunitiesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CommunitiesColumns defines and stores column names for the table communities.
type CommunitiesColumns struct {
	RepoId      string //
	Id          string //
	Label       string //
	Cohesion    string //
	SymbolCount string //
	Keywords    string //
	Description string //
	Metadata    string //
	CreatedAt   string //
}

// communitiesColumns holds the columns for the table communities.
var communitiesColumns = CommunitiesColumns{
	RepoId:      "repo_id",
	Id:          "id",
	Label:       "label",
	Cohesion:    "cohesion",
	SymbolCount: "symbol_count",
	Keywords:    "keywords",
	Description: "description",
	Metadata:    "metadata",
	CreatedAt:   "created_at",
}

// NewCommunitiesDao creates and returns a new DAO object for table data access.
func NewCommunitiesDao(handlers ...gdb.ModelHandler) *CommunitiesDao {
	return &CommunitiesDao{
		group:    "default",
		table:    "communities",
		columns:  communitiesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CommunitiesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CommunitiesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CommunitiesDao) Columns() CommunitiesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CommunitiesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CommunitiesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CommunitiesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
