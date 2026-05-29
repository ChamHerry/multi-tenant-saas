package apikey

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/logic/rbac"
	"repomind-temp/internal/service"
	"repomind-temp/utility/uuid"
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type sAPIKey struct{}

func init() {
	service.RegisterAPIKey(&sAPIKey{})
}

func (s *sAPIKey) CreatePersonal(ctx context.Context, in service.CreatePersonalAPIKeyInput) (*service.CreatedAPIKey, error) {
	return s.createPersonal(ctx, in)
}

func (s *sAPIKey) createPersonal(ctx context.Context, in service.CreatePersonalAPIKeyInput) (*service.CreatedAPIKey, error) {
	input, err := validatePersonalCreateInput(in)
	if err != nil {
		return nil, err
	}
	if err = ensureActiveUser(ctx, input.UserID); err != nil {
		return nil, err
	}
	if err = enforcePersonalKeyLimit(ctx, input.UserID); err != nil {
		return nil, err
	}
	for _, grant := range input.Grants {
		if err = ensureActiveMembership(ctx, grant.TenantID, input.UserID); err != nil {
			return nil, err
		}
	}
	secret, err := apiKeySecret(ctx)
	if err != nil {
		return nil, err
	}
	rawKey, err := generateRawKey(input.UserID)
	if err != nil {
		return nil, err
	}
	keyID := uuid.GenerateV4()
	if err = validateInternalID(keyID); err != nil {
		return nil, err
	}
	keyPrefix := rawKeyPrefix(rawKey)
	hash := hashRawKey(rawKey, secret)
	var created *service.APIKey
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record, err := tx.Ctx(ctx).GetOne(`
INSERT INTO public.api_keys(id, tenant_id, user_id, name, key_hash, key_prefix, scopes, expires_at, key_type, created_by_user_id, created_at)
VALUES (?, NULLIF(?, '')::uuid, ?, ?, ?, ?, ?::text[], NULLIF(?, '')::timestamptz, ?, ?, now())
RETURNING id, tenant_id, user_id, name, key_type, key_prefix, scopes, last_used_at, expires_at, created_at, revoked_at, created_by_user_id`,
			keyID, "", input.UserID, input.Name, hash, keyPrefix, textArrayLiteral(input.Scopes), expiresAtString(input.ExpiresAt), service.APIKeyTypePersonal, input.UserID)
		if err != nil {
			return gerror.Wrap(err, "insert api key")
		}
		item, err := mapAPIKey(record)
		if err != nil {
			return err
		}
		for _, grant := range input.Grants {
			grantID := uuid.GenerateV4()
			if err = validateInternalID(grantID); err != nil {
				return err
			}
			grantRecord, err := tx.Ctx(ctx).GetOne(`
INSERT INTO public.api_key_tenant_grants(id, api_key_id, tenant_id, scopes, status, granted_by_user_id, created_at, updated_at)
VALUES (?, ?, ?, ?::text[], 'active', ?, now(), now())
RETURNING id, api_key_id, tenant_id, scopes, status, granted_by_user_id, revoked_by_user_id, created_at, updated_at, revoked_at`,
				grantID, keyID, grant.TenantID, textArrayLiteral(grant.Scopes), input.UserID)
			if err != nil {
				return gerror.Wrap(err, "insert api key tenant grant")
			}
			mappedGrant, err := mapAPIKeyGrant(grantRecord)
			if err != nil {
				return err
			}
			item.TenantGrants = append(item.TenantGrants, *mappedGrant)
		}
		created = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &service.CreatedAPIKey{APIKey: *created, RawKey: rawKey}, nil
}

