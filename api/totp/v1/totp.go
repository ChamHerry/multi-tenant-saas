package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"

	authv1 "multi-tenant-saas/api/auth/v1"
)

type AuthUserRes = authv1.AuthUserRes

type VerifyTOTPReq struct {
	g.Meta    `path:"/auth/verify-totp" tags:"Auth" method:"post" summary:"Verify TOTP code and authenticate"`
	TOTPToken string `json:"totp_token" v:"required"`
	Code      string `json:"code" v:"required"`
}

type SetupReq struct {
	g.Meta   `path:"/me/totp/setup" tags:"MyProfile" method:"post" summary:"Initiate TOTP setup"`
	Password string `json:"password" v:"required"`
}

type SetupRes struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
}

type EnableReq struct {
	g.Meta `path:"/me/totp/enable" tags:"MyProfile" method:"post" summary:"Enable TOTP with code verification"`
	Code   string `json:"code" v:"required"`
}

type EnableRes struct {
	OK          bool     `json:"ok"`
	BackupCodes []string `json:"backup_codes"`
}

type DisableReq struct {
	g.Meta   `path:"/me/totp/disable" tags:"MyProfile" method:"post" summary:"Disable TOTP"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type RegenerateBackupCodesReq struct {
	g.Meta   `path:"/me/totp/backup-codes/regenerate" tags:"MyProfile" method:"post" summary:"Regenerate backup codes"`
	Password string `json:"password" v:"required"`
}

type RegenerateBackupCodesRes struct {
	BackupCodes []string `json:"backup_codes"`
}

type StatusReq struct {
	g.Meta `path:"/me/totp/status" tags:"MyProfile" method:"get" summary:"Get TOTP status"`
}

type StatusRes struct {
	Enabled              bool       `json:"enabled"`
	SetupInitiated       bool       `json:"setup_initiated"`
	EnabledAt            *time.Time `json:"enabled_at"`
	BackupCodesRemaining int        `json:"backup_codes_remaining"`
	CreatedAt            *time.Time `json:"created_at"`
}

type ActionRes struct {
	OK bool `json:"ok"`
}
