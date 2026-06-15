// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// EmailVerificationTokens is the golang structure of table email_verification_tokens for DAO operations like Where/Data.
type EmailVerificationTokens struct {
	g.Meta    `orm:"table:email_verification_tokens, do:true"`
	Id        any //
	UserId    any //
	Email     any //
	TokenHash any //
	ExpiresAt any //
	UsedAt    any //
	CreatedAt any //
}
