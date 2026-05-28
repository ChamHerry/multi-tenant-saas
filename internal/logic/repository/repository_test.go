package repository

import (
	"strings"
	"testing"

	"repomind-temp/internal/service"
)

func TestCanReadRepositoryVisibility(t *testing.T) {
	active := service.RepoAccessSubject{TenantID: "tenant", UserID: "user"}
	apiKey := service.RepoAccessSubject{TenantID: "tenant", APIKeyID: "key"}
	cases := []struct {
		name       string
		visibility service.RepositoryVisibility
		subject    service.RepoAccessSubject
		hasGrant   bool
		want       bool
	}{
		{name: "public anonymous", visibility: service.RepositoryVisibilityPublic, want: true},
		{name: "internal active user", visibility: service.RepositoryVisibilityInternal, subject: active, want: true},
		{name: "internal api key without user", visibility: service.RepositoryVisibilityInternal, subject: apiKey, want: false},
		{name: "private no grant", visibility: service.RepositoryVisibilityPrivate, subject: active, want: false},
		{name: "private with grant", visibility: service.RepositoryVisibilityPrivate, subject: active, hasGrant: true, want: true},
		{name: "unknown with grant", visibility: service.RepositoryVisibilityUnknown, subject: active, hasGrant: true, want: true},
		{name: "platform admin", visibility: service.RepositoryVisibilityPrivate, subject: service.RepoAccessSubject{PlatformAdmin: true}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanReadRepositoryVisibility(tc.visibility, tc.subject, tc.hasGrant); got != tc.want {
				t.Fatalf("CanReadRepositoryVisibility() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAccessibleRepoPredicate(t *testing.T) {
	policy := &sRepoAccessPolicy{}
	predicate, err := policy.AccessibleRepoPredicate("repo", service.RepoAccessSubject{TenantID: "tenant-id", UserID: "user-id", APIKeyID: "api-key-id"})
	if err != nil {
		t.Fatalf("AccessibleRepoPredicate() error = %v", err)
	}
	for _, want := range []string{"repo.deleted_at IS NULL", "repo.visibility = 'public'", "repo_access_grants", "subject_type='user'", "subject_type='tenant'", "subject_type='api_key'"} {
		if !strings.Contains(predicate.SQL, want) {
			t.Fatalf("predicate SQL missing %q:\n%s", want, predicate.SQL)
		}
	}
	if len(predicate.Args) != 5 {
		t.Fatalf("predicate args length = %d, want 5", len(predicate.Args))
	}
}

func TestAccessibleRepoPredicateRejectsUnsafeAlias(t *testing.T) {
	policy := &sRepoAccessPolicy{}
	if _, err := policy.AccessibleRepoPredicate("repos; DROP TABLE public.repos", service.RepoAccessSubject{}); err == nil {
		t.Fatal("AccessibleRepoPredicate() expected unsafe alias error")
	}
}
