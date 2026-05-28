// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserPasswordCredentials is the golang structure of table user_password_credentials for DAO operations like Where/Data.
type UserPasswordCredentials struct {
	g.Meta            `orm:"table:user_password_credentials, do:true"`
	UserId            any //
	PasswordHash      any //
	HashAlg           any //
	HashCost          any //
	PasswordChangedAt any //
	CreatedAt         any //
	UpdatedAt         any //
}
