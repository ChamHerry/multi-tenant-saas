package repository

import (
	"context"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

var (
	uuidPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	sqlAliasPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

type sRepoAccessPolicy struct{}

func init() {
	service.RegisterRepoAccessPolicy(&sRepoAccessPolicy{})
}

func (s *sRepoAccessPolicy) CanReadRepo(ctx context.Context, subject service.RepoAccessSubject, repoID string) (bool, error) {
	if err := validateUUID(repoID, "repo id"); err != nil {
		return false, err
	}
	record, err := g.DB().GetOne(ctx, `SELECT id, visibility FROM public.repos WHERE id=? AND deleted_at IS NULL`, repoID)
	if err != nil {
		return false, gerror.Wrap(err, "select repository")
	}
	if record.IsEmpty() {
		return false, gerror.NewCode(gcode.CodeNotFound, "repository not found")
	}
	visibility := service.RepositoryVisibility(record["visibility"].String())
	if CanReadRepositoryVisibility(visibility, subject, false) {
		return true, nil
	}
	if subject.UserID == "" && subject.TenantID == "" && subject.APIKeyID == "" {
		return false, nil
	}
	grantRecord, err := g.DB().GetOne(ctx, `
SELECT count(1) AS total
FROM public.repo_access_grants
WHERE repo_id=?
  AND permission IN ('read','write','admin')
  AND (expires_at IS NULL OR expires_at > now())
  AND (
    (subject_type='user' AND subject_id::text = ?)
    OR (subject_type='tenant' AND subject_id::text = ?)
    OR (subject_type='api_key' AND subject_id::text = ?)
  )`, repoID, subject.UserID, subject.TenantID, subject.APIKeyID)
	if err != nil {
		return false, gerror.Wrap(err, "select repository access grant")
	}
	return grantRecord["total"].Int() > 0, nil
}

func (s *sRepoAccessPolicy) AccessibleRepoPredicate(alias string, subject service.RepoAccessSubject) (service.RepoAccessPredicate, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		alias = "r"
	}
	if !sqlAliasPattern.MatchString(alias) {
		return service.RepoAccessPredicate{}, gerror.NewCode(gcode.CodeInvalidParameter, "repository SQL alias is invalid")
	}
	qualifier := alias + "."
	grantClauses := []string{"false"}
	args := []any{subject.PlatformAdmin, subject.UserID}
	if subject.UserID != "" {
		grantClauses = append(grantClauses, "(rag.subject_type='user' AND rag.subject_id::text = ?)")
		args = append(args, subject.UserID)
	}
	if subject.TenantID != "" {
		grantClauses = append(grantClauses, "(rag.subject_type='tenant' AND rag.subject_id::text = ?)")
		args = append(args, subject.TenantID)
	}
	if subject.APIKeyID != "" {
		grantClauses = append(grantClauses, "(rag.subject_type='api_key' AND rag.subject_id::text = ?)")
		args = append(args, subject.APIKeyID)
	}
	predicate := `
` + qualifier + `deleted_at IS NULL
AND (
  ?
  OR ` + qualifier + `visibility = 'public'
  OR (` + qualifier + `visibility = 'internal' AND ? <> '')
  OR EXISTS (
    SELECT 1
    FROM public.repo_access_grants rag
    WHERE rag.repo_id = ` + qualifier + `id
      AND rag.permission IN ('read','write','admin')
      AND (rag.expires_at IS NULL OR rag.expires_at > now())
      AND (` + strings.Join(grantClauses, " OR ") + `)
  )
)`
	return service.RepoAccessPredicate{SQL: predicate, Args: args}, nil
}

func SubjectFromContext(ctx context.Context) service.RepoAccessSubject {
	var subject service.RepoAccessSubject
	if tc, ok := service.TenantContextFromCtx(ctx); ok {
		subject.TenantID = tc.TenantID
		subject.UserID = tc.UserID
		subject.APIKeyID = tc.APIKeyID
	}
	if identity, ok := service.AuthIdentityFromCtx(ctx); ok {
		if subject.TenantID == "" {
			subject.TenantID = identity.TenantID
		}
		if subject.UserID == "" {
			subject.UserID = identity.UserID
		}
		if subject.APIKeyID == "" {
			subject.APIKeyID = identity.APIKeyID
		}
	}
	if _, ok := service.PlatformAdminContextFromCtx(ctx); ok {
		subject.PlatformAdmin = true
	}
	return subject
}

func CanReadRepositoryVisibility(visibility service.RepositoryVisibility, subject service.RepoAccessSubject, hasGrant bool) bool {
	if subject.PlatformAdmin {
		return true
	}
	switch visibility {
	case service.RepositoryVisibilityPublic:
		return true
	case service.RepositoryVisibilityInternal:
		return subject.UserID != ""
	case service.RepositoryVisibilityPrivate, service.RepositoryVisibilityUnknown:
		return hasGrant
	default:
		return false
	}
}

func validateUUID(value string, label string) error {
	if value == "" {
		return gerror.NewCodef(gcode.CodeMissingParameter, "%s is required", label)
	}
	if !uuidPattern.MatchString(value) {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "%s is invalid", label)
	}
	return nil
}
