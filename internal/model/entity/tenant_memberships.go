// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// TenantMemberships is the golang structure for table tenant_memberships.
type TenantMemberships struct {
	Id              uuid.UUID `json:"id"                 orm:"id"                 description:""` //
	TenantId        uuid.UUID `json:"tenant_id"          orm:"tenant_id"          description:""` //
	UserId          uuid.UUID `json:"user_id"            orm:"user_id"            description:""` //
	Role            string    `json:"role"               orm:"role"               description:""` //
	Status          string    `json:"status"             orm:"status"             description:""` //
	InvitedByUserId uuid.UUID `json:"invited_by_user_id" orm:"invited_by_user_id" description:""` //
	JoinedAt        time.Time `json:"joined_at"          orm:"joined_at"          description:""` //
	CreatedAt       time.Time `json:"created_at"         orm:"created_at"         description:""` //
	UpdatedAt       time.Time `json:"updated_at"         orm:"updated_at"         description:""` //
	DeletedAt       time.Time `json:"deleted_at"         orm:"deleted_at"         description:""` //
}
