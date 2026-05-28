// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantMembershipsDao is the data access object for the table tenant_memberships.
type TenantMembershipsDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  TenantMembershipsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// TenantMembershipsColumns defines and stores column names for the table tenant_memberships.
type TenantMembershipsColumns struct {
	Id              string //
	TenantId        string //
	UserId          string //
	Role            string //
	Status          string //
	InvitedByUserId string //
	JoinedAt        string //
	CreatedAt       string //
	UpdatedAt       string //
	DeletedAt       string //
}

// tenantMembershipsColumns holds the columns for the table tenant_memberships.
var tenantMembershipsColumns = TenantMembershipsColumns{
	Id:              "id",
	TenantId:        "tenant_id",
	UserId:          "user_id",
	Role:            "role",
	Status:          "status",
	InvitedByUserId: "invited_by_user_id",
	JoinedAt:        "joined_at",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewTenantMembershipsDao creates and returns a new DAO object for table data access.
func NewTenantMembershipsDao(handlers ...gdb.ModelHandler) *TenantMembershipsDao {
	return &TenantMembershipsDao{
		group:    "default",
		table:    "tenant_memberships",
		columns:  tenantMembershipsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TenantMembershipsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TenantMembershipsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TenantMembershipsDao) Columns() TenantMembershipsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TenantMembershipsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TenantMembershipsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TenantMembershipsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
