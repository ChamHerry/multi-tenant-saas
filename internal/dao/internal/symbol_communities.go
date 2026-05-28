// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SymbolCommunitiesDao is the data access object for the table symbol_communities.
type SymbolCommunitiesDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  SymbolCommunitiesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// SymbolCommunitiesColumns defines and stores column names for the table symbol_communities.
type SymbolCommunitiesColumns struct {
	RepoId      string //
	SymbolUid   string //
	CommunityId string //
	Score       string //
}

// symbolCommunitiesColumns holds the columns for the table symbol_communities.
var symbolCommunitiesColumns = SymbolCommunitiesColumns{
	RepoId:      "repo_id",
	SymbolUid:   "symbol_uid",
	CommunityId: "community_id",
	Score:       "score",
}

// NewSymbolCommunitiesDao creates and returns a new DAO object for table data access.
func NewSymbolCommunitiesDao(handlers ...gdb.ModelHandler) *SymbolCommunitiesDao {
	return &SymbolCommunitiesDao{
		group:    "default",
		table:    "symbol_communities",
		columns:  symbolCommunitiesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SymbolCommunitiesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SymbolCommunitiesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SymbolCommunitiesDao) Columns() SymbolCommunitiesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SymbolCommunitiesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SymbolCommunitiesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SymbolCommunitiesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
