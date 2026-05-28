import { useState } from 'react'
import { Download } from 'lucide-react'
import { useExportTenantAuditLogs, useTenantAuditLogs } from '@/features/audit/audit-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { EmptyState } from '@/shared/ui/EmptyState'
import { Input } from '@/shared/ui/Input'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function TenantAuditPage() {
  const tenantId = useTenantStore((state) => state.currentTenantId)
  const [action, setAction] = useState('')
  const logs = useTenantAuditLogs(tenantId, { action })
  const exportLogs = useExportTenantAuditLogs(tenantId)
  if (!tenantId) return <EmptyState title="请先选择组织" description="审计查询需要当前组织上下文。" />
  const download = async () => {
    const result = await exportLogs.mutateAsync({ action })
    const content = result.job.content ?? ''
    const blob = new Blob([content], { type: result.job.content_type ?? 'application/x-ndjson' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = result.job.file_name ?? `tenant-${tenantId}-audit.jsonl`
    link.click()
    URL.revokeObjectURL(url)
  }
  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end"><div><Badge tone="orange">Audit</Badge><h1 className="mt-3 text-3xl font-black text-ink">组织审计日志</h1><p className="mt-2 text-sm text-muted">按 action / 时间 / 用户过滤组织范围内的管理事件。</p></div><Button variant="secondary" onClick={() => void download()} isLoading={exportLogs.isPending} leftIcon={<Download className="size-4" />}>导出 JSONL</Button></div>
      {logs.error || exportLogs.error ? <ErrorView error={logs.error ?? exportLogs.error} title="审计日志加载失败" /> : null}
      <Card>
        <CardHeader title="过滤" />
        <Input label="Action" value={action} onChange={(event) => setAction(event.target.value)} placeholder="tenant.invitation.create" />
      </Card>
      <Card>
        <CardHeader title="日志" description={`Total: ${logs.data?.total ?? 0}`} />
        {logs.isLoading ? <LoadingView label="加载审计日志..." /> : (logs.data?.logs.length ?? 0) === 0 ? <EmptyState title="暂无日志" /> : (
          <div className="overflow-x-auto"><Table><thead><tr><Th>时间</Th><Th>Action</Th><Th>资源</Th><Th>用户</Th><Th>Metadata</Th></tr></thead><tbody>{logs.data?.logs.map((item) => <tr key={item.id}><Td>{new Date(item.created_at).toLocaleString()}</Td><Td><Badge>{item.action}</Badge></Td><Td>{item.resource_type}<div className="text-xs text-subtle">{item.resource_id}</div></Td><Td>{item.user_id}</Td><Td><code className="text-xs">{JSON.stringify(item.metadata)}</code></Td></tr>)}</tbody></Table></div>
        )}
      </Card>
    </div>
  )
}
