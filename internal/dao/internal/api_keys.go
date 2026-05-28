// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ApiKeysDao is the data access object for the table api_keys.
type ApiKeysDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ApiKeysColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ApiKeysColumns defines and stores column names for the table api_keys.
type ApiKeysColumns struct {
	Id              string //
	TenantId        string //
	UserId          string //
	Name            string //
	KeyHash         string //
	Scopes          string //
	LastUsedAt      string //
	ExpiresAt       string //
	CreatedAt       string //
	RevokedAt       string //
	KeyPrefix       string //
	CreatedByUserId string //
}

// apiKeysColumns holds the columns for the table api_keys.
var apiKeysColumns = ApiKeysColumns{
	Id:              "id",
	TenantId:        "tenant_id",
	UserId:          "user_id",
	Name:            "name",
	KeyHash:         "key_hash",
	Scopes:          "scopes",
	LastUsedAt:      "last_used_at",
	ExpiresAt:       "expires_at",
	CreatedAt:       "created_at",
	RevokedAt:       "revoked_at",
	KeyPrefix:       "key_prefix",
	CreatedByUserId: "created_by_user_id",
}

// NewApiKeysDao creates and returns a new DAO object for table data access.
func NewApiKeysDao(handlers ...gdb.ModelHandler) *ApiKeysDao {
	return &ApiKeysDao{
		group:    "default",
		table:    "api_keys",
		columns:  apiKeysColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ApiKeysDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ApiKeysDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ApiKeysDao) Columns() ApiKeysColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ApiKeysDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ApiKeysDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ApiKeysDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
