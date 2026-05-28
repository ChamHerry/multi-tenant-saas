// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// UserIdentities is the golang structure of table user_identities for DAO operations like Where/Data.
type UserIdentities struct {
	g.Meta        `orm:"table:user_identities, do:true"`
	Id            any         //
	UserId        any         //
	Provider      any         //
	AuthId        any         //
	Email         any         //
	EmailVerified any         //
	RawProfile    *gjson.Json //
	LastLoginAt   any         //
	CreatedAt     any         //
	UpdatedAt     any         //
}
