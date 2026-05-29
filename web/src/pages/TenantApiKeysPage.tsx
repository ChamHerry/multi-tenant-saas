import { FormEvent, useCallback, useMemo, useState } from "react";
import { Copy, KeyRound, Plus } from "lucide-react";
import { useTenantAccess } from "@/features/access/access-hooks";
import {
  useCreateTenantAPIKey,
  useRevokeTenantAPIKey,
  useTenantAPIKeys,
} from "@/features/tenant-api-keys/tenant-api-key-hooks";
import {
  createTenantAPIKeyInputSchema,
  knownScopes,
  tenantScopes,
  type TenantAPIKey,
} from "@/features/tenant-api-keys/tenant-api-key-types";
import { useTenantStore } from "@/features/tenants/tenant-store";
import {
  Badge,
  Button,
  Card,
  CardHeader,
  Dialog,
  EmptyState,
  Input,
  ErrorView,
  LoadingView,
  Table,
  Td,
  Th,
} from "@/shared/ui";
import { validateForm, type FieldErrors } from "@/shared/lib/validate";

const defaultScopes = ["user:read", "user:tenant:read", "tenant:read"];

function currentTenantGrant(item: TenantAPIKey, tenantId: string) {
  return item.tenant_grants?.find((grant) => grant.tenant_id === tenantId);
}

