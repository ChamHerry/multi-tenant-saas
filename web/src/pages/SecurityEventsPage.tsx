import { useTranslation } from 'react-i18next'
import { useSecurityEvents } from '@/features/audit/audit-hooks'
import { useDateTimeFormatter } from '@/shared/i18n'
import { Badge, Card, CardHeader, EmptyState, ErrorView, LoadingView, Table, Td, Th } from '@/shared/ui'

export function SecurityEventsPage() {
  const { t } = useTranslation()
  const dateTimeFormatter = useDateTimeFormatter()
  const logs = useSecurityEvents({ limit: 100 })
  return (
    <div className="space-y-6">
      {logs.error ? <ErrorView error={logs.error} title={t('audit.security.loadFailed')} /> : null}
      <Card><CardHeader title={t('audit.security.title')} />{logs.isLoading ? <LoadingView label={t('audit.security.loading')} /> : (logs.data?.logs.length ?? 0) === 0 ? <EmptyState title={t('audit.security.empty')} /> : <div className="overflow-x-auto"><Table><thead><tr><Th>{t('audit.common.headers.time')}</Th><Th>{t('audit.common.headers.action')}</Th><Th>{t('audit.common.headers.ip')}</Th><Th>{t('audit.common.headers.userAgent')}</Th></tr></thead><tbody>{logs.data?.logs.map((item) => <tr key={item.id}><Td>{dateTimeFormatter.format(new Date(item.created_at))}</Td><Td><Badge>{item.action}</Badge></Td><Td>{item.ip}</Td><Td className="max-w-xl break-all">{item.user_agent}</Td></tr>)}</tbody></Table></div>}</Card>
    </div>
  )
}
