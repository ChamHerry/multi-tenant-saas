package service

import (
	"context"
	"time"
)

const (
	SetupCodeRequired           = "SYSTEM_SETUP_REQUIRED"
	SetupCodeAlreadyInitialized = "SYSTEM_ALREADY_INITIALIZED"
)

type SetupCheck struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

type SetupChecks struct {
	Database   SetupCheck `json:"database"`
	Migrations SetupCheck `json:"migrations"`
	Encryption SetupCheck `json:"encryption"`
	Redis      SetupCheck `json:"redis"`
}

type SetupStatus struct {
	Initialized         bool        `json:"initialized"`
	RequiresSetup       bool        `json:"requires_setup"`
	LegacyInitialized   bool        `json:"legacy_initialized"`
	InitializedAt       *time.Time  `json:"initialized_at,omitempty"`
	InitializedByUserID string      `json:"initialized_by_user_id,omitempty"`
	Version             string      `json:"version,omitempty"`
	Missing             []string    `json:"missing"`
	Checks              SetupChecks `json:"checks"`
}

type SetupAdminInput struct {
	Email       string
	Password    string
	DisplayName string
}

type SetupRuntimeInput struct {
	ServerEnv             string
	WebBaseURL            string
	GenerateSessionSecret bool
	GenerateAPIKeySecret  bool
}

type CompleteSetupInput struct {
	Admin     SetupAdminInput
	Runtime   SetupRuntimeInput
	IP        string
	UserAgent string
}

type CompleteSetupResult struct {
	Initialized bool  `json:"initialized"`
	User        *User `json:"user"`
}

type ISystemSetup interface {
	State(ctx context.Context) (*SetupStatus, error)
	EnsureLegacyState(ctx context.Context) error
	ValidateRuntime(ctx context.Context) error
	ValidateRuntimeOrSetupPending(ctx context.Context) error
	Complete(ctx context.Context, in CompleteSetupInput) (*CompleteSetupResult, error)
}

var localSystemSetup ISystemSetup

func SystemSetup() ISystemSetup {
	if localSystemSetup == nil {
		panic("implement not found for interface ISystemSetup")
	}
	return localSystemSetup
}

func RegisterSystemSetup(i ISystemSetup) {
	localSystemSetup = i
}
