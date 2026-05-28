import { useState } from 'react'
import { useAdminAuditLogs } from '@/features/admin/admin-hooks'
import { Badge } from '@/shared/ui/Badge'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Input } from '@/shared/ui/Input'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function AdminAuditPage() {
  const [tenantId, setTenantId] = useState('')
  const [action, setAction] = useState('')
  const logs = useAdminAuditLogs({ tenant_id: tenantId, action, limit: 100 })
  return (
    <div className="space-y-6">
      <div><Badge tone="red">Platform Audit</Badge><h1 className="mt-3 text-3xl font-black text-ink">平台审计日志</h1><p className="mt-2 text-sm text-muted">平台 auditor/super_admin 可跨组织查询审计事件。</p></div>
      {logs.error ? <ErrorView error={logs.error} title="平台审计加载失败" /> : null}
      <Card><CardHeader title="过滤" /><div className="grid gap-4 md:grid-cols-2"><Input label="组织 ID" value={tenantId} onChange={(event) => setTenantId(event.target.value)} /><Input label="Action" value={action} onChange={(event) => setAction(event.target.value)} /></div></Card>
      <Card><CardHeader title="日志" description={`Total: ${logs.data?.total ?? 0}`} />{logs.isLoading ? <LoadingView label="加载审计..." /> : <div className="overflow-x-auto"><Table><thead><tr><Th>时间</Th><Th>组织</Th><Th>用户</Th><Th>Action</Th><Th>资源</Th></tr></thead><tbody>{logs.data?.logs.map((item) => <tr key={item.id}><Td>{new Date(item.created_at).toLocaleString()}</Td><Td>{item.tenant_id}</Td><Td>{item.user_id}</Td><Td><Badge>{item.action}</Badge></Td><Td>{item.resource_type}<div className="text-xs text-subtle">{item.resource_id}</div></Td></tr>)}</tbody></Table></div>}</Card>
    </div>
  )
}
