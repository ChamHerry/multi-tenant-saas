// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ApiKeyTenantGrantsDao is the data access object for the table api_key_tenant_grants.
type ApiKeyTenantGrantsDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  ApiKeyTenantGrantsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// ApiKeyTenantGrantsColumns defines and stores column names for the table api_key_tenant_grants.
type ApiKeyTenantGrantsColumns struct {
	Id              string //
	ApiKeyId        string //
	TenantId        string //
	Scopes          string //
	Status          string //
	GrantedByUserId string //
	RevokedByUserId string //
	CreatedAt       string //
	UpdatedAt       string //
	RevokedAt       string //
}

// apiKeyTenantGrantsColumns holds the columns for the table api_key_tenant_grants.
var apiKeyTenantGrantsColumns = ApiKeyTenantGrantsColumns{
	Id:              "id",
	ApiKeyId:        "api_key_id",
	TenantId:        "tenant_id",
	Scopes:          "scopes",
	Status:          "status",
	GrantedByUserId: "granted_by_user_id",
	RevokedByUserId: "revoked_by_user_id",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	RevokedAt:       "revoked_at",
}

// NewApiKeyTenantGrantsDao creates and returns a new DAO object for table data access.
func NewApiKeyTenantGrantsDao(handlers ...gdb.ModelHandler) *ApiKeyTenantGrantsDao {
	return &ApiKeyTenantGrantsDao{
		group:    "default",
		table:    "api_key_tenant_grants",
		columns:  apiKeyTenantGrantsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ApiKeyTenantGrantsDao) DB() gdb.DB { return g.DB(dao.group) }

// Table returns the table name of the current DAO.
func (dao *ApiKeyTenantGrantsDao) Table() string { return dao.table }

// Columns returns all column names of the current DAO.
func (dao *ApiKeyTenantGrantsDao) Columns() ApiKeyTenantGrantsColumns { return dao.columns }

// Group returns the database configuration group name of the current DAO.
func (dao *ApiKeyTenantGrantsDao) Group() string { return dao.group }

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ApiKeyTenantGrantsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ApiKeyTenantGrantsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