func (s *sAPIKey) ListPersonal(ctx context.Context, userID, tenantID string) ([]service.APIKey, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	result, err := g.DB().GetAll(ctx, `
SELECT DISTINCT ak.id, ak.tenant_id, ak.user_id, ak.name, ak.key_type, ak.key_prefix, ak.scopes, ak.last_used_at, ak.expires_at, ak.created_at, ak.revoked_at, ak.created_by_user_id
FROM public.api_keys ak
JOIN public.api_key_tenant_grants g ON g.api_key_id = ak.id
WHERE ak.user_id=? AND ak.key_type=? AND g.tenant_id=?
ORDER BY ak.created_at DESC`, userID, service.APIKeyTypePersonal, tenantID)
	if err != nil {
		return nil, gerror.Wrap(err, "list tenant-scoped api keys")
	}
	items := make([]service.APIKey, 0, len(result))
	for _, record := range result {
		item, err := mapAPIKey(record)
		if err != nil {
			return nil, err
		}
		grants, err := listKeyGrantsForTenant(ctx, item.ID, tenantID)
		if err != nil {
			return nil, err
		}
		item.TenantGrants = grants
		items = append(items, *item)
	}
	return items, nil
}

func (s *sAPIKey) RevokePersonal(ctx context.Context, userID, tenantID, apiKeyID string) error {
	if err := validateInternalID(userID); err != nil {
		return err
	}
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if err := validateInternalID(apiKeyID); err != nil {
		return err
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Ctx(ctx).Exec(`
UPDATE public.api_keys
SET revoked_at=COALESCE(revoked_at, now())
WHERE user_id=?
  AND id=?
  AND revoked_at IS NULL
  AND id IN (
    SELECT g.api_key_id
    FROM public.api_key_tenant_grants g
    WHERE g.api_key_id=? AND g.tenant_id=?
  )`, userID, apiKeyID, apiKeyID, tenantID)
		if err != nil {
			return gerror.Wrap(err, "revoke tenant-scoped api key")
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return gerror.Newf("tenant-scoped api key %s not found for user %s in tenant %s", apiKeyID, userID, tenantID)
		}
		_, err = tx.Ctx(ctx).Exec(`
UPDATE public.api_key_tenant_grants
SET status='revoked', revoked_at=COALESCE(revoked_at, now()), revoked_by_user_id=?, updated_at=now()
WHERE api_key_id=? AND revoked_at IS NULL`, userID, apiKeyID)
		return gerror.Wrap(err, "revoke tenant-scoped api key grants")
	})
}

func (s *sAPIKey) ResolveTenantGrant(ctx context.Context, apiKeyID, tenantID string) (*service.APIKeyTenantGrant, error) {
	if err := validateInternalID(apiKeyID); err != nil {
		return nil, err
	}
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	record, err := g.DB().GetOne(ctx, `
SELECT
    g.id,
    g.api_key_id,
    g.tenant_id,
    g.scopes,
    g.status,
    g.granted_by_user_id,
    g.revoked_by_user_id,
    g.created_at,
    g.updated_at,
    g.revoked_at,
    t.slug AS tenant_slug,
    t.name AS tenant_name,
    ak.user_id,
    ak.name AS key_name,
    ak.key_prefix
FROM public.api_key_tenant_grants g
JOIN public.api_keys ak ON ak.id = g.api_key_id
JOIN public.users u ON u.id = ak.user_id
JOIN public.tenants t ON t.id = g.tenant_id
JOIN public.tenant_memberships tm ON tm.tenant_id = g.tenant_id AND tm.user_id = ak.user_id
WHERE g.api_key_id=?
  AND g.tenant_id=?
  AND g.status='active'
  AND g.revoked_at IS NULL
  AND ak.revoked_at IS NULL
  AND (ak.expires_at IS NULL OR ak.expires_at > now())
  AND u.status='active'
  AND u.deleted_at IS NULL
  AND t.status='active'
  AND t.deleted_at IS NULL
  AND tm.status='active'
  AND tm.deleted_at IS NULL
LIMIT 1`, apiKeyID, tenantID)
	if err != nil {
		return nil, gerror.Wrap(err, "select api key tenant grant")
	}
	if record.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key is not granted to this tenant")
	}
	return mapAPIKeyGrant(record)
}

