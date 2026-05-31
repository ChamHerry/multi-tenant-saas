import { useState } from 'react'
import { Download } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useExportTenantAuditLogs, useTenantAuditLogs } from '@/features/audit/audit-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { useDateTimeFormatter } from '@/shared/i18n'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { EmptyState } from '@/shared/ui/EmptyState'
import { Input } from '@/shared/ui/Input'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function TenantAuditPage() {
  const { t } = useTranslation()
  const dateTimeFormatter = useDateTimeFormatter()
  const tenantId = useTenantStore((state) => state.currentTenantId)
  const [action, setAction] = useState('')
  const logs = useTenantAuditLogs(tenantId, { action })
  const exportLogs = useExportTenantAuditLogs(tenantId)
  if (!tenantId) return <EmptyState title={t('audit.tenant.requireTenantTitle')} description={t('audit.tenant.requireTenantDescription')} />
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
      <div className="flex justify-end"><Button variant="secondary" onClick={() => void download()} isLoading={exportLogs.isPending} leftIcon={<Download className="size-4" />}>{t('audit.common.exportJsonl')}</Button></div>
      {logs.error || exportLogs.error ? <ErrorView error={logs.error ?? exportLogs.error} title={t('audit.tenant.loadFailed')} /> : null}
      <Card>
        <CardHeader title={t('audit.common.filter')} />
        <Input label={t('audit.common.headers.action')} value={action} onChange={(event) => setAction(event.target.value)} placeholder={t('audit.tenant.actionPlaceholder')} />
      </Card>
      <Card>
        <CardHeader title={t('audit.common.logs')} description={t('common.total', { count: logs.data?.total ?? 0 })} />
        {logs.isLoading ? <LoadingView label={t('audit.common.loading')} /> : (logs.data?.logs.length ?? 0) === 0 ? <EmptyState title={t('audit.common.noLogs')} /> : (
          <div className="overflow-x-auto"><Table><thead><tr><Th>{t('audit.common.headers.time')}</Th><Th>{t('audit.common.headers.action')}</Th><Th>{t('audit.common.headers.resource')}</Th><Th>{t('audit.common.headers.user')}</Th><Th>{t('audit.common.headers.metadata')}</Th></tr></thead><tbody>{logs.data?.logs.map((item) => <tr key={item.id}><Td>{dateTimeFormatter.format(new Date(item.created_at))}</Td><Td><Badge>{item.action}</Badge></Td><Td>{item.resource_type}<div className="text-xs text-subtle">{item.resource_id}</div></Td><Td>{item.user_id}</Td><Td><code className="text-xs">{JSON.stringify(item.metadata)}</code></Td></tr>)}</tbody></Table></div>
        )}
      </Card>
    </div>
  )
}
