import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAdminAuditLogs } from '@/features/admin/admin-hooks'
import { useDateTimeFormatter } from '@/shared/i18n'
import { Badge } from '@/shared/ui/Badge'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Input } from '@/shared/ui/Input'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function AdminAuditPage() {
  const { t } = useTranslation()
  const dateTimeFormatter = useDateTimeFormatter()
  const [tenantId, setTenantId] = useState('')
  const [action, setAction] = useState('')
  const logs = useAdminAuditLogs({ tenant_id: tenantId, action, limit: 100 })
  return (
    <div className="space-y-6">
      {logs.error ? <ErrorView error={logs.error} title={t('audit.admin.loadFailed')} /> : null}
      <Card><CardHeader title={t('audit.common.filter')} /><div className="grid gap-4 md:grid-cols-2"><Input label={t('audit.admin.tenantIdLabel')} value={tenantId} onChange={(event) => setTenantId(event.target.value)} /><Input label={t('audit.common.headers.action')} value={action} onChange={(event) => setAction(event.target.value)} /></div></Card>
      <Card><CardHeader title={t('audit.common.logs')} description={t('common.total', { count: logs.data?.total ?? 0 })} />{logs.isLoading ? <LoadingView label={t('audit.admin.loading')} /> : <div className="overflow-x-auto"><Table><thead><tr><Th>{t('audit.common.headers.time')}</Th><Th>{t('audit.common.headers.organization')}</Th><Th>{t('audit.common.headers.user')}</Th><Th>{t('audit.common.headers.action')}</Th><Th>{t('audit.common.headers.resource')}</Th></tr></thead><tbody>{logs.data?.logs.map((item) => <tr key={item.id}><Td>{dateTimeFormatter.format(new Date(item.created_at))}</Td><Td>{item.tenant_id}</Td><Td>{item.user_id}</Td><Td><Badge>{item.action}</Badge></Td><Td>{item.resource_type}<div className="text-xs text-subtle">{item.resource_id}</div></Td></tr>)}</tbody></Table></div>}</Card>
    </div>
  )
}
