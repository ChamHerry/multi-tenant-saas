// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PasswordResetTokensDao is the data access object for the table password_reset_tokens.
type PasswordResetTokensDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  PasswordResetTokensColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// PasswordResetTokensColumns defines and stores column names for the table password_reset_tokens.
type PasswordResetTokensColumns struct {
	Id          string
	UserId      string
	TokenHash   string
	ExpiresAt   string
	UsedAt      string
	CreatedAt   string
	RequestedIp string
}

// passwordResetTokensColumns holds the columns for the table password_reset_tokens.
var passwordResetTokensColumns = PasswordResetTokensColumns{
	Id:          "id",
	UserId:      "user_id",
	TokenHash:   "token_hash",
	ExpiresAt:   "expires_at",
	UsedAt:      "used_at",
	CreatedAt:   "created_at",
	RequestedIp: "requested_ip",
}

// NewPasswordResetTokensDao creates and returns a new DAO object for table data access.
func NewPasswordResetTokensDao(handlers ...gdb.ModelHandler) *PasswordResetTokensDao {
	return &PasswordResetTokensDao{
		group:    "default",
		table:    "password_reset_tokens",
		columns:  passwordResetTokensColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PasswordResetTokensDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PasswordResetTokensDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PasswordResetTokensDao) Columns() PasswordResetTokensColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PasswordResetTokensDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PasswordResetTokensDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *PasswordResetTokensDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