func (s *sAPIKey) Authenticate(ctx context.Context, rawKey string) (*service.AuthIdentity, error) {
	if rawKey == "" || !isSupportedRawKey(rawKey) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "invalid api key")
	}
	secret, err := apiKeySecret(ctx)
	if err != nil {
		return nil, err
	}
	hash := hashRawKey(rawKey, secret)
	record, err := g.DB().GetOne(ctx, `
SELECT ak.id, ak.tenant_id, ak.user_id, ak.scopes, ak.key_type
FROM public.api_keys ak
JOIN public.users u ON u.id = ak.user_id
WHERE ak.key_hash=?
  AND ak.revoked_at IS NULL
  AND (ak.expires_at IS NULL OR ak.expires_at > now())
  AND u.status='active'
  AND u.deleted_at IS NULL
LIMIT 1`, hash)
	if err != nil {
		return nil, gerror.Wrap(err, "select api key")
	}
	if record.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key is invalid, revoked, expired, or not attached to an active user")
	}
	scopes := parseTextArray(record["scopes"].String())
	if len(scopes) == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key has no scopes")
	}
	apiKeyID := record["id"].String()
	_, _ = g.DB().Exec(ctx, `UPDATE public.api_keys SET last_used_at=now() WHERE id=?`, apiKeyID)
	return &service.AuthIdentity{
		UserID:   record["user_id"].String(),
		TenantID: "",
		Type:     "api_key",
		Scopes:   scopes,
		APIKeyID: apiKeyID,
		KeyType:  record["key_type"].String(),
	}, nil
}

func validatePersonalCreateInput(in service.CreatePersonalAPIKeyInput) (service.CreatePersonalAPIKeyInput, error) {
	if err := validateInternalID(in.UserID); err != nil {
		return in, err
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, gerror.NewCode(gcode.CodeMissingParameter, "api key name is required")
	}
	scopes, err := normalizeKeyScopes(in.Scopes)
	if err != nil {
		return in, err
	}
	in.Scopes = scopes
	for i, grant := range in.Grants {
		if err = validateInternalID(grant.TenantID); err != nil {
			return in, err
		}
		grantScopes, err := normalizeGrantScopes(grant.Scopes, scopes)
		if err != nil {
			return in, err
		}
		in.Grants[i].Scopes = grantScopes
	}
	return in, nil
}

func normalizeKeyScopes(scopes []string) ([]string, error) {
	items, err := normalizeScopes(scopes, false)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "api key scopes are required")
	}
	return items, nil
}

func normalizeGrantScopes(scopes []string, keyScopes []string) ([]string, error) {
	items, err := normalizeScopes(scopes, true)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return items, nil
	}
	for _, scope := range items {
		if !scopeIncludedInKey(scope, keyScopes) {
			return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "grant scope %q is not included in api key scopes", scope)
		}
	}
	return items, nil
}

func normalizeScopes(scopes []string, allowEmpty bool) ([]string, error) {
	seen := map[string]struct{}{}
	items := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "api key scope cannot be empty")
		}
		if !rbac.KnownPermission(scope) {
			return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "unknown api key scope %q", scope)
		}
		if _, ok := seen[scope]; ok {
			return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "duplicate api key scope %q", scope)
		}
		seen[scope] = struct{}{}
		items = append(items, scope)
	}
	if !allowEmpty && len(items) == 0 {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "api key scopes are required")
	}
	return items, nil
}

func scopeIncludedInKey(scope string, keyScopes []string) bool {
	for _, keyScope := range keyScopes {
		if keyScope == "*" || keyScope == scope {
			return true
		}
	}
	return false
}

func ensureActiveUser(ctx context.Context, userID string) error {
	record, err := g.DB().GetOne(ctx, `
SELECT 1
FROM public.users
WHERE id=? AND status='active' AND deleted_at IS NULL
LIMIT 1`, userID)
	if err != nil {
		return gerror.Wrap(err, "select active api key user")
	}
	if record.IsEmpty() {
		return gerror.Newf("active user %s not found", userID)
	}
	return nil
}

