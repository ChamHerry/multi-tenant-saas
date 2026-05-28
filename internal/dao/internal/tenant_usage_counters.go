// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantUsageCountersDao is the data access object for the table tenant_usage_counters.
type TenantUsageCountersDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  TenantUsageCountersColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// TenantUsageCountersColumns defines and stores column names for the table tenant_usage_counters.
type TenantUsageCountersColumns struct {
	TenantId      string //
	Metric        string //
	PeriodStart   string //
	PeriodEnd     string //
	Used          string //
	Reserved      string //
	LimitSnapshot string //
	UpdatedAt     string //
}

// tenantUsageCountersColumns holds the columns for the table tenant_usage_counters.
var tenantUsageCountersColumns = TenantUsageCountersColumns{
	TenantId:      "tenant_id",
	Metric:        "metric",
	PeriodStart:   "period_start",
	PeriodEnd:     "period_end",
	Used:          "used",
	Reserved:      "reserved",
	LimitSnapshot: "limit_snapshot",
	UpdatedAt:     "updated_at",
}

// NewTenantUsageCountersDao creates and returns a new DAO object for table data access.
func NewTenantUsageCountersDao(handlers ...gdb.ModelHandler) *TenantUsageCountersDao {
	return &TenantUsageCountersDao{
		group:    "default",
		table:    "tenant_usage_counters",
		columns:  tenantUsageCountersColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TenantUsageCountersDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TenantUsageCountersDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TenantUsageCountersDao) Columns() TenantUsageCountersColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TenantUsageCountersDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TenantUsageCountersDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TenantUsageCountersDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
