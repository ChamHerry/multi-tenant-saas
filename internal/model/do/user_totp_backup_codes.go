// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// UserTotpBackupCodes is the golang structure of table user_totp_backup_codes for DAO operations like Where/Data.
type UserTotpBackupCodes struct {
	g.Meta    `orm:"table:user_totp_backup_codes, do:true"`
	Id        any //
	UserId    any //
	CodeHash  any //
	UsedAt    any //
	CreatedAt any //
}
