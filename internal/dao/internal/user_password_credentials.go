// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserPasswordCredentialsDao is the data access object for the table user_password_credentials.
type UserPasswordCredentialsDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  UserPasswordCredentialsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// UserPasswordCredentialsColumns defines and stores column names for the table user_password_credentials.
type UserPasswordCredentialsColumns struct {
	UserId            string //
	PasswordHash      string //
	HashAlg           string //
	HashCost          string //
	PasswordChangedAt string //
	CreatedAt         string //
	UpdatedAt         string //
}

// userPasswordCredentialsColumns holds the columns for the table user_password_credentials.
var userPasswordCredentialsColumns = UserPasswordCredentialsColumns{
	UserId:            "user_id",
	PasswordHash:      "password_hash",
	HashAlg:           "hash_alg",
	HashCost:          "hash_cost",
	PasswordChangedAt: "password_changed_at",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewUserPasswordCredentialsDao creates and returns a new DAO object for table data access.
func NewUserPasswordCredentialsDao(handlers ...gdb.ModelHandler) *UserPasswordCredentialsDao {
	return &UserPasswordCredentialsDao{
		group:    "default",
		table:    "user_password_credentials",
		columns:  userPasswordCredentialsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserPasswordCredentialsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserPasswordCredentialsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserPasswordCredentialsDao) Columns() UserPasswordCredentialsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserPasswordCredentialsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserPasswordCredentialsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserPasswordCredentialsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
