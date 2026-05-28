// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProcessesDao is the data access object for the table processes.
type ProcessesDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ProcessesColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ProcessesColumns defines and stores column names for the table processes.
type ProcessesColumns struct {
	RepoId        string //
	Id            string //
	Label         string //
	ProcessType   string //
	StepCount     string //
	Communities   string //
	EntryPointUid string //
	TerminalUid   string //
	Summary       string //
	Metadata      string //
	CreatedAt     string //
	UpdatedAt     string //
}

// processesColumns holds the columns for the table processes.
var processesColumns = ProcessesColumns{
	RepoId:        "repo_id",
	Id:            "id",
	Label:         "label",
	ProcessType:   "process_type",
	StepCount:     "step_count",
	Communities:   "communities",
	EntryPointUid: "entry_point_uid",
	TerminalUid:   "terminal_uid",
	Summary:       "summary",
	Metadata:      "metadata",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewProcessesDao creates and returns a new DAO object for table data access.
func NewProcessesDao(handlers ...gdb.ModelHandler) *ProcessesDao {
	return &ProcessesDao{
		group:    "default",
		table:    "processes",
		columns:  processesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProcessesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProcessesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProcessesDao) Columns() ProcessesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProcessesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProcessesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProcessesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
