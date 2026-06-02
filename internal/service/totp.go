package service

import (
	"context"
	"time"
)

type TOTPSetupResult struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
}

type TOTPVerifyResult struct {
	Valid                bool `json:"valid"`
	UsedBackupCode       bool `json:"used_backup_code"`
	BackupCodesRemaining int  `json:"backup_codes_remaining,omitempty"`
}

type TOTPStatus struct {
	Enabled              bool       `json:"enabled"`
	SetupInitiated       bool       `json:"setup_initiated"`
	EnabledAt            *time.Time `json:"enabled_at"`
	BackupCodesRemaining int        `json:"backup_codes_remaining"`
	CreatedAt            *time.Time `json:"created_at"`
}

type ITOTP interface {
	Setup(ctx context.Context, userID, password string) (*TOTPSetupResult, error)
	Enable(ctx context.Context, userID, code string) (backupCodes []string, err error)
	Disable(ctx context.Context, userID, password, code string) error
	ValidateTOTP(ctx context.Context, userID, code string) (*TOTPVerifyResult, error)
	IsEnabled(ctx context.Context, userID string) (bool, error)
	GetStatus(ctx context.Context, userID string) (*TOTPStatus, error)
	RegenerateBackupCodes(ctx context.Context, userID, password string) (backupCodes []string, err error)
	GenerateToken(ctx context.Context, userID string) (string, error)
	ValidateToken(ctx context.Context, token string) (userID string, err error)
}

var localTOTP ITOTP

func TOTP() ITOTP {
	if localTOTP == nil {
		panic("implement not found for interface ITOTP")
	}
	return localTOTP
}

func RegisterTOTP(i ITOTP) {
	localTOTP = i
}
