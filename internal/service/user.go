package service

import (
	"context"
	"time"
)

type EnsureUserByIdentityInput struct {
	Provider      string
	AuthID        string
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
	RawProfile    map[string]any
	Metadata      map[string]any
}

type User struct {
	ID          string         `json:"id"`
	Email       string         `json:"email"`
	DisplayName string         `json:"display_name"`
	AvatarURL   string         `json:"avatar_url"`
	Status      string         `json:"status"`
	LastLoginAt *time.Time     `json:"last_login_at,omitempty"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type UserIdentity struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Provider      string     `json:"provider"`
	AuthID        string     `json:"auth_id"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type UpdateProfileInput struct {
	UserID      string
	DisplayName string
	AvatarURL   string
}

type IUser interface {
	EnsureUserByIdentity(ctx context.Context, in EnsureUserByIdentityInput) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, in UpdateProfileInput) (*User, error)
}

var localUser IUser

func UserService() IUser {
	if localUser == nil {
		panic("implement not found for interface IUser")
	}
	return localUser
}

func RegisterUser(i IUser) {
	localUser = i
}
