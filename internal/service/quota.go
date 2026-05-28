package service

import (
	"context"
	"time"
)

type QuotaMetric string

const (
	MetricMemberCount QuotaMetric = "member.count"
)

type QuotaUsage struct {
	Metric    QuotaMetric `json:"metric"`
	Used      int64       `json:"used"`
	Reserved  int64       `json:"reserved"`
	Limit     *int64      `json:"limit,omitempty"`
	Remaining *int64      `json:"remaining,omitempty"`
}

type TenantQuotaView struct {
	TenantID string       `json:"tenant_id"`
	Plan     string       `json:"plan"`
	Items    []QuotaUsage `json:"items"`
}

type QuotaReservationInput struct {
	TenantID     string
	Metric       QuotaMetric
	Delta        int64
	ResourceType string
	ResourceID   string
	ExpiresAt    *time.Time
}

type QuotaReservation struct {
	ID           string      `json:"id"`
	TenantID     string      `json:"tenant_id"`
	Metric       QuotaMetric `json:"metric"`
	Delta        int64       `json:"delta"`
	Status       string      `json:"status"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	ExpiresAt    *time.Time  `json:"expires_at,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type IQuota interface {
	GetTenantQuota(ctx context.Context, tenantID string) (*TenantQuotaView, error)
	Require(ctx context.Context, tenantID string, metric QuotaMetric, delta int64) error
	RequireAndReserve(ctx context.Context, in QuotaReservationInput) (*QuotaReservation, error)
	CommitReservation(ctx context.Context, reservationID string) error
	ReleaseReservation(ctx context.Context, reservationID string) error
	Recalculate(ctx context.Context, tenantID string) error
}

var localQuota IQuota

func Quota() IQuota {
	if localQuota == nil {
		panic("implement not found for interface IQuota")
	}
	return localQuota
}

func RegisterQuota(i IQuota) {
	localQuota = i
}
