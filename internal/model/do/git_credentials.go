// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GitCredentials is the golang structure of table git_credentials for DAO operations like Where/Data.
type GitCredentials struct {
	g.Meta             `orm:"table:git_credentials, do:true"`
	Id                 any      //
	TenantId           any      //
	OwnerUserId        any      //
	Provider           any      //
	BaseUrl            any      //
	AccessTokenCipher  []byte   //
	RefreshTokenCipher []byte   //
	TokenExpiresAt     any      //
	Scopes             []string //
	Status             any      //
	CreatedAt          any      //
	UpdatedAt          any      //
	DeletedAt          any      //
}
