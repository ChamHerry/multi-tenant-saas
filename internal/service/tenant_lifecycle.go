package service

import (
	"context"
	"time"
)

type TenantLifecycleJob struct {
	ID                string         `json:"id"`
	TenantID          string         `json:"tenant_id"`
	Type              string         `json:"type"`
	Status            string         `json:"status"`
	RequestedByUserID string         `json:"requested_by_user_id,omitempty"`
	ScheduledAt       time.Time      `json:"scheduled_at"`
	StartedAt         *time.Time     `json:"started_at,omitempty"`
	FinishedAt        *time.Time     `json:"finished_at,omitempty"`
	ErrorMessage      string         `json:"error_message,omitempty"`
	ArtifactURI       string         `json:"artifact_uri,omitempty"`
	Metadata          map[string]any `json:"metadata"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type ITenantLifecycle interface {
	RequestExport(ctx context.Context, tenantID, actorUserID string) (*TenantLifecycleJob, error)
	RequestPurge(ctx context.Context, tenantID, actorUserID string, after time.Time) (*TenantLifecycleJob, error)
	RunPendingJobs(ctx context.Context, limit int) error
	CancelJob(ctx context.Context, jobID, actorUserID string) error
}

var localTenantLifecycle ITenantLifecycle

func TenantLifecycle() ITenantLifecycle {
	if localTenantLifecycle == nil {
		panic("implement not found for interface ITenantLifecycle")
	}
	return localTenantLifecycle
}

func RegisterTenantLifecycle(i ITenantLifecycle) {
	localTenantLifecycle = i
}