func ensureActiveMembership(ctx context.Context, tenantID, userID string) error {
	record, err := g.DB().GetOne(ctx, `
SELECT 1
FROM public.tenant_memberships tm
JOIN public.users u ON u.id = tm.user_id
JOIN public.tenants t ON t.id = tm.tenant_id
WHERE tm.tenant_id=? AND tm.user_id=?
  AND tm.status='active' AND tm.deleted_at IS NULL
  AND u.status='active' AND u.deleted_at IS NULL
  AND t.status='active' AND t.deleted_at IS NULL
LIMIT 1`, tenantID, userID)
	if err != nil {
		return gerror.Wrap(err, "select active api key membership")
	}
	if record.IsEmpty() {
		return gerror.Newf("active membership tenant=%s user=%s not found", tenantID, userID)
	}
	return nil
}

func enforcePersonalKeyLimit(ctx context.Context, userID string) error {
	limit := g.Cfg().MustGet(ctx, "auth.apiKey.maxPersonalKeysPerUser", 10).Int()
	if limit <= 0 {
		return nil
	}
	record, err := g.DB().GetOne(ctx, `
SELECT count(*) AS used
FROM public.api_keys
WHERE user_id=? AND key_type=? AND revoked_at IS NULL`, userID, service.APIKeyTypePersonal)
	if err != nil {
		return gerror.Wrap(err, "select personal api key usage")
	}
	if record["used"].Int() >= limit {
		return gerror.NewCodef(gcode.CodeNotAuthorized, "personal api key limit exceeded: limit=%d", limit)
	}
	return nil
}

