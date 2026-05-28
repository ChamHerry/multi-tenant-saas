// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SymbolSearchDao is the data access object for the table symbol_search.
type SymbolSearchDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  SymbolSearchColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// SymbolSearchColumns defines and stores column names for the table symbol_search.
type SymbolSearchColumns struct {
	Id                 string //
	RepoId             string //
	SymbolUid          string //
	Name               string //
	Kind               string //
	FilePath           string //
	Content            string //
	Summary            string //
	CommunityId        string //
	ProcessId          string //
	NameTsv            string //
	ContentTsv         string //
	Embedding          string //
	EmbeddingModel     string //
	EmbeddingDimension string //
	SearchBoost        string //
	Metadata           string //
	CreatedAt          string //
	UpdatedAt          string //
}

// symbolSearchColumns holds the columns for the table symbol_search.
var symbolSearchColumns = SymbolSearchColumns{
	Id:                 "id",
	RepoId:             "repo_id",
	SymbolUid:          "symbol_uid",
	Name:               "name",
	Kind:               "kind",
	FilePath:           "file_path",
	Content:            "content",
	Summary:            "summary",
	CommunityId:        "community_id",
	ProcessId:          "process_id",
	NameTsv:            "name_tsv",
	ContentTsv:         "content_tsv",
	Embedding:          "embedding",
	EmbeddingModel:     "embedding_model",
	EmbeddingDimension: "embedding_dimension",
	SearchBoost:        "search_boost",
	Metadata:           "metadata",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewSymbolSearchDao creates and returns a new DAO object for table data access.
func NewSymbolSearchDao(handlers ...gdb.ModelHandler) *SymbolSearchDao {
	return &SymbolSearchDao{
		group:    "default",
		table:    "symbol_search",
		columns:  symbolSearchColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SymbolSearchDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SymbolSearchDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SymbolSearchDao) Columns() SymbolSearchColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SymbolSearchDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SymbolSearchDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SymbolSearchDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
