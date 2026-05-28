// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/google/uuid"
)

// TenantInvitations is the golang structure for table tenant_invitations.
type TenantInvitations struct {
	Id               uuid.UUID   `json:"id"                  orm:"id"                  description:""` //
	TenantId         uuid.UUID   `json:"tenant_id"           orm:"tenant_id"           description:""` //
	InviteeEmail     string      `json:"invitee_email"       orm:"invitee_email"       description:""` //
	InviteeUserId    uuid.UUID   `json:"invitee_user_id"     orm:"invitee_user_id"     description:""` //
	Role             string      `json:"role"                orm:"role"                description:""` //
	Status           string      `json:"status"              orm:"status"              description:""` //
	TokenHash        string      `json:"token_hash"          orm:"token_hash"          description:""` //
	InvitedByUserId  uuid.UUID   `json:"invited_by_user_id"  orm:"invited_by_user_id"  description:""` //
	AcceptedByUserId uuid.UUID   `json:"accepted_by_user_id" orm:"accepted_by_user_id" description:""` //
	Message          string      `json:"message"             orm:"message"             description:""` //
	ExpiresAt        time.Time   `json:"expires_at"          orm:"expires_at"          description:""` //
	AcceptedAt       time.Time   `json:"accepted_at"         orm:"accepted_at"         description:""` //
	DeclinedAt       time.Time   `json:"declined_at"         orm:"declined_at"         description:""` //
	RevokedAt        time.Time   `json:"revoked_at"          orm:"revoked_at"          description:""` //
	ResentAt         time.Time   `json:"resent_at"           orm:"resent_at"           description:""` //
	Metadata         *gjson.Json `json:"metadata"            orm:"metadata"            description:""` //
	CreatedAt        time.Time   `json:"created_at"          orm:"created_at"          description:""` //
	UpdatedAt        time.Time   `json:"updated_at"          orm:"updated_at"          description:""` //
}
