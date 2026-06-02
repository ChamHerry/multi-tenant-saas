package systemconfigadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/model/entity"
	"multi-tenant-saas/internal/service"
)

const maskedSecretValue = "********"

var (
	keyPattern      = regexp.MustCompile(`^[a-zA-Z0-9._:-]+$`)
	validValueTypes = map[string]struct{}{
		"string": {},
		"number": {},
		"bool":   {},
		"json":   {},
		"secret": {},
	}
	protectedKeys = map[string]struct{}{
		"auth.session.secret":          {},
		"auth.session.cookie.name":     {},
		"auth.session.cookie.csrfName": {},
		"auth.apiKey.secret":           {},
		"server.env":                   {},
		"automigrate.enabled":          {},
	}
)

type sSystemConfigAdmin struct{}

func init() {
	service.RegisterSystemConfigAdmin(&sSystemConfigAdmin{})
}

func (s *sSystemConfigAdmin) List(ctx context.Context, filter service.SystemConfigAdminListFilter) (*service.SystemConfigAdminList, error) {
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL, args := buildListWhere(filter)
	cols := dao.SystemConfig.Columns()

	total, err := dao.SystemConfig.Ctx(ctx).Where(whereSQL, args...).Count()
	if err != nil {
		return nil, gerror.Wrap(err, "count system config")
	}

	var rows []*entity.SystemConfig
	err = dao.SystemConfig.Ctx(ctx).
		Where(whereSQL, args...).
		OrderAsc(cols.Key).
		Limit(limit).
		Offset(offset).
		Scan(&rows)
	if err != nil {
		return nil, gerror.Wrap(err, "list system config")
	}

	items := make([]service.SystemConfigAdminItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapItem(row))
	}
	return &service.SystemConfigAdminList{Items: items, Total: total}, nil
}

func (s *sSystemConfigAdmin) Get(ctx context.Context, key string) (*service.SystemConfigAdminItem, error) {
	key, err := normalizeKey(key)
	if err != nil {
		return nil, err
	}
	config, err := getByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, gerror.NewCode(gcode.CodeNotFound, "system config not found")
	}
	item := mapItem(config)
	return &item, nil
}

func (s *sSystemConfigAdmin) Upsert(ctx context.Context, input service.SystemConfigAdminUpsertInput) (*service.SystemConfigAdminItem, error) {
	key, err := normalizeKey(input.Key)
	if err != nil {
		return nil, err
	}
	valueType := strings.TrimSpace(input.ValueType)
	if _, ok := validValueTypes[valueType]; !ok {
		return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "invalid value_type %q", valueType)
	}
	if isSensitiveKey(key) && valueType != "secret" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "sensitive config keys must use secret value_type")
	}

	existing, err := getByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	if !input.ValueProvided {
		if existing == nil || !isAdminSecret(existing) || valueType != "secret" {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "value is required")
		}
		if err = s.updateMetadataOnly(ctx, existing, strings.TrimSpace(input.Description), input.ActorUserID); err != nil {
			return nil, err
		}
		return s.Get(ctx, key)
	}

	normalizedValue, err := validateAndNormalizeValue(valueType, input.Value)
	if err != nil {
		return nil, err
	}
	if err = service.Config().Set(ctx, &service.ConfigSetParams{
		Key:         key,
		Value:       normalizedValue,
		ValueType:   valueType,
		Description: strings.TrimSpace(input.Description),
	}); err != nil {
		return nil, err
	}
	updated, err := getByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, gerror.NewCode(gcode.CodeNotFound, "system config not found after update")
	}
	if err = writeAudit(ctx, input.ActorUserID, "system_config.updated", key, existing, updated); err != nil {
		return nil, err
	}
	item := mapItem(updated)
	return &item, nil
}

func (s *sSystemConfigAdmin) Delete(ctx context.Context, input service.SystemConfigAdminDeleteInput) error {
	key, err := normalizeKey(input.Key)
	if err != nil {
		return err
	}
	if _, ok := protectedKeys[key]; ok {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "system config %q is protected", key)
	}
	existing, err := getByKey(ctx, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return gerror.NewCode(gcode.CodeNotFound, "system config not found")
	}
	if err = service.Config().Delete(ctx, key); err != nil {
		return err
	}
	return writeAudit(ctx, input.ActorUserID, "system_config.deleted", key, existing, nil)
}

func (s *sSystemConfigAdmin) updateMetadataOnly(ctx context.Context, existing *entity.SystemConfig, description, actorUserID string) error {
	cols := dao.SystemConfig.Columns()
	_, err := dao.SystemConfig.Ctx(ctx).
		Where(cols.Key, existing.Key).
		Data(g.Map{
			cols.Description: description,
			cols.UpdatedAt:   time.Now(),
		}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "update secret config metadata")
	}
	updated, err := getByKey(ctx, existing.Key)
	if err != nil {
		return err
	}
	return writeAudit(ctx, actorUserID, "system_config.updated", existing.Key, existing, updated)
}

