import { FormEvent, useMemo, useState } from "react";
import { Copy, KeyRound, Plus } from "lucide-react";
import {
  useCreatePersonalAPIKey,
  usePersonalAPIKeys,
  useRevokePersonalAPIKey,
} from "@/features/api-keys/api-key-hooks";
import { knownScopes, tenantScopes } from "@/features/api-keys/api-key-types";
import { useTenantStore } from "@/features/tenants/tenant-store";
import { Badge } from "@/shared/ui/Badge";
import { Button } from "@/shared/ui/Button";
import { Card, CardHeader } from "@/shared/ui/Card";
import { EmptyState } from "@/shared/ui/EmptyState";
import { Input } from "@/shared/ui/Input";
import { ErrorView, LoadingView } from "@/shared/ui/StatusView";
import { Table, Td, Th } from "@/shared/ui/Table";

export function PersonalApiKeysPage() {
  const currentTenantId = useTenantStore((state) => state.currentTenantId);
  const apiKeys = usePersonalAPIKeys();
  const createAPIKey = useCreatePersonalAPIKey();
  const revokeAPIKey = useRevokePersonalAPIKey();
  const [name, setName] = useState("");
  const [scopes, setScopes] = useState<string[]>([
    "user:read",
    "user:tenant:read",
    "tenant:read",
  ]);
  const [expiresAt, setExpiresAt] = useState("");
  const [grantCurrentTenant, setGrantCurrentTenant] = useState(
    Boolean(currentTenantId),
  );
  const [rawKey, setRawKey] = useState("");
  const tenantScopeSet = useMemo(() => new Set<string>(tenantScopes), []);

  const toggleScope = (scope: string) => {
    setScopes((current) =>
      current.includes(scope)
        ? current.filter((item) => item !== scope)
        : [...current, scope],
    );
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const grantScopes = scopes.filter((scope) => tenantScopeSet.has(scope));
    createAPIKey.mutate(
      {
        name: name.trim(),
        scopes,
        expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
        grants:
          grantCurrentTenant && currentTenantId
            ? [{ tenant_id: currentTenantId, scopes: grantScopes }]
            : [],
      },
      {
        onSuccess: (data) => {
          setRawKey(data.raw_key);
          setName("");
          setExpiresAt("");
        },
      },
    );
  };

  const firstError = apiKeys.error ?? createAPIKey.error ?? revokeAPIKey.error;

  return (
    <div className="space-y-6">
      <div>
        <Badge tone="purple">Personal API Keys</Badge>
        <h1 className="mt-3 text-3xl font-black text-ink">个人 API Key</h1>
        <p className="mt-2 text-sm leading-6 text-muted">
          API Key 绑定当前用户身份；调用租户接口时必须显式提供
          X-Tenant-ID，最终权限由用户当前租户角色、key scopes 和租户 grant
          scopes 共同收窄。
        </p>
      </div>

      {firstError ? (
        <ErrorView error={firstError} title="个人 API Key 操作失败" />
      ) : null}
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

      <section className="grid gap-5 lg:grid-cols-[0.85fr_1.15fr]">
        <Card>
          <CardHeader
            title="创建个人 Key"
            description="user:* scopes 控制 /me 接口；tenant:* scopes 仍会被当前租户角色与 grant 收窄。"
          />
          <form className="space-y-4" onSubmit={submit}>
            <Input
              label="名称"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="my-automation"
              required
            />
            <Input
              label="过期时间，可选"
              type="datetime-local"
              value={expiresAt}
              onChange={(event) => setExpiresAt(event.target.value)}
            />
            <label className="flex items-start gap-2 rounded-panel border border-line bg-surface-soft px-3 py-2 text-sm text-muted">
              <input
                type="checkbox"
                checked={grantCurrentTenant}
                disabled={!currentTenantId}
                onChange={(event) =>
                  setGrantCurrentTenant(event.target.checked)
                }
              />
              <span>
                允许此 Key 访问当前租户
                <span className="block text-xs text-subtle">
                  {currentTenantId
                    ? `tenant_id: ${currentTenantId}`
                    : "未选择租户时只创建用户级 key；如需访问租户资源，请先选择租户后重新创建。"}
                </span>
              </span>
            </label>
            <div>
              <div className="mb-2 text-xs font-bold uppercase tracking-wide text-muted">
                Scopes
              </div>
              <div className="grid gap-2 sm:grid-cols-2">
                {knownScopes.map((scope) => (
                  <label
                    key={scope}
                    className="flex items-center gap-2 rounded-panel border border-line bg-surface-soft px-3 py-2 text-sm text-muted"
                  >
                    <input
                      type="checkbox"
                      checked={scopes.includes(scope)}
                      onChange={() => toggleScope(scope)}
                    />
                    {scope}
                  </label>
                ))}
              </div>
            </div>
            <Button
              type="submit"
              disabled={scopes.length === 0}
              isLoading={createAPIKey.isPending}
              leftIcon={<Plus className="size-4" />}
            >
              创建个人 Key
            </Button>
          </form>
        </Card>

        <Card>
          <CardHeader
            title="我的 Key 列表"
            description="调用 GET /api/v1/api-keys，raw key 不会再次出现。"
          />
          {apiKeys.isLoading ? (
            <LoadingView label="加载个人 API Keys..." />
          ) : (apiKeys.data?.api_keys.length ?? 0) === 0 ? (
            <EmptyState
              title="暂无个人 API Key"
              description="创建后可用于 Bearer 认证，并通过 X-Tenant-ID 选择租户上下文。"
            />
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <thead>
                  <tr>
                    <Th>名称</Th>
                    <Th>前缀</Th>
                    <Th>Scopes</Th>
                    <Th>租户授权</Th>
                    <Th>状态</Th>
                    <Th>操作</Th>
                  </tr>
                </thead>
                <tbody>
                  {apiKeys.data?.api_keys.map((item) => (
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
                        <div className="space-y-1 text-xs text-muted">
                          {(item.tenant_grants?.length ?? 0) === 0 ? (
                            <span>无</span>
                          ) : null}
                          {item.tenant_grants?.map((grant) => (
                            <div key={grant.id}>
                              <Badge tone={grant.revoked_at ? "red" : "green"}>
                                {grant.revoked_at ? "revoked" : "active"}
                              </Badge>{" "}
                              {grant.tenant_name ||
                                grant.tenant_slug ||
                                grant.tenant_id}
                            </div>
                          ))}
                        </div>
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
                  ))}
                </tbody>
              </Table>
            </div>
          )}
        </Card>
      </section>
    </div>
  );
}
