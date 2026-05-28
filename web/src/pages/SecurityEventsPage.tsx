import { useSecurityEvents } from '@/features/audit/audit-hooks'
import { Badge } from '@/shared/ui/Badge'
import { Card, CardHeader } from '@/shared/ui/Card'
import { EmptyState } from '@/shared/ui/EmptyState'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function SecurityEventsPage() {
  const logs = useSecurityEvents({ limit: 100 })
  return (
    <div className="space-y-6">
      {logs.error ? <ErrorView error={logs.error} title="安全事件加载失败" /> : null}
      <Card><CardHeader title="安全事件" />{logs.isLoading ? <LoadingView label="加载安全事件..." /> : (logs.data?.logs.length ?? 0) === 0 ? <EmptyState title="暂无安全事件" /> : <div className="overflow-x-auto"><Table><thead><tr><Th>时间</Th><Th>Action</Th><Th>IP</Th><Th>User-Agent</Th></tr></thead><tbody>{logs.data?.logs.map((item) => <tr key={item.id}><Td>{new Date(item.created_at).toLocaleString()}</Td><Td><Badge>{item.action}</Badge></Td><Td>{item.ip}</Td><Td className="max-w-xl break-all">{item.user_agent}</Td></tr>)}</tbody></Table></div>}</Card>
    </div>
  )
}
