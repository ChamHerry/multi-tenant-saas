// ===========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ===========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserTotpBackupCodesDao is the data access object for the table user_totp_backup_codes.
type UserTotpBackupCodesDao struct {
	table    string
	group    string
	columns  UserTotpBackupCodesColumns
	handlers []gdb.ModelHandler
}

// UserTotpBackupCodesColumns defines and stores column names for the table user_totp_backup_codes.
type UserTotpBackupCodesColumns struct {
	Id        string //
	UserId    string //
	CodeHash  string //
	UsedAt    string //
	CreatedAt string //
}

var userTotpBackupCodesColumns = UserTotpBackupCodesColumns{
	Id:        "id",
	UserId:    "user_id",
	CodeHash:  "code_hash",
	UsedAt:    "used_at",
	CreatedAt: "created_at",
}

// NewUserTotpBackupCodesDao creates and returns a new DAO object for table data access.
func NewUserTotpBackupCodesDao(handlers ...gdb.ModelHandler) *UserTotpBackupCodesDao {
	return &UserTotpBackupCodesDao{
		group:    "default",
		table:    "user_totp_backup_codes",
		columns:  userTotpBackupCodesColumns,
		handlers: handlers,
	}
}

func (dao *UserTotpBackupCodesDao) DB() gdb.DB                          { return g.DB(dao.group) }
func (dao *UserTotpBackupCodesDao) Table() string                       { return dao.table }
func (dao *UserTotpBackupCodesDao) Columns() UserTotpBackupCodesColumns { return dao.columns }
func (dao *UserTotpBackupCodesDao) Group() string                       { return dao.group }

func (dao *UserTotpBackupCodesDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

func (dao *UserTotpBackupCodesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
