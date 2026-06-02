// ===========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ===========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserTotpConfigsDao is the data access object for the table user_totp_configs.
type UserTotpConfigsDao struct {
	table    string
	group    string
	columns  UserTotpConfigsColumns
	handlers []gdb.ModelHandler
}

// UserTotpConfigsColumns defines and stores column names for the table user_totp_configs.
type UserTotpConfigsColumns struct {
	Id              string //
	UserId          string //
	SecretEncrypted string //
	Enabled         string //
	EnabledAt       string //
	CreatedAt       string //
	UpdatedAt       string //
}

var userTotpConfigsColumns = UserTotpConfigsColumns{
	Id:              "id",
	UserId:          "user_id",
	SecretEncrypted: "secret_encrypted",
	Enabled:         "enabled",
	EnabledAt:       "enabled_at",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

// NewUserTotpConfigsDao creates and returns a new DAO object for table data access.
func NewUserTotpConfigsDao(handlers ...gdb.ModelHandler) *UserTotpConfigsDao {
	return &UserTotpConfigsDao{
		group:    "default",
		table:    "user_totp_configs",
		columns:  userTotpConfigsColumns,
		handlers: handlers,
	}
}

func (dao *UserTotpConfigsDao) DB() gdb.DB                      { return g.DB(dao.group) }
func (dao *UserTotpConfigsDao) Table() string                   { return dao.table }
func (dao *UserTotpConfigsDao) Columns() UserTotpConfigsColumns { return dao.columns }
func (dao *UserTotpConfigsDao) Group() string                   { return dao.group }

func (dao *UserTotpConfigsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

func (dao *UserTotpConfigsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
