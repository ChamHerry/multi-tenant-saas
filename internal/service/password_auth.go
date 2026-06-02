package service

import "context"

type PasswordLoginInput struct {
	Email     string
	Password  string
	IP        string
	UserAgent string
}

type PasswordLoginResult struct {
	User    *User               `json:"user"`
	Session *AuthSession        `json:"session"`
	Cookies *AuthSessionCookies `json:"-"`
}

type CreatePasswordUserInput struct {
	Email         string
	Password      string
	DisplayName   string
	EmailVerified bool
}

type RegisterPasswordUserInput struct {
	Email       string
	Password    string
	DisplayName string
}

type IPasswordAuth interface {
	Login(ctx context.Context, in PasswordLoginInput) (*PasswordLoginResult, error)
	RegisterPasswordUser(ctx context.Context, in RegisterPasswordUserInput) (*User, error)
	CreatePasswordUser(ctx context.Context, in CreatePasswordUserInput) (*User, error)
	SetPassword(ctx context.Context, email, password string) (*User, error)
	ChangePassword(ctx context.Context, userID, currentSessionID, oldPassword, newPassword string) error
	// ForgotPassword initiates the password reset flow by email.
	// It always returns nil error to prevent user enumeration,
	// even when the email does not exist.
	ForgotPassword(ctx context.Context, email, ip string) error
	// ResetPassword verifies a reset token and sets a new password.
	// On success, all existing sessions for the user are revoked.
	ResetPassword(ctx context.Context, token, newPassword, ip string) error
}

var localPasswordAuth IPasswordAuth

func PasswordAuth() IPasswordAuth {
	if localPasswordAuth == nil {
		panic("implement not found for interface IPasswordAuth")
	}
	return localPasswordAuth
}

func RegisterPasswordAuth(i IPasswordAuth) {
	localPasswordAuth = i
}