func buildListWhere(filter service.SystemConfigAdminListFilter) (string, []any) {
	where := []string{"1=1"}
	args := []any{}
	if category := strings.TrimSpace(filter.Category); category != "" {
		where = append(where, "(key = ? OR key LIKE ?)")
		args = append(args, category, category+".%")
	}
	if query := strings.TrimSpace(filter.Query); query != "" {
		where = append(where, "(key ILIKE ? OR description ILIKE ?)")
		like := "%" + query + "%"
		args = append(args, like, like)
	}
	return strings.Join(where, " AND "), args
}

func getByKey(ctx context.Context, key string) (*entity.SystemConfig, error) {
	var config *entity.SystemConfig
	err := dao.SystemConfig.Ctx(ctx).
		Where(dao.SystemConfig.Columns().Key, key).
		Scan(&config)
	return config, err
}

func normalizeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "key is required")
	}
	if len(key) > 255 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "key must be at most 255 characters")
	}
	if !keyPattern.MatchString(key) {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "key contains unsupported characters")
	}
	return key, nil
}

func validateAndNormalizeValue(valueType, value string) (string, error) {
	switch valueType {
	case "string", "secret":
		return value, nil
	case "number":
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "number value is required")
		}
		if _, err := strconv.ParseFloat(trimmed, 64); err != nil {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "number value is invalid")
		}
		return trimmed, nil
	case "bool":
		trimmed := strings.TrimSpace(value)
		parsed, err := strconv.ParseBool(trimmed)
		if err != nil {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "bool value is invalid")
		}
		if parsed {
			return "true", nil
		}
		return "false", nil
	case "json":
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "json value is required")
		}
		decoder := json.NewDecoder(bytes.NewReader([]byte(trimmed)))
		decoder.UseNumber()
		var decoded any
		if err := decoder.Decode(&decoded); err != nil {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "json value is invalid")
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "json value has trailing data")
		}
		encoded, err := json.Marshal(decoded)
		if err != nil {
			return "", gerror.Wrap(err, "normalize json value")
		}
		return string(encoded), nil
	default:
		return "", gerror.NewCodef(gcode.CodeInvalidParameter, "invalid value_type %q", valueType)
	}
}

func mapItem(config *entity.SystemConfig) service.SystemConfigAdminItem {
	isSecret := isAdminSecret(config)
	hasValue := config.Value != ""
	valueType := config.ValueType
	if isSecret {
		valueType = "secret"
	}
	item := service.SystemConfigAdminItem{
		Key:         config.Key,
		ValueType:   valueType,
		Description: config.Description,
		Category:    categoryForKey(config.Key),
		IsEncrypted: config.IsEncrypted,
		IsSecret:    isSecret,
		HasValue:    hasValue,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
	}
	if isSecret {
		if hasValue {
			item.MaskedValue = maskedSecretValue
		}
		return item
	}
	item.Value = config.Value
	return item
}

func categoryForKey(key string) string {
	parts := strings.SplitN(key, ".", 2)
	if strings.TrimSpace(parts[0]) == "" {
		return "general"
	}
	return parts[0]
}

func normalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func writeAudit(ctx context.Context, actorUserID, action, key string, before, after *entity.SystemConfig) error {
	metadata := map[string]any{
		"key": key,
	}
	if before != nil {
		metadata["old_value_type"] = before.ValueType
		metadata["old_is_secret"] = isAdminSecret(before)
		metadata["old_has_value"] = before.Value != ""
		metadata["old_masked"] = maskedForAudit(before)
	}
	if after != nil {
		metadata["value_type"] = after.ValueType
		metadata["is_secret"] = isAdminSecret(after)
		metadata["has_value"] = after.Value != ""
		metadata["new_masked"] = maskedForAudit(after)
	}
	return service.Audit().Write(ctx, service.AuditLogInput{
		UserID:       actorUserID,
		Action:       action,
		ResourceType: "system_config",
		ResourceID:   key,
		Metadata:     metadata,
	})
}

func maskedForAudit(config *entity.SystemConfig) string {
	if config == nil || config.Value == "" {
		return ""
	}
	if isAdminSecret(config) {
		return maskedSecretValue
	}
	return fmt.Sprintf("<%s:%d>", config.ValueType, len(config.Value))
}

func isAdminSecret(config *entity.SystemConfig) bool {
	return config.ValueType == "secret" || config.IsEncrypted || isSensitiveKey(config.Key)
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(lower, ".secret") || strings.HasSuffix(lower, ":secret")
}
