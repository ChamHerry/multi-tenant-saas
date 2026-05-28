import { FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'
import { Building2, Plus } from 'lucide-react'
import { useMyTenants } from '@/features/auth/auth-hooks'
import { useCreateTenant } from '@/features/tenants/tenant-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Dialog } from '@/shared/ui/Dialog'
import { Input } from '@/shared/ui/Input'
import { Select } from '@/shared/ui/Select'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

const TENANT_SLUG_MAX_LENGTH = 80
const TENANT_SLUG_PATTERN = /^[a-z0-9][a-z0-9-]{1,78}[a-z0-9]$/
const TENANT_SLUG_ERROR = 'Slug 需为 3-80 位小写字母、数字或短横线，且首尾必须是字母或数字'

function slugifyTenantSlug(value: string) {
  return value
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, TENANT_SLUG_MAX_LENGTH)
    .replace(/^-+|-+$/g, '')
}

export function TenantsPage() {
  const tenants = useMyTenants()
  const createTenant = useCreateTenant()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const setCurrentTenantId = useTenantStore((state) => state.setCurrentTenantId)
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [plan, setPlan] = useState('free')
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [isSlugManuallyEdited, setIsSlugManuallyEdited] = useState(false)

  const trimmedName = name.trim()
  const trimmedSlug = slug.trim()
  const slugError =
    trimmedName && !trimmedSlug
      ? 'Slug 不能为空；请使用英文小写字母、数字或短横线'
      : trimmedSlug && !TENANT_SLUG_PATTERN.test(trimmedSlug)
        ? TENANT_SLUG_ERROR
        : undefined
  const canSubmitCreateTenant = Boolean(trimmedName) && TENANT_SLUG_PATTERN.test(trimmedSlug)

  const resetCreateTenantForm = () => {
    setName('')
    setSlug('')
    setPlan('free')
    setIsSlugManuallyEdited(false)
  }

  const closeCreateDialog = () => {
    setIsCreateOpen(false)
    resetCreateTenantForm()
  }

  const handleNameChange = (value: string) => {
    setName(value)
    if (!isSlugManuallyEdited) {
      setSlug(slugifyTenantSlug(value))
    }
  }

  const handleSlugChange = (value: string) => {
    setIsSlugManuallyEdited(true)
    setSlug(slugifyTenantSlug(value))
  }

  const regenerateSlug = () => {
    setIsSlugManuallyEdited(false)
    setSlug(slugifyTenantSlug(name))
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!canSubmitCreateTenant) return

    createTenant.mutate(
      { name: trimmedName, slug: trimmedSlug, plan },
      {
        onSuccess: () => {
          resetCreateTenantForm()
          setIsCreateOpen(false)
        },
      },
    )
  }

  if (tenants.isLoading) return <LoadingView label="加载组织列表..." />

  return (
    <div className="space-y-6">
      {tenants.error ? <ErrorView error={tenants.error} title="组织列表加载失败" /> : null}
      {createTenant.error ? <ErrorView error={createTenant.error} title="组织创建失败" /> : null}

      <div className="flex justify-end">
        <Button leftIcon={<Plus className="size-4" />} onClick={() => setIsCreateOpen(true)}>
          添加组织
        </Button>
      </div>

      <Dialog open={isCreateOpen} title="添加组织" onClose={closeCreateDialog}>
        <form className="space-y-4" onSubmit={submit}>
          <Input
            label="组织名称"
            value={name}
            onChange={(event) => handleNameChange(event.target.value)}
            placeholder="Acme Demo"
            required
          />
          <div className="space-y-2">
            <Input
              label="Slug"
              value={slug}
              onChange={(event) => handleSlugChange(event.target.value)}
              placeholder="acme-demo"
              required
              error={slugError}
              hint="用于组织 URL/API 标识；可手动修改。"
            />
            <div className="flex justify-end">
              <Button type="button" variant="ghost" size="sm" onClick={regenerateSlug}>
                按名称生成
              </Button>
            </div>
          </div>
          <Select label="套餐" value={plan} onChange={(event) => setPlan(event.target.value)}>
            <option value="free">free</option>
            <option value="pro">pro</option>
            <option value="enterprise">enterprise</option>
          </Select>
          <Button type="submit" disabled={!canSubmitCreateTenant} isLoading={createTenant.isPending} leftIcon={<Plus className="size-4" />}>添加并切换</Button>
        </form>
      </Dialog>

      <section>
        <Card>
          <CardHeader title="我的组织" description="选择一个组织后，成员、邀请、API Key、审计和配额功能会使用该上下文。" />
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
