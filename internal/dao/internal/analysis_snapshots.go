// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AnalysisSnapshotsDao is the data access object for the table analysis_snapshots.
type AnalysisSnapshotsDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  AnalysisSnapshotsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// AnalysisSnapshotsColumns defines and stores column names for the table analysis_snapshots.
type AnalysisSnapshotsColumns struct {
	Id               string //
	RepoId           string //
	CommitHash       string //
	AnalyzeJobId     string //
	FilesCount       string //
	SymbolsCount     string //
	RelationsCount   string //
	ProcessesCount   string //
	CommunitiesCount string //
	Stats            string //
	CreatedAt        string //
}

// analysisSnapshotsColumns holds the columns for the table analysis_snapshots.
var analysisSnapshotsColumns = AnalysisSnapshotsColumns{
	Id:               "id",
	RepoId:           "repo_id",
	CommitHash:       "commit_hash",
	AnalyzeJobId:     "analyze_job_id",
	FilesCount:       "files_count",
	SymbolsCount:     "symbols_count",
	RelationsCount:   "relations_count",
	ProcessesCount:   "processes_count",
	CommunitiesCount: "communities_count",
	Stats:            "stats",
	CreatedAt:        "created_at",
}

// NewAnalysisSnapshotsDao creates and returns a new DAO object for table data access.
func NewAnalysisSnapshotsDao(handlers ...gdb.ModelHandler) *AnalysisSnapshotsDao {
	return &AnalysisSnapshotsDao{
		group:    "default",
		table:    "analysis_snapshots",
		columns:  analysisSnapshotsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AnalysisSnapshotsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AnalysisSnapshotsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AnalysisSnapshotsDao) Columns() AnalysisSnapshotsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AnalysisSnapshotsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AnalysisSnapshotsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AnalysisSnapshotsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
