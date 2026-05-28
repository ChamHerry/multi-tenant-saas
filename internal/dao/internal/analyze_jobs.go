// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AnalyzeJobsDao is the data access object for the table analyze_jobs.
type AnalyzeJobsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AnalyzeJobsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AnalyzeJobsColumns defines and stores column names for the table analyze_jobs.
type AnalyzeJobsColumns struct {
	Id                  string //
	RequestedByTenantId string //
	RepoId              string //
	Trigger             string //
	Mode                string //
	Status              string //
	Priority            string //
	Progress            string //
	WorkerId            string //
	FromCommit          string //
	ToCommit            string //
	ChangedFiles        string //
	Stats               string //
	ErrorCode           string //
	ErrorMsg            string //
	RetryCount          string //
	MaxRetries          string //
	LockedAt            string //
	StartedAt           string //
	FinishedAt          string //
	CreatedAt           string //
	UpdatedAt           string //
}

// analyzeJobsColumns holds the columns for the table analyze_jobs.
var analyzeJobsColumns = AnalyzeJobsColumns{
	Id:                  "id",
	RequestedByTenantId: "requested_by_tenant_id",
	RepoId:              "repo_id",
	Trigger:             "trigger",
	Mode:                "mode",
	Status:              "status",
	Priority:            "priority",
	Progress:            "progress",
	WorkerId:            "worker_id",
	FromCommit:          "from_commit",
	ToCommit:            "to_commit",
	ChangedFiles:        "changed_files",
	Stats:               "stats",
	ErrorCode:           "error_code",
	ErrorMsg:            "error_msg",
	RetryCount:          "retry_count",
	MaxRetries:          "max_retries",
	LockedAt:            "locked_at",
	StartedAt:           "started_at",
	FinishedAt:          "finished_at",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
}

// NewAnalyzeJobsDao creates and returns a new DAO object for table data access.
func NewAnalyzeJobsDao(handlers ...gdb.ModelHandler) *AnalyzeJobsDao {
	return &AnalyzeJobsDao{
		group:    "default",
		table:    "analyze_jobs",
		columns:  analyzeJobsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AnalyzeJobsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AnalyzeJobsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AnalyzeJobsDao) Columns() AnalyzeJobsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AnalyzeJobsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AnalyzeJobsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AnalyzeJobsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
