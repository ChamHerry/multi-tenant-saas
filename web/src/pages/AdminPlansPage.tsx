import { FormEvent, useState } from 'react'
import { useAdminBillingMutations, useAdminPlans } from '@/features/admin/admin-hooks'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Dialog } from '@/shared/ui/Dialog'
import { Input } from '@/shared/ui/Input'
import { Select } from '@/shared/ui/Select'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

const plans = ['free', 'pro', 'enterprise']

function positiveIntOrUndefined(value: string) {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined
}

export function AdminPlansPage() {
  const [plan, setPlan] = useState('')
  const [tenantId, setTenantId] = useState('')
  const [targetPlan, setTargetPlan] = useState('pro')
  const [maxMembers, setMaxMembers] = useState('')
  const [isPlanDialogOpen, setIsPlanDialogOpen] = useState(false)
  const [isQuotaDialogOpen, setIsQuotaDialogOpen] = useState(false)
  const planQuery = useAdminPlans({ plan })
  const mutations = useAdminBillingMutations()
  const firstError = planQuery.error ?? mutations.updateTenantPlan.error ?? mutations.updateTenantQuota.error

  const updatePlan = (event: FormEvent) => {
    event.preventDefault()
    mutations.updateTenantPlan.mutate(
      { tenantId: tenantId.trim(), plan: targetPlan },
      { onSuccess: () => setIsPlanDialogOpen(false) },
    )
  }

  const updateQuota = (event: FormEvent) => {
    event.preventDefault()
    mutations.updateTenantQuota.mutate(
      {
        tenantId: tenantId.trim(),
        quota: {
          max_members: positiveIntOrUndefined(maxMembers),
        },
      },
      { onSuccess: () => setIsQuotaDialogOpen(false) },
    )
  }

  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title="套餐/配额操作失败" /> : null}

      <div className="flex flex-wrap justify-end gap-2">
        <Button onClick={() => setIsPlanDialogOpen(true)}>更新套餐</Button>
        <Button variant="secondary" onClick={() => setIsQuotaDialogOpen(true)}>更新资源配额</Button>
      </div>

      <Dialog open={isPlanDialogOpen} title="单组织套餐" onClose={() => setIsPlanDialogOpen(false)}>
        <form className="space-y-4" onSubmit={updatePlan}>
          <Input label="组织 ID" value={tenantId} onChange={(event) => setTenantId(event.target.value)} required />
          <Select label="目标 Plan" value={targetPlan} onChange={(event) => setTargetPlan(event.target.value)}>{plans.map((item) => <option key={item} value={item}>{item}</option>)}</Select>
          <Button type="submit" isLoading={mutations.updateTenantPlan.isPending}>更新套餐</Button>
        </form>
      </Dialog>

      <Dialog open={isQuotaDialogOpen} title="单组织资源配额" onClose={() => setIsQuotaDialogOpen(false)}>
        <form className="space-y-4" onSubmit={updateQuota}>
          <Input label="组织 ID" value={tenantId} onChange={(event) => setTenantId(event.target.value)} required />
          <Input label="Max members" type="number" min={1} value={maxMembers} onChange={(event) => setMaxMembers(event.target.value)} />
          <Button type="submit" variant="secondary" isLoading={mutations.updateTenantQuota.isPending}>更新资源配额</Button>
        </form>
      </Dialog>

      <section>
        <Card>
          <CardHeader title="套餐权益" description={`Total: ${planQuery.data?.total ?? 0}`} />
          <div className="mb-4 max-w-xs"><Select label="Plan 过滤" value={plan} onChange={(event) => setPlan(event.target.value)}><option value="">全部</option>{plans.map((item) => <option key={item} value={item}>{item}</option>)}</Select></div>
          {planQuery.isLoading ? <LoadingView label="加载套餐权益..." /> : <div className="overflow-x-auto"><Table><thead><tr><Th>Plan</Th><Th>Feature</Th><Th>开关</Th><Th>限制</Th></tr></thead><tbody>{planQuery.data?.items.map((item) => <tr key={`${item.plan}:${item.feature_key}`}><Td><Badge>{item.plan}</Badge></Td><Td>{item.feature_key}</Td><Td><Badge tone={item.enabled ? 'green' : 'orange'}>{item.enabled ? 'enabled' : 'disabled'}</Badge></Td><Td>{item.limit_value ?? 'unlimited'}</Td></tr>)}</tbody></Table></div>}
        </Card>
      </section>
    </div>
  )
}
