// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// UserTotpConfigs is the golang structure of table user_totp_configs for DAO operations like Where/Data.
type UserTotpConfigs struct {
	g.Meta          `orm:"table:user_totp_configs, do:true"`
	Id              any //
	UserId          any //
	SecretEncrypted any //
	Enabled         any //
	EnabledAt       any //
	CreatedAt       any //
	UpdatedAt       any //
}
