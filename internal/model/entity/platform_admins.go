// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// PlatformAdmins is the golang structure for table platform_admins.
type PlatformAdmins struct {
	UserId          uuid.UUID `json:"user_id"            orm:"user_id"            description:""` //
	Role            string    `json:"role"               orm:"role"               description:""` //
	Status          string    `json:"status"             orm:"status"             description:""` //
	CreatedByUserId uuid.UUID `json:"created_by_user_id" orm:"created_by_user_id" description:""` //
	CreatedAt       time.Time `json:"created_at"         orm:"created_at"         description:""` //
	UpdatedAt       time.Time `json:"updated_at"         orm:"updated_at"         description:""` //
}
