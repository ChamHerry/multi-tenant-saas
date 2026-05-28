// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SymbolsDao is the data access object for the table symbols.
type SymbolsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SymbolsColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SymbolsColumns defines and stores column names for the table symbols.
type SymbolsColumns struct {
	RepoId         string //
	Uid            string //
	FileId         string //
	Name           string //
	Kind           string //
	FilePath       string //
	StartLine      string //
	EndLine        string //
	Content        string //
	ParameterCount string //
	ReturnType     string //
	Properties     string //
	ContentHash    string //
	CreatedAt      string //
	UpdatedAt      string //
}

// symbolsColumns holds the columns for the table symbols.
var symbolsColumns = SymbolsColumns{
	RepoId:         "repo_id",
	Uid:            "uid",
	FileId:         "file_id",
	Name:           "name",
	Kind:           "kind",
	FilePath:       "file_path",
	StartLine:      "start_line",
	EndLine:        "end_line",
	Content:        "content",
	ParameterCount: "parameter_count",
	ReturnType:     "return_type",
	Properties:     "properties",
	ContentHash:    "content_hash",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
}

// NewSymbolsDao creates and returns a new DAO object for table data access.
func NewSymbolsDao(handlers ...gdb.ModelHandler) *SymbolsDao {
	return &SymbolsDao{
		group:    "default",
		table:    "symbols",
		columns:  symbolsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SymbolsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SymbolsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SymbolsDao) Columns() SymbolsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SymbolsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SymbolsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SymbolsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
