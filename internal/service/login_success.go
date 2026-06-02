package service

import "context"

type CompleteLoginSuccessInput struct {
	UserID    string
	Email     string
	SessionID string
	IP        string
	UserAgent string
}

type ILoginSuccess interface {
	Complete(ctx context.Context, in CompleteLoginSuccessInput) error
}

var localLoginSuccess ILoginSuccess

func LoginSuccess() ILoginSuccess {
	if localLoginSuccess == nil {
		panic("implement not found for interface ILoginSuccess")
	}
	return localLoginSuccess
}

func RegisterLoginSuccess(i ILoginSuccess) {
	localLoginSuccess = i
}
