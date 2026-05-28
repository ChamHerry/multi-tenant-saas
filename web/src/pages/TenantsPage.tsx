import { FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'
import { Building2, Plus } from 'lucide-react'
import { useMyTenants } from '@/features/auth/auth-hooks'
import { useCreateTenant } from '@/features/tenants/tenant-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Input } from '@/shared/ui/Input'
import { Select } from '@/shared/ui/Select'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

function slugify(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 80)
}

export function TenantsPage() {
  const tenants = useMyTenants()
  const createTenant = useCreateTenant()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const setCurrentTenantId = useTenantStore((state) => state.setCurrentTenantId)
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [plan, setPlan] = useState('free')

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createTenant.mutate(
      { name: name.trim(), slug: slug.trim(), plan },
      {
        onSuccess: () => {
          setName('')
          setSlug('')
          setPlan('free')
        },
      },
    )
  }

  if (tenants.isLoading) return <LoadingView label="加载组织列表..." />

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <Badge>Organizations</Badge>
          <h1 className="mt-3 text-3xl font-black text-ink">组织管理</h1>
          <p className="mt-2 text-sm leading-6 text-muted">一个用户可以属于多个组织。注册后系统会自动准备默认组织，进入组织功能时后端会按需完成初始化。</p>
        </div>
      </div>

      {tenants.error ? <ErrorView error={tenants.error} title="组织列表加载失败" /> : null}
      {createTenant.error ? <ErrorView error={createTenant.error} title="组织创建失败" /> : null}

      <section className="grid gap-5 lg:grid-cols-[0.85fr_1.15fr]">
        <Card>
          <CardHeader title="创建组织" description="创建后会立即出现在组织列表中；后端会在首次使用时自动准备运行环境。" />
          <TenantCreateForm
            name={name}
            slug={slug}
            plan={plan}
            isLoading={createTenant.isPending}
            onNameChange={(value) => {
              setName(value)
              if (!slug) setSlug(slugify(value))
            }}
            onSlugChange={(value) => setSlug(slugify(value))}
            onPlanChange={setPlan}
            onSubmit={submit}
          />
        </Card>

        <Card>
          <CardHeader title="我的组织" description="选择一个组织后，成员、API Key、代码分析等功能会使用该上下文。" />
          <div className="overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <Th>组织</Th>
                  <Th>角色</Th>
                  <Th>状态</Th>
                  <Th>操作</Th>
                </tr>
              </thead>
              <tbody>
                {tenants.data?.tenants.map((tenant) => (
                  <tr key={tenant.tenant_id}>
                    <Td>
                      <div className="font-bold text-ink">{tenant.tenant_name}</div>
                      <div className="mt-1 text-xs text-subtle">{tenant.tenant_slug}</div>
                    </Td>
                    <Td><Badge tone={tenant.role === 'owner' ? 'green' : tenant.role === 'viewer' ? 'gray' : 'purple'}>{tenant.role}</Badge></Td>
                    <Td><Badge tone={tenant.tenant_status === 'active' ? 'green' : 'orange'}>{tenant.tenant_status}</Badge></Td>
                    <Td>
                      <div className="flex gap-2">
                        <Button size="sm" variant={currentTenantId === tenant.tenant_id ? 'primary' : 'secondary'} onClick={() => setCurrentTenantId(tenant.tenant_id)}>
                          {currentTenantId === tenant.tenant_id ? '当前' : '切换'}
                        </Button>
                        <Link to={`/tenants/${tenant.tenant_id}`}><Button size="sm" variant="ghost" leftIcon={<Building2 className="size-3" />}>详情</Button></Link>
                      </div>
                    </Td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        </Card>
      </section>
    </div>
  )
}

type TenantCreateFormProps = {
  name: string
  slug: string
  plan: string
  isLoading?: boolean
  onNameChange: (value: string) => void
  onSlugChange: (value: string) => void
  onPlanChange: (value: string) => void
  onSubmit: (event: FormEvent) => void
}

function TenantCreateForm({ name, slug, plan, isLoading, onNameChange, onSlugChange, onPlanChange, onSubmit }: TenantCreateFormProps) {
  return (
    <form className="space-y-4" onSubmit={onSubmit}>
      <Input
        label="组织名称"
        value={name}
        onChange={(event) => onNameChange(event.target.value)}
        placeholder="RepoMind Demo"
        required
      />
      <Input label="Slug" value={slug} onChange={(event) => onSlugChange(event.target.value)} placeholder="repomind-demo" required />
      <Select label="套餐" value={plan} onChange={(event) => onPlanChange(event.target.value)}>
        <option value="free">free</option>
        <option value="pro">pro</option>
        <option value="enterprise">enterprise</option>
      </Select>
      <Button type="submit" isLoading={isLoading} leftIcon={<Plus className="size-4" />}>创建并切换</Button>
    </form>
  )
}
