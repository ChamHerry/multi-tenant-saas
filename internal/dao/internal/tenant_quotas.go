// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantQuotasDao is the data access object for the table tenant_quotas.
type TenantQuotasDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  TenantQuotasColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// TenantQuotasColumns defines and stores column names for the table tenant_quotas.
type TenantQuotasColumns struct {
	TenantId          string //
	MaxDailyRequests  string //
	MaxConcurrentJobs string //
	MaxMembers        string //
	MaxApiKeys        string //
	UpdatedAt         string //
}

// tenantQuotasColumns holds the columns for the table tenant_quotas.
var tenantQuotasColumns = TenantQuotasColumns{
	TenantId:          "tenant_id",
	MaxDailyRequests:  "max_daily_requests",
	MaxConcurrentJobs: "max_concurrent_jobs",
	MaxMembers:        "max_members",
	MaxApiKeys:        "max_api_keys",
	UpdatedAt:         "updated_at",
}

// NewTenantQuotasDao creates and returns a new DAO object for table data access.
func NewTenantQuotasDao(handlers ...gdb.ModelHandler) *TenantQuotasDao {
	return &TenantQuotasDao{
		group:    "default",
		table:    "tenant_quotas",
		columns:  tenantQuotasColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TenantQuotasDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TenantQuotasDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TenantQuotasDao) Columns() TenantQuotasColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TenantQuotasDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TenantQuotasDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TenantQuotasDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
