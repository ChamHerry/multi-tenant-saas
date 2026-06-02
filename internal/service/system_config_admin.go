package service

import (
	"context"
	"time"
)

type SystemConfigAdminItem struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	ValueType   string    `json:"value_type"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	IsEncrypted bool      `json:"is_encrypted"`
	IsSecret    bool      `json:"is_secret"`
	HasValue    bool      `json:"has_value"`
	MaskedValue string    `json:"masked_value,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SystemConfigAdminListFilter struct {
	Query    string
	Category string
	Limit    int
	Offset   int
}

type SystemConfigAdminList struct {
	Items []SystemConfigAdminItem `json:"items"`
	Total int                     `json:"total"`
}

type SystemConfigAdminUpsertInput struct {
	Key           string
	Value         string
	ValueProvided bool
	ValueType     string
	Description   string
	ActorUserID   string
}

type SystemConfigAdminDeleteInput struct {
	Key         string
	ActorUserID string
}

type ISystemConfigAdmin interface {
	List(ctx context.Context, filter SystemConfigAdminListFilter) (*SystemConfigAdminList, error)
	Get(ctx context.Context, key string) (*SystemConfigAdminItem, error)
	Upsert(ctx context.Context, input SystemConfigAdminUpsertInput) (*SystemConfigAdminItem, error)
	Delete(ctx context.Context, input SystemConfigAdminDeleteInput) error
}

var localSystemConfigAdmin ISystemConfigAdmin

func SystemConfigAdmin() ISystemConfigAdmin {
	if localSystemConfigAdmin == nil {
		panic("implement not found for interface ISystemConfigAdmin")
	}
	return localSystemConfigAdmin
}

func RegisterSystemConfigAdmin(i ISystemConfigAdmin) {
	localSystemConfigAdmin = i
}
