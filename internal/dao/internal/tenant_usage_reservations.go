// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantUsageReservationsDao is the data access object for the table tenant_usage_reservations.
type TenantUsageReservationsDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  TenantUsageReservationsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// TenantUsageReservationsColumns defines and stores column names for the table tenant_usage_reservations.
type TenantUsageReservationsColumns struct {
	Id           string //
	TenantId     string //
	Metric       string //
	Delta        string //
	Status       string //
	ResourceType string //
	ResourceId   string //
	ExpiresAt    string //
	CreatedAt    string //
	UpdatedAt    string //
}

// tenantUsageReservationsColumns holds the columns for the table tenant_usage_reservations.
var tenantUsageReservationsColumns = TenantUsageReservationsColumns{
	Id:           "id",
	TenantId:     "tenant_id",
	Metric:       "metric",
	Delta:        "delta",
	Status:       "status",
	ResourceType: "resource_type",
	ResourceId:   "resource_id",
	ExpiresAt:    "expires_at",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewTenantUsageReservationsDao creates and returns a new DAO object for table data access.
func NewTenantUsageReservationsDao(handlers ...gdb.ModelHandler) *TenantUsageReservationsDao {
	return &TenantUsageReservationsDao{
		group:    "default",
		table:    "tenant_usage_reservations",
		columns:  tenantUsageReservationsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TenantUsageReservationsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TenantUsageReservationsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TenantUsageReservationsDao) Columns() TenantUsageReservationsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TenantUsageReservationsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TenantUsageReservationsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TenantUsageReservationsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
