// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AuthLoginAttemptsDao is the data access object for the table auth_login_attempts.
type AuthLoginAttemptsDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  AuthLoginAttemptsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// AuthLoginAttemptsColumns defines and stores column names for the table auth_login_attempts.
type AuthLoginAttemptsColumns struct {
	Id            string //
	LoginKey      string //
	Ip            string //
	FailedCount   string //
	LockedUntil   string //
	LastFailedAt  string //
	LastSuccessAt string //
	CreatedAt     string //
	UpdatedAt     string //
}

// authLoginAttemptsColumns holds the columns for the table auth_login_attempts.
var authLoginAttemptsColumns = AuthLoginAttemptsColumns{
	Id:            "id",
	LoginKey:      "login_key",
	Ip:            "ip",
	FailedCount:   "failed_count",
	LockedUntil:   "locked_until",
	LastFailedAt:  "last_failed_at",
	LastSuccessAt: "last_success_at",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewAuthLoginAttemptsDao creates and returns a new DAO object for table data access.
func NewAuthLoginAttemptsDao(handlers ...gdb.ModelHandler) *AuthLoginAttemptsDao {
	return &AuthLoginAttemptsDao{
		group:    "default",
		table:    "auth_login_attempts",
		columns:  authLoginAttemptsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AuthLoginAttemptsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AuthLoginAttemptsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AuthLoginAttemptsDao) Columns() AuthLoginAttemptsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AuthLoginAttemptsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AuthLoginAttemptsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AuthLoginAttemptsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
