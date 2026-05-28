import { RotateCcw } from 'lucide-react'
import { useCanTenant } from '@/features/access/access-hooks'
import { useRecalculateTenantUsage, useTenantQuota } from '@/features/quota/quota-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { EmptyState } from '@/shared/ui/EmptyState'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'

export function TenantUsagePage() {
  const tenantId = useTenantStore((state) => state.currentTenantId)
  const quota = useTenantQuota(tenantId)
  const recalculate = useRecalculateTenantUsage(tenantId)
  const canManageBilling = useCanTenant('tenant:billing:manage', tenantId)
  if (!tenantId) return <EmptyState title="请先选择组织" description="用量统计需要当前组织上下文。" />
  const firstError = quota.error ?? recalculate.error
  return (
    <div className="space-y-6">
      <div className="flex justify-end">
        <Button
          variant="secondary"
          disabled={!canManageBilling}
          onClick={() => recalculate.mutate()}
          isLoading={recalculate.isPending}
          leftIcon={<RotateCcw className="size-4" />}
        >
          重新统计
        </Button>
      </div>
      {firstError ? <ErrorView error={firstError} title="配额加载失败" /> : null}
      {quota.isLoading ? <LoadingView label="加载配额..." /> : (
        <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          {quota.data?.quota.items.map((item) => {
            const used = item.used + item.reserved
            const percent = item.limit ? Math.min(100, Math.round((used / item.limit) * 100)) : 0
            return (
              <Card key={item.metric}>
                <CardHeader title={item.metric} description={`已用 ${item.used}，预留 ${item.reserved}`} action={<Badge tone={item.limit && used >= item.limit ? 'red' : 'blue'}>{item.limit ?? 'unlimited'}</Badge>} />
                <div className="h-2 overflow-hidden rounded-full bg-surface-soft"><div className="h-full rounded-full bg-brand" style={{ width: item.limit ? `${percent}%` : '12%' }} /></div>
                <div className="mt-3 text-sm text-muted">剩余：{item.remaining ?? '无限制'}</div>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
