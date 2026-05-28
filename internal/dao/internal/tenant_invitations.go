// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantInvitationsDao is the data access object for the table tenant_invitations.
type TenantInvitationsDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  TenantInvitationsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// TenantInvitationsColumns defines and stores column names for the table tenant_invitations.
type TenantInvitationsColumns struct {
	Id               string //
	TenantId         string //
	InviteeEmail     string //
	InviteeUserId    string //
	Role             string //
	Status           string //
	TokenHash        string //
	InvitedByUserId  string //
	AcceptedByUserId string //
	Message          string //
	ExpiresAt        string //
	AcceptedAt       string //
	DeclinedAt       string //
	RevokedAt        string //
	ResentAt         string //
	Metadata         string //
	CreatedAt        string //
	UpdatedAt        string //
}

// tenantInvitationsColumns holds the columns for the table tenant_invitations.
var tenantInvitationsColumns = TenantInvitationsColumns{
	Id:               "id",
	TenantId:         "tenant_id",
	InviteeEmail:     "invitee_email",
	InviteeUserId:    "invitee_user_id",
	Role:             "role",
	Status:           "status",
	TokenHash:        "token_hash",
	InvitedByUserId:  "invited_by_user_id",
	AcceptedByUserId: "accepted_by_user_id",
	Message:          "message",
	ExpiresAt:        "expires_at",
	AcceptedAt:       "accepted_at",
	DeclinedAt:       "declined_at",
	RevokedAt:        "revoked_at",
	ResentAt:         "resent_at",
	Metadata:         "metadata",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
}

// NewTenantInvitationsDao creates and returns a new DAO object for table data access.
func NewTenantInvitationsDao(handlers ...gdb.ModelHandler) *TenantInvitationsDao {
	return &TenantInvitationsDao{
		group:    "default",
		table:    "tenant_invitations",
		columns:  tenantInvitationsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TenantInvitationsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TenantInvitationsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TenantInvitationsDao) Columns() TenantInvitationsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TenantInvitationsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TenantInvitationsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TenantInvitationsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
