import { useState } from 'react'
import { Shield } from 'lucide-react'
import { useAccess, useCanPlatform } from '@/features/access/access-hooks'
import { useAdminTenantActions, useAdminTenants } from '@/features/admin/admin-hooks'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Input } from '@/shared/ui/Input'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function AdminTenantsPage() {
  const [query, setQuery] = useState('')
  const params = { query }
  const access = useAccess()
  const canManageTenants = useCanPlatform('platform:tenant:manage')
  const tenants = useAdminTenants(params)
  const actions = useAdminTenantActions(params)
  const firstError = access.error ?? tenants.error ?? actions.suspend.error ?? actions.restore.error
  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title="平台管理不可用" /> : null}
      <Card><CardHeader title="管理员会话" />{access.isLoading ? <LoadingView label="校验平台管理员..." /> : <div className="text-sm text-muted">角色：<Badge tone="red">{access.data?.platform_admin?.role}</Badge></div>}</Card>
      <Card><CardHeader title="过滤" /><Input label="搜索" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="tenant name / slug / id" /></Card>
      <Card>
        <CardHeader title="组织列表" description={`Total: ${tenants.data?.total ?? 0}`} />
        {tenants.isLoading ? <LoadingView label="加载组织..." /> : <div className="overflow-x-auto"><Table><thead><tr><Th>组织</Th><Th>状态</Th><Th>操作</Th></tr></thead><tbody>{tenants.data?.items.map((tenant) => <tr key={tenant.id}><Td><div className="font-bold text-ink">{tenant.name}</div><div className="text-xs text-subtle">{tenant.id}</div></Td><Td><Badge tone={tenant.status === 'active' ? 'green' : 'orange'}>{tenant.status}</Badge></Td><Td><div className="flex flex-wrap gap-2"><Button size="sm" variant="secondary" leftIcon={<Shield className="size-3" />} onClick={() => actions.suspend.mutate(tenant.id)} disabled={!canManageTenants || tenant.status !== 'active'}>暂停</Button><Button size="sm" variant="secondary" onClick={() => actions.restore.mutate(tenant.id)} disabled={!canManageTenants || tenant.status === 'active'}>恢复</Button></div></Td></tr>)}</tbody></Table></div>}
      </Card>
    </div>
  )
}
