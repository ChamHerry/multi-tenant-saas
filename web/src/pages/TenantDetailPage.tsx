import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Save, ShieldAlert } from 'lucide-react'
import { useCanTenant } from '@/features/access/access-hooks'
import { useTenantAction, useTenantDetail, useUpdateTenant } from '@/features/tenants/tenant-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import type { Tenant } from '@/features/tenants/tenant-types'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Dialog } from '@/shared/ui/Dialog'
import { usePageTitle } from '@/layouts/page-title'
import { Input } from '@/shared/ui/Input'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'

function TenantEditForm({ tenantId, tenant, canManage, onSaved }: { tenantId: string; tenant: Tenant; canManage: boolean; onSaved?: () => void }) {
  const updateTenant = useUpdateTenant(tenantId)
  const [name, setName] = useState(tenant.name)
  const [slug, setSlug] = useState(tenant.slug)

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateTenant.mutate(
      {
        name: name.trim(),
        slug: slug.trim(),
      },
      { onSuccess: onSaved },
    )
  }

  return (
    <>
      {updateTenant.error ? <ErrorView error={updateTenant.error} title="组织更新失败" /> : null}
      <form className="grid gap-4 sm:grid-cols-2" onSubmit={submit}>
        <Input label="名称" value={name} onChange={(event) => setName(event.target.value)} disabled={!canManage} />
        <Input label="Slug" value={slug} onChange={(event) => setSlug(event.target.value)} disabled={!canManage} />
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
  const [isEditOpen, setIsEditOpen] = useState(false)

  usePageTitle(tenant?.name ?? tenantId ?? '组织详情')

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
      <div className="flex flex-wrap justify-end gap-2">
        <Button disabled={!tenant || !canManageTenant} onClick={() => setIsEditOpen(true)}>编辑基础信息</Button>
        <Button variant="secondary" onClick={() => navigate('/tenants')}>返回组织列表</Button>
      </div>

      {firstError ? <ErrorView error={firstError} title="组织操作失败" /> : null}

      <Dialog open={isEditOpen} title="编辑基础信息" onClose={() => setIsEditOpen(false)}>
        {tenant ? <TenantEditForm key={tenant.id} tenantId={tenantId} tenant={tenant} canManage={canManageTenant} onSaved={() => setIsEditOpen(false)} /> : <ErrorView title="组织不存在" error="后端未返回组织数据" />}
      </Dialog>

      <section>
        <Card>
          <CardHeader title="模板能力边界" description="当前项目只保留通用多租户 SaaS 管理能力，不包含具体业务域数据模型。" />
          <div className="space-y-3 text-sm text-muted">
            <div className="rounded-panel bg-surface-soft p-3"><span className="font-bold text-ink">Access boundary</span><p>成员角色、API Key scope、平台管理员权限共同控制组织管理入口。</p></div>
            <div className="rounded-panel bg-surface-soft p-3"><span className="font-bold text-ink">Extension point</span><p>后续业务表应显式绑定 tenant_id，并复用现有 RBAC 和审计服务。</p></div>
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
