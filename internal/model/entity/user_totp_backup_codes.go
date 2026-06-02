// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserTotpBackupCodes is the golang structure for table user_totp_backup_codes.
type UserTotpBackupCodes struct {
	Id        uuid.UUID `json:"id"         orm:"id"         description:""` //
	UserId    uuid.UUID `json:"user_id"    orm:"user_id"    description:""` //
	CodeHash  string    `json:"code_hash"  orm:"code_hash"  description:""` //
	UsedAt    time.Time `json:"used_at"    orm:"used_at"    description:""` //
	CreatedAt time.Time `json:"created_at" orm:"created_at" description:""` //
}