func listKeyGrants(ctx context.Context, apiKeyID string) ([]service.APIKeyTenantGrant, error) {
	result, err := g.DB().GetAll(ctx, `
SELECT
    g.id,
    g.api_key_id,
    g.tenant_id,
    g.scopes,
    g.status,
    g.granted_by_user_id,
    g.revoked_by_user_id,
    g.created_at,
    g.updated_at,
    g.revoked_at,
    t.slug AS tenant_slug,
    t.name AS tenant_name
FROM public.api_key_tenant_grants g
JOIN public.tenants t ON t.id = g.tenant_id
WHERE g.api_key_id=?
ORDER BY g.created_at DESC`, apiKeyID)
	if err != nil {
		return nil, gerror.Wrap(err, "list api key grants")
	}
	items := make([]service.APIKeyTenantGrant, 0, len(result))
	for _, record := range result {
		item, err := mapAPIKeyGrant(record)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func listKeyGrantsForTenant(ctx context.Context, apiKeyID, tenantID string) ([]service.APIKeyTenantGrant, error) {
	result, err := g.DB().GetAll(ctx, `
SELECT
    g.id,
    g.api_key_id,
    g.tenant_id,
    g.scopes,
    g.status,
    g.granted_by_user_id,
    g.revoked_by_user_id,
    g.created_at,
    g.updated_at,
    g.revoked_at,
    t.slug AS tenant_slug,
    t.name AS tenant_name
FROM public.api_key_tenant_grants g
JOIN public.tenants t ON t.id = g.tenant_id
WHERE g.api_key_id=? AND g.tenant_id=?
ORDER BY g.created_at DESC`, apiKeyID, tenantID)
	if err != nil {
		return nil, gerror.Wrap(err, "list tenant api key grants")
	}
	items := make([]service.APIKeyTenantGrant, 0, len(result))
	for _, record := range result {
		item, err := mapAPIKeyGrant(record)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func apiKeySecret(ctx context.Context) (string, error) {
	secret := strings.TrimSpace(g.Cfg().MustGet(ctx, "auth.apiKey.secret", "").String())
	if secret == "" {
		return "", gerror.NewCode(gcode.CodeMissingConfiguration, "auth.apiKey.secret is required")
	}
	return secret, nil
}

func generateRawKey(subjectID string) (string, error) {
	if err := validateInternalID(subjectID); err != nil {
		return "", err
	}
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", gerror.Wrap(err, "generate api key random bytes")
	}
	shortSubject := strings.ReplaceAll(subjectID, "-", "")[:8]
	return "saas_" + shortSubject + "_" + hex.EncodeToString(buf), nil
}

func isSupportedRawKey(rawKey string) bool {
	return strings.HasPrefix(rawKey, "saas_") || strings.HasPrefix(rawKey, "st_") || strings.HasPrefix(rawKey, "rpm_")
}

func rawKeyPrefix(rawKey string) string {
	parts := strings.Split(rawKey, "_")
	if len(parts) < 2 {
		return rawKey
	}
	return parts[0] + "_" + parts[1]
}

func hashRawKey(rawKey, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(rawKey))
	return hex.EncodeToString(mac.Sum(nil))
}

func mapAPIKey(record gdb.Record) (*service.APIKey, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	tenantID := record["tenant_id"].String()
	if tenantID != "" {
		if err := validateInternalID(tenantID); err != nil {
			return nil, err
		}
	}
	userID := record["user_id"].String()
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	createdByUserID := record["created_by_user_id"].String()
	if createdByUserID != "" {
		if err := validateInternalID(createdByUserID); err != nil {
			return nil, err
		}
	}
	keyType := record["key_type"].String()
	if keyType == "" {
		keyType = service.APIKeyTypePersonal
	}
	return &service.APIKey{
		ID:              id,
		TenantID:        tenantID,
		UserID:          userID,
		Name:            record["name"].String(),
		KeyType:         keyType,
		KeyPrefix:       record["key_prefix"].String(),
		Scopes:          parseTextArray(record["scopes"].String()),
		LastUsedAt:      nullableTime(record["last_used_at"]),
		ExpiresAt:       nullableTime(record["expires_at"]),
		CreatedAt:       record["created_at"].Time(),
		RevokedAt:       nullableTime(record["revoked_at"]),
		CreatedByUserID: createdByUserID,
	}, nil
}

func mapAPIKeyGrant(record gdb.Record) (*service.APIKeyTenantGrant, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	apiKeyID := record["api_key_id"].String()
	if err := validateInternalID(apiKeyID); err != nil {
		return nil, err
	}
	tenantID := record["tenant_id"].String()
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	grantedByUserID := record["granted_by_user_id"].String()
	if grantedByUserID != "" {
		if err := validateInternalID(grantedByUserID); err != nil {
			return nil, err
		}
	}
	revokedByUserID := record["revoked_by_user_id"].String()
	if revokedByUserID != "" {
		if err := validateInternalID(revokedByUserID); err != nil {
			return nil, err
		}
	}
	return &service.APIKeyTenantGrant{
		ID:              id,
		APIKeyID:        apiKeyID,
		TenantID:        tenantID,
		TenantSlug:      record["tenant_slug"].String(),
		TenantName:      record["tenant_name"].String(),
		UserID:          record["user_id"].String(),
		KeyName:         record["key_name"].String(),
		KeyPrefix:       record["key_prefix"].String(),
		Scopes:          parseTextArray(record["scopes"].String()),
		Status:          record["status"].String(),
		GrantedByUserID: grantedByUserID,
		RevokedByUserID: revokedByUserID,
		CreatedAt:       record["created_at"].Time(),
		UpdatedAt:       record["updated_at"].Time(),
		RevokedAt:       nullableTime(record["revoked_at"]),
	}, nil
}

func textArrayLiteral(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.ReplaceAll(item, `\`, `\\`)
		item = strings.ReplaceAll(item, `"`, `\"`)
		quoted = append(quoted, `"`+item+`"`)
	}
	return `{` + strings.Join(quoted, `,`) + `}`
}

func expiresAtString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTextArray(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "{}" || value == "[]" {
		return nil
	}
	if strings.HasPrefix(value, "[") {
		var items []string
		if err := json.Unmarshal([]byte(value), &items); err == nil {
			return items
		}
	}
	value = strings.TrimPrefix(strings.TrimSuffix(value, "}"), "{")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, ` "`)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

func nullableTime(value any) *time.Time {
	v, ok := value.(interface {
		IsNil() bool
		Time(...string) time.Time
	})
	if !ok || v.IsNil() {
		return nil
	}
	t := v.Time()
	return &t
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.Newf("invalid internal uuid %q", id)
	}
	return nil
}
