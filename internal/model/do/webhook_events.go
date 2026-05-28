// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// WebhookEvents is the golang structure of table webhook_events for DAO operations like Where/Data.
type WebhookEvents struct {
	g.Meta             `orm:"table:webhook_events, do:true"`
	Id                 any         //
	ReceivedByTenantId any         //
	RepoId             any         //
	Provider           any         //
	EventType          any         //
	DeliveryId         any         //
	PayloadHash        any         //
	Payload            *gjson.Json //
	Status             any         //
	AnalyzeJobId       any         //
	ReceivedAt         any         //
	ProcessedAt        any         //
	ErrorMsg           any         //
}