export function TenantApiKeysPage() {
  const tenantId = useTenantStore((state) => state.currentTenantId);
  const { tenantAccess } = useTenantAccess(tenantId);
  const apiKeys = useTenantAPIKeys(tenantId);
  const createAPIKey = useCreateTenantAPIKey(tenantId);
  const revokeAPIKey = useRevokeTenantAPIKey(tenantId);
  const [name, setName] = useState("");
  const [scopes, setScopes] = useState<string[]>(defaultScopes);
  const [expiresAt, setExpiresAt] = useState("");
  const [rawKey, setRawKey] = useState("");
  const [isCreateKeyOpen, setIsCreateKeyOpen] = useState(false);
  const [errors, setErrors] = useState<FieldErrors>({});
  const tenantScopeSet = useMemo(() => new Set<string>(tenantScopes), []);
  const allowedTenantScopeSet = useMemo(
    () =>
      new Set<string>(
        tenantAccess?.permissions?.map((permission) => String(permission)) ??
          [],
      ),
    [tenantAccess?.permissions],
  );

  const canSelectScope = useCallback(
    (scope: string) =>
      !tenantScopeSet.has(scope) || allowedTenantScopeSet.has(scope),
    [allowedTenantScopeSet, tenantScopeSet],
  );

  const selectedScopes = useMemo(() => {
    const next = scopes.filter((scope) => canSelectScope(scope));
    if (next.length > 0) {
      return next;
    }
    return defaultScopes.filter((scope) => canSelectScope(scope));
  }, [canSelectScope, scopes]);

  const toggleScope = (scope: string) => {
    if (!canSelectScope(scope)) {
      return;
    }
    setScopes((current) => {
      const base = current.filter((item) => canSelectScope(item));
      const active =
        base.length > 0
          ? base
          : defaultScopes.filter((item) => canSelectScope(item));
      return active.includes(scope)
        ? active.filter((item) => item !== scope)
        : [...active, scope];
    });
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const result = validateForm(createTenantAPIKeyInputSchema, {
      name: name.trim(),
      scopes: selectedScopes,
      expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
    });
    if (result.errors) {
      setErrors(result.errors);
      return;
    }
    setErrors({});
    createAPIKey.mutate(result.data, {
      onSuccess: (data) => {
        setRawKey(data.raw_key);
        setName("");
        setExpiresAt("");
        setIsCreateKeyOpen(false);
      },
    });
  };

  if (!tenantId) {
    return (
      <EmptyState
        title="请先创建或加入组织"
        description="组织 API Keys 需要当前组织上下文。"
      />
    );
  }

  const firstError = apiKeys.error ?? createAPIKey.error ?? revokeAPIKey.error;
  const tenantName = tenantAccess?.tenant_name ?? tenantId;

  return (
    <div className="space-y-6">
      {firstError ? (
        <ErrorView error={firstError} title="组织 API Key 操作失败" />
      ) : null}

      <div className="flex justify-end">
        <Button
          leftIcon={<Plus className="size-4" />}
          onClick={() => setIsCreateKeyOpen(true)}
        >
          创建组织 Key
        </Button>
      </div>

      {rawKey ? (
        <Card className="border-success/30 bg-success-soft hover:border-success/40">
          <CardHeader
            title="请立即复制 raw key"
            description="后端只返回一次 raw_key。刷新页面后只能看到 key_prefix。"
          />
          <div className="flex flex-col gap-3 sm:flex-row">
            <code className="min-w-0 flex-1 overflow-x-auto rounded-panel bg-white p-3 text-sm font-bold text-success-strong">
              {rawKey}
            </code>
            <Button
              type="button"
              variant="secondary"
              leftIcon={<Copy className="size-4" />}
              onClick={() => void navigator.clipboard?.writeText(rawKey)}
            >
              复制
            </Button>
          </div>
        </Card>
      ) : null}

      <Dialog
        open={isCreateKeyOpen}
        title={`为 ${tenantName} 创建组织 Key`}
        onClose={() => setIsCreateKeyOpen(false)}
      >
        <form className="space-y-4" onSubmit={submit}>
          <Input
            label="名称"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="tenant-automation"
            error={errors.name}
          />
          <Input
            label="过期时间，可选"
            type="datetime-local"
            value={expiresAt}
            onChange={(event) => setExpiresAt(event.target.value)}
          />
          {errors.scopes ? (
            <span className="text-xs text-danger">{errors.scopes}</span>
          ) : null}
          <div>
            <div className="mb-2 text-xs font-bold uppercase tracking-wide text-muted">
              Scopes
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
              {knownScopes.map((scope) => {
                const disabled = !canSelectScope(scope);
                return (
                  <label
                    key={scope}
                    className="flex items-center gap-2 rounded-panel border border-line bg-surface-soft px-3 py-2 text-sm text-muted"
                  >
                    <input
                      type="checkbox"
                      checked={selectedScopes.includes(scope)}
                      disabled={disabled}
                      onChange={() => toggleScope(scope)}
                    />
                    <span>
                      {scope}
                      {disabled ? (
                        <span className="ml-2 text-xs text-subtle">
                          当前组织权限不足
                        </span>
                      ) : null}
                    </span>
                  </label>
                );
              })}
            </div>
          </div>
          <Button
            type="submit"
            disabled={selectedScopes.length === 0}
            isLoading={createAPIKey.isPending}
            leftIcon={<Plus className="size-4" />}
          >
            创建组织 Key
          </Button>
        </form>
      </Dialog>

      <Card>
        <CardHeader
          title="当前组织 Key 列表"
          description={`仅展示当前用户在 ${tenantName} 下创建/持有的 Key（GET /api/v1/tenants/{tenant}/api-keys）。`}
        />
        {apiKeys.isLoading ? (
          <LoadingView label="加载组织 API Keys..." />
        ) : (apiKeys.data?.api_keys.length ?? 0) === 0 ? (
          <EmptyState
            title="暂无组织 API Key"
            description="创建后可用于当前组织上下文下的接口调用；调用组织接口时必须携带 tenant selector。"
          />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <Th>名称</Th>
                  <Th>前缀</Th>
                  <Th>Key Scopes</Th>
                  <Th>组织授权</Th>
                  <Th>状态</Th>
                  <Th>操作</Th>
                </tr>
              </thead>
              <tbody>
                {apiKeys.data?.api_keys.map((item) => {
                  const grant = currentTenantGrant(item, tenantId);
                  return (
                    <tr key={item.id}>
                      <Td>
                        <div className="flex items-center gap-2 font-bold text-ink">
                          <KeyRound className="size-4 text-brand" />
                          {item.name}
                        </div>
                        <div className="mt-1 text-xs text-subtle">
                          {item.id}
                        </div>
                      </Td>
                      <Td>
                        <code className="text-xs text-muted">
                          {item.key_prefix}
                        </code>
                      </Td>
                      <Td>
                        <div className="flex max-w-xs flex-wrap gap-1">
                          {item.scopes.map((scope) => (
                            <Badge key={scope} tone="blue">
                              {scope}
                            </Badge>
                          ))}
                        </div>
                      </Td>
                      <Td>
                        {grant ? (
                          <div className="space-y-2 text-xs text-muted">
                            <div>
                              <Badge tone={grant.revoked_at ? "red" : "green"}>
                                {grant.revoked_at ? "revoked" : "active"}
                              </Badge>{" "}
                              {grant.tenant_name ||
                                grant.tenant_slug ||
                                grant.tenant_id}
                            </div>
                            <div className="flex max-w-xs flex-wrap gap-1">
                              {grant.scopes.length === 0 ? (
                                <span>无 tenant scope</span>
                              ) : null}
                              {grant.scopes.map((scope) => (
                                <Badge key={scope} tone="purple">
                                  {scope}
                                </Badge>
                              ))}
                            </div>
                          </div>
                        ) : (
                          <span className="text-xs text-subtle">
                            当前组织无有效授权
                          </span>
                        )}
                      </Td>
                      <Td>
                        <Badge tone={item.revoked_at ? "red" : "green"}>
                          {item.revoked_at ? "revoked" : "active"}
                        </Badge>
                      </Td>
                      <Td>
                        <Button
                          size="sm"
                          variant="danger"
                          disabled={Boolean(item.revoked_at)}
                          isLoading={revokeAPIKey.isPending}
                          onClick={() => revokeAPIKey.mutate(item.id)}
                        >
                          吊销
                        </Button>
                      </Td>
                    </tr>
                  );
                })}
              </tbody>
            </Table>
          </div>
        )}
      </Card>
    </div>
  );
}
