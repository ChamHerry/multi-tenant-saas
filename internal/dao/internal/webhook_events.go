// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WebhookEventsDao is the data access object for the table webhook_events.
type WebhookEventsDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  WebhookEventsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// WebhookEventsColumns defines and stores column names for the table webhook_events.
type WebhookEventsColumns struct {
	Id                 string //
	ReceivedByTenantId string //
	RepoId             string //
	Provider           string //
	EventType          string //
	DeliveryId         string //
	PayloadHash        string //
	Payload            string //
	Status             string //
	AnalyzeJobId       string //
	ReceivedAt         string //
	ProcessedAt        string //
	ErrorMsg           string //
}

// webhookEventsColumns holds the columns for the table webhook_events.
var webhookEventsColumns = WebhookEventsColumns{
	Id:                 "id",
	ReceivedByTenantId: "received_by_tenant_id",
	RepoId:             "repo_id",
	Provider:           "provider",
	EventType:          "event_type",
	DeliveryId:         "delivery_id",
	PayloadHash:        "payload_hash",
	Payload:            "payload",
	Status:             "status",
	AnalyzeJobId:       "analyze_job_id",
	ReceivedAt:         "received_at",
	ProcessedAt:        "processed_at",
	ErrorMsg:           "error_msg",
}

// NewWebhookEventsDao creates and returns a new DAO object for table data access.
func NewWebhookEventsDao(handlers ...gdb.ModelHandler) *WebhookEventsDao {
	return &WebhookEventsDao{
		group:    "default",
		table:    "webhook_events",
		columns:  webhookEventsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WebhookEventsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WebhookEventsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WebhookEventsDao) Columns() WebhookEventsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WebhookEventsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WebhookEventsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WebhookEventsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
