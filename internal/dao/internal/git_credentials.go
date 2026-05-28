// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GitCredentialsDao is the data access object for the table git_credentials.
type GitCredentialsDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  GitCredentialsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// GitCredentialsColumns defines and stores column names for the table git_credentials.
type GitCredentialsColumns struct {
	Id                 string //
	TenantId           string //
	OwnerUserId        string //
	Provider           string //
	BaseUrl            string //
	AccessTokenCipher  string //
	RefreshTokenCipher string //
	TokenExpiresAt     string //
	Scopes             string //
	Status             string //
	CreatedAt          string //
	UpdatedAt          string //
	DeletedAt          string //
}

// gitCredentialsColumns holds the columns for the table git_credentials.
var gitCredentialsColumns = GitCredentialsColumns{
	Id:                 "id",
	TenantId:           "tenant_id",
	OwnerUserId:        "owner_user_id",
	Provider:           "provider",
	BaseUrl:            "base_url",
	AccessTokenCipher:  "access_token_cipher",
	RefreshTokenCipher: "refresh_token_cipher",
	TokenExpiresAt:     "token_expires_at",
	Scopes:             "scopes",
	Status:             "status",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	DeletedAt:          "deleted_at",
}

// NewGitCredentialsDao creates and returns a new DAO object for table data access.
func NewGitCredentialsDao(handlers ...gdb.ModelHandler) *GitCredentialsDao {
	return &GitCredentialsDao{
		group:    "default",
		table:    "git_credentials",
		columns:  gitCredentialsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *GitCredentialsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *GitCredentialsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *GitCredentialsDao) Columns() GitCredentialsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *GitCredentialsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *GitCredentialsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *GitCredentialsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
