// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AuthSessionsDao is the data access object for the table auth_sessions.
type AuthSessionsDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  AuthSessionsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// AuthSessionsColumns defines and stores column names for the table auth_sessions.
type AuthSessionsColumns struct {
	Id            string //
	UserId        string //
	SecretHash    string //
	CsrfHash      string //
	UserAgent     string //
	Ip            string //
	LastUsedAt    string //
	ExpiresAt     string //
	IdleExpiresAt string //
	RevokedAt     string //
	RevokeReason  string //
	CreatedAt     string //
	UpdatedAt     string //
}

// authSessionsColumns holds the columns for the table auth_sessions.
var authSessionsColumns = AuthSessionsColumns{
	Id:            "id",
	UserId:        "user_id",
	SecretHash:    "secret_hash",
	CsrfHash:      "csrf_hash",
	UserAgent:     "user_agent",
	Ip:            "ip",
	LastUsedAt:    "last_used_at",
	ExpiresAt:     "expires_at",
	IdleExpiresAt: "idle_expires_at",
	RevokedAt:     "revoked_at",
	RevokeReason:  "revoke_reason",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewAuthSessionsDao creates and returns a new DAO object for table data access.
func NewAuthSessionsDao(handlers ...gdb.ModelHandler) *AuthSessionsDao {
	return &AuthSessionsDao{
		group:    "default",
		table:    "auth_sessions",
		columns:  authSessionsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AuthSessionsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AuthSessionsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AuthSessionsDao) Columns() AuthSessionsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AuthSessionsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AuthSessionsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AuthSessionsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
