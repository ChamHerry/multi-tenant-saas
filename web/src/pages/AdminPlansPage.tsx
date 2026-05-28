import { FormEvent, useState } from 'react'
import { useAdminBillingMutations, useAdminPlans } from '@/features/admin/admin-hooks'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
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
  const [maxRepos, setMaxRepos] = useState('')
  const [maxSymbols, setMaxSymbols] = useState('')
  const [maxStorageMB, setMaxStorageMB] = useState('')
  const planQuery = useAdminPlans({ plan })
  const mutations = useAdminBillingMutations()
  const firstError = planQuery.error ?? mutations.updateTenantPlan.error ?? mutations.updateTenantQuota.error

  const updatePlan = (event: FormEvent) => {
    event.preventDefault()
    mutations.updateTenantPlan.mutate({ tenantId: tenantId.trim(), plan: targetPlan })
  }

  const updateQuota = (event: FormEvent) => {
    event.preventDefault()
    mutations.updateTenantQuota.mutate({
      tenantId: tenantId.trim(),
      quota: {
        max_repos: positiveIntOrUndefined(maxRepos),
        max_symbols: positiveIntOrUndefined(maxSymbols),
        max_storage_mb: positiveIntOrUndefined(maxStorageMB),
      },
    })
  }

  return (
    <div className="space-y-6">
      <div>
        <Badge tone="red">Billing Admin</Badge>
        <h1 className="mt-3 text-3xl font-black text-ink">套餐与配额管理</h1>
        <p className="mt-2 text-sm text-muted">查看内置 plan entitlement，并对单组织执行套餐/配额覆盖。</p>
      </div>
      {firstError ? <ErrorView error={firstError} title="套餐/配额操作失败" /> : null}
      <section className="grid gap-5 xl:grid-cols-[1.1fr_0.9fr]">
        <Card>
          <CardHeader title="套餐权益" description={`Total: ${planQuery.data?.total ?? 0}`} />
          <div className="mb-4 max-w-xs"><Select label="Plan 过滤" value={plan} onChange={(event) => setPlan(event.target.value)}><option value="">全部</option>{plans.map((item) => <option key={item} value={item}>{item}</option>)}</Select></div>
          {planQuery.isLoading ? <LoadingView label="加载套餐权益..." /> : <div className="overflow-x-auto"><Table><thead><tr><Th>Plan</Th><Th>Feature</Th><Th>开关</Th><Th>限制</Th></tr></thead><tbody>{planQuery.data?.items.map((item) => <tr key={`${item.plan}:${item.feature_key}`}><Td><Badge>{item.plan}</Badge></Td><Td>{item.feature_key}</Td><Td><Badge tone={item.enabled ? 'green' : 'orange'}>{item.enabled ? 'enabled' : 'disabled'}</Badge></Td><Td>{item.limit_value ?? 'unlimited'}</Td></tr>)}</tbody></Table></div>}
        </Card>
        <div className="space-y-5">
          <Card>
            <CardHeader title="单组织套餐" description="更新 public.tenants.plan 与 active subscription。" />
            <form className="space-y-4" onSubmit={updatePlan}>
              <Input label="组织 ID" value={tenantId} onChange={(event) => setTenantId(event.target.value)} required />
              <Select label="目标 Plan" value={targetPlan} onChange={(event) => setTargetPlan(event.target.value)}>{plans.map((item) => <option key={item} value={item}>{item}</option>)}</Select>
              <Button type="submit" isLoading={mutations.updateTenantPlan.isPending}>更新套餐</Button>
            </form>
          </Card>
          <Card>
            <CardHeader title="单组织资源配额" description="只填写要覆盖的正整数；留空保持不变。" />
            <form className="space-y-4" onSubmit={updateQuota}>
              <Input label="组织 ID" value={tenantId} onChange={(event) => setTenantId(event.target.value)} required />
              <Input label="Max repos" type="number" min={1} value={maxRepos} onChange={(event) => setMaxRepos(event.target.value)} />
              <Input label="Max symbols" type="number" min={1} value={maxSymbols} onChange={(event) => setMaxSymbols(event.target.value)} />
              <Input label="Max storage MB" type="number" min={1} value={maxStorageMB} onChange={(event) => setMaxStorageMB(event.target.value)} />
              <Button type="submit" variant="secondary" isLoading={mutations.updateTenantQuota.isPending}>更新资源配额</Button>
            </form>
          </Card>
        </div>
      </section>
    </div>
  )
}
