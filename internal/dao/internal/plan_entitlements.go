// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PlanEntitlementsDao is the data access object for the table plan_entitlements.
type PlanEntitlementsDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  PlanEntitlementsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// PlanEntitlementsColumns defines and stores column names for the table plan_entitlements.
type PlanEntitlementsColumns struct {
	Plan       string //
	FeatureKey string //
	Enabled    string //
	LimitValue string //
	Metadata   string //
	CreatedAt  string //
	UpdatedAt  string //
}

// planEntitlementsColumns holds the columns for the table plan_entitlements.
var planEntitlementsColumns = PlanEntitlementsColumns{
	Plan:       "plan",
	FeatureKey: "feature_key",
	Enabled:    "enabled",
	LimitValue: "limit_value",
	Metadata:   "metadata",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewPlanEntitlementsDao creates and returns a new DAO object for table data access.
func NewPlanEntitlementsDao(handlers ...gdb.ModelHandler) *PlanEntitlementsDao {
	return &PlanEntitlementsDao{
		group:    "default",
		table:    "plan_entitlements",
		columns:  planEntitlementsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PlanEntitlementsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PlanEntitlementsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PlanEntitlementsDao) Columns() PlanEntitlementsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PlanEntitlementsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PlanEntitlementsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PlanEntitlementsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
