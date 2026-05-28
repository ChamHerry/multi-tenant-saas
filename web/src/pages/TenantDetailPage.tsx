import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Save, ShieldAlert } from 'lucide-react'
import { useCanTenant } from '@/features/access/access-hooks'
import { useTenantAction, useTenantDetail, useUpdateTenant } from '@/features/tenants/tenant-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import type { Tenant } from '@/features/tenants/tenant-types'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Input } from '@/shared/ui/Input'
import { Select } from '@/shared/ui/Select'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'

function TenantEditForm({ tenantId, tenant, canManage }: { tenantId: string; tenant: Tenant; canManage: boolean }) {
  const updateTenant = useUpdateTenant(tenantId)
  const [name, setName] = useState(tenant.name)
  const [slug, setSlug] = useState(tenant.slug)
  const [plan, setPlan] = useState(tenant.plan)
  const [maxRepos, setMaxRepos] = useState(String(tenant.max_repos))
  const [maxSymbols, setMaxSymbols] = useState(String(tenant.max_symbols))
  const [maxStorage, setMaxStorage] = useState(String(tenant.max_storage_mb))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateTenant.mutate({
      name: name.trim(),
      slug: slug.trim(),
      plan,
      max_repos: Number(maxRepos) || undefined,
      max_symbols: Number(maxSymbols) || undefined,
      max_storage_mb: Number(maxStorage) || undefined,
    })
  }

  return (
    <>
      {updateTenant.error ? <ErrorView error={updateTenant.error} title="组织更新失败" /> : null}
      <form className="grid gap-4 sm:grid-cols-2" onSubmit={submit}>
        <Input label="名称" value={name} onChange={(event) => setName(event.target.value)} disabled={!canManage} />
        <Input label="Slug" value={slug} onChange={(event) => setSlug(event.target.value)} disabled={!canManage} />
        <Select label="套餐" value={plan} onChange={(event) => setPlan(event.target.value)} disabled={!canManage}>
          <option value="free">free</option>
          <option value="pro">pro</option>
          <option value="enterprise">enterprise</option>
        </Select>
        <Input label="Max Repos" type="number" value={maxRepos} onChange={(event) => setMaxRepos(event.target.value)} disabled={!canManage} />
        <Input label="Max Symbols" type="number" value={maxSymbols} onChange={(event) => setMaxSymbols(event.target.value)} disabled={!canManage} />
        <Input label="Max Storage MB" type="number" value={maxStorage} onChange={(event) => setMaxStorage(event.target.value)} disabled={!canManage} />
        <div className="sm:col-span-2">
          <Button type="submit" disabled={!canManage} isLoading={updateTenant.isPending} leftIcon={<Save className="size-4" />}>保存修改</Button>
        </div>
      </form>
    </>
  )
}

export function TenantDetailPage() {
  const { tenantId = '' } = useParams()
  const navigate = useNavigate()
  const setCurrentTenantId = useTenantStore((state) => state.setCurrentTenantId)
  const detail = useTenantDetail(tenantId)
  const actions = useTenantAction(tenantId)
  const canManageTenant = useCanTenant('tenant:manage', tenantId)
  const tenant = detail.data?.tenant

  useEffect(() => {
    if (tenantId) setCurrentTenantId(tenantId)
  }, [setCurrentTenantId, tenantId])

  const firstError = useMemo(
    () => detail.error ?? actions.suspend.error ?? actions.restore.error ?? actions.remove.error,
    [actions.remove.error, actions.restore.error, actions.suspend.error, detail.error],
  )

  if (!tenantId) return <ErrorView title="缺少组织 ID" error="URL 中没有组织 ID" />
  if (detail.isLoading) return <LoadingView label="加载组织详情..." />

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <Badge tone="purple">Tenant Detail</Badge>
          <h1 className="mt-3 text-3xl font-black text-ink">{tenant?.name ?? tenantId}</h1>
          <p className="mt-2 text-sm text-muted">查看/更新组织基础信息；底层仓库和索引统一存放在 public schema。</p>
        </div>
        <Button variant="secondary" onClick={() => navigate('/tenants')}>返回组织列表</Button>
      </div>

      {firstError ? <ErrorView error={firstError} title="组织操作失败" /> : null}

      <section className="grid gap-5 lg:grid-cols-[1fr_0.9fr]">
        <Card>
          <CardHeader title="基础信息" description={canManageTenant ? '调用 PATCH /api/v1/tenants/{tenant}。owner 可以更新。' : '当前角色只有查看权限，不能更新组织基础信息。'} />
          {tenant ? <TenantEditForm key={tenant.id} tenantId={tenantId} tenant={tenant} canManage={canManageTenant} /> : <ErrorView title="组织不存在" error="后端未返回组织数据" />}
        </Card>

        <Card>
          <CardHeader title="共享仓库模型" description="组织只作为 RBAC、配额和审计边界；代码仓库和索引数据统一在 public schema 中共享。" />
          <div className="space-y-3 text-sm text-muted">
            <div className="rounded-panel bg-surface-soft p-3"><span className="font-bold text-ink">Public repo</span><p>公开仓库底层索引可被所有有登录上下文的组织复用，不再为每个组织创建 schema/graph。</p></div>
            <div className="rounded-panel bg-surface-soft p-3"><span className="font-bold text-ink">Private repo</span><p>私有仓库必须通过仓库授权策略裁剪，禁止绕过 repository access predicate 直查索引表。</p></div>
          </div>
        </Card>
      </section>

      <Card>
        <CardHeader title="危险操作" description="这些操作依赖 tenant:manage 权限，并会写审计日志。" />
        <div className="flex flex-wrap gap-3">
          <Button variant="secondary" disabled={!canManageTenant} onClick={() => actions.suspend.mutate()} isLoading={actions.suspend.isPending} leftIcon={<ShieldAlert className="size-4" />}>暂停组织</Button>
          <Button variant="secondary" disabled={!canManageTenant} onClick={() => actions.restore.mutate()} isLoading={actions.restore.isPending}>恢复组织</Button>
          <Button variant="danger" disabled={!canManageTenant} onClick={() => actions.remove.mutate()} isLoading={actions.remove.isPending}>软删除组织</Button>
        </div>
      </Card>
    </div>
  )
}
