// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// EmailVerificationTokensDao is the data access object for the table email_verification_tokens.
type EmailVerificationTokensDao struct {
	table    string                          // table is the underlying table name of the DAO.
	group    string                          // group is the database configuration group name of the current DAO.
	columns  EmailVerificationTokensColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler              // handlers for customized model modification.
}

// EmailVerificationTokensColumns defines and stores column names for the table email_verification_tokens.
type EmailVerificationTokensColumns struct {
	Id        string
	UserId    string
	Email     string
	TokenHash string
	ExpiresAt string
	UsedAt    string
	CreatedAt string
}

// emailVerificationTokensColumns holds the columns for the table email_verification_tokens.
var emailVerificationTokensColumns = EmailVerificationTokensColumns{
	Id:        "id",
	UserId:    "user_id",
	Email:     "email",
	TokenHash: "token_hash",
	ExpiresAt: "expires_at",
	UsedAt:    "used_at",
	CreatedAt: "created_at",
}

// NewEmailVerificationTokensDao creates and returns a new DAO object for table data access.
func NewEmailVerificationTokensDao(handlers ...gdb.ModelHandler) *EmailVerificationTokensDao {
	return &EmailVerificationTokensDao{
		group:    "default",
		table:    "email_verification_tokens",
		columns:  emailVerificationTokensColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *EmailVerificationTokensDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *EmailVerificationTokensDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *EmailVerificationTokensDao) Columns() EmailVerificationTokensColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *EmailVerificationTokensDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *EmailVerificationTokensDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *EmailVerificationTokensDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
