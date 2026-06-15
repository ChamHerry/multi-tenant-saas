// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PasswordResetTokens is the golang structure of table password_reset_tokens for DAO operations like Where/Data.
type PasswordResetTokens struct {
	g.Meta      `orm:"table:password_reset_tokens, do:true"`
	Id          any //
	UserId      any //
	TokenHash   any //
	ExpiresAt   any //
	UsedAt      any //
	CreatedAt   any //
	RequestedIp any //
}
