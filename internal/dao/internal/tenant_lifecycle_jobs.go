// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantLifecycleJobsDao is the data access object for the table tenant_lifecycle_jobs.
type TenantLifecycleJobsDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  TenantLifecycleJobsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// TenantLifecycleJobsColumns defines and stores column names for the table tenant_lifecycle_jobs.
type TenantLifecycleJobsColumns struct {
	Id                string //
	TenantId          string //
	Type              string //
	Status            string //
	RequestedByUserId string //
	ScheduledAt       string //
	StartedAt         string //
	FinishedAt        string //
	ErrorMessage      string //
	ArtifactUri       string //
	Metadata          string //
	CreatedAt         string //
	UpdatedAt         string //
}

// tenantLifecycleJobsColumns holds the columns for the table tenant_lifecycle_jobs.
var tenantLifecycleJobsColumns = TenantLifecycleJobsColumns{
	Id:                "id",
	TenantId:          "tenant_id",
	Type:              "type",
	Status:            "status",
	RequestedByUserId: "requested_by_user_id",
	ScheduledAt:       "scheduled_at",
	StartedAt:         "started_at",
	FinishedAt:        "finished_at",
	ErrorMessage:      "error_message",
	ArtifactUri:       "artifact_uri",
	Metadata:          "metadata",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewTenantLifecycleJobsDao creates and returns a new DAO object for table data access.
func NewTenantLifecycleJobsDao(handlers ...gdb.ModelHandler) *TenantLifecycleJobsDao {
	return &TenantLifecycleJobsDao{
		group:    "default",
		table:    "tenant_lifecycle_jobs",
		columns:  tenantLifecycleJobsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TenantLifecycleJobsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TenantLifecycleJobsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TenantLifecycleJobsDao) Columns() TenantLifecycleJobsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TenantLifecycleJobsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TenantLifecycleJobsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TenantLifecycleJobsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
