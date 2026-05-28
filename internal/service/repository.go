package service

import "context"

type RepositoryVisibility string

const (
	RepositoryVisibilityPublic   RepositoryVisibility = "public"
	RepositoryVisibilityPrivate  RepositoryVisibility = "private"
	RepositoryVisibilityInternal RepositoryVisibility = "internal"
	RepositoryVisibilityUnknown  RepositoryVisibility = "unknown"
)

type Repository struct {
	ID            string               `json:"id"`
	Provider      string               `json:"provider"`
	CodeHostURL   string               `json:"code_host_url"`
	ExternalID    string               `json:"external_id,omitempty"`
	OwnerName     string               `json:"owner_name,omitempty"`
	Name          string               `json:"name"`
	FullName      string               `json:"full_name"`
	RemoteURL     string               `json:"remote_url"`
	Visibility    RepositoryVisibility `json:"visibility"`
	AnalyzeStatus string               `json:"analyze_status"`
}

type RepoAccessSubject struct {
	TenantID      string
	UserID        string
	APIKeyID      string
	PlatformAdmin bool
}

type RepoAccessPredicate struct {
	SQL  string
	Args []any
}

type IRepoAccessPolicy interface {
	CanReadRepo(ctx context.Context, subject RepoAccessSubject, repoID string) (bool, error)
	AccessibleRepoPredicate(alias string, subject RepoAccessSubject) (RepoAccessPredicate, error)
}

var localRepoAccessPolicy IRepoAccessPolicy

func RepoAccessPolicy() IRepoAccessPolicy {
	if localRepoAccessPolicy == nil {
		panic("implement not found for interface IRepoAccessPolicy")
	}
	return localRepoAccessPolicy
}

func RegisterRepoAccessPolicy(i IRepoAccessPolicy) {
	localRepoAccessPolicy = i
}
