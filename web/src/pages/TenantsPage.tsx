import { FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'
import { Building2, Plus } from 'lucide-react'
import TinyPinyin from 'tiny-pinyin'
import { useMyTenants } from '@/features/auth/auth-hooks'
import { useCreateTenant } from '@/features/tenants/tenant-hooks'
import { createTenantInputSchema, TENANT_SLUG_MAX_LENGTH } from '@/features/tenants/tenant-types'
import { useTenantStore } from '@/features/tenants/tenant-store'
import {
  Badge,
  Button,
  Card,
  CardHeader,
  Dialog,
  Input,
  ErrorView,
  LoadingView,
  Table,
  Td,
  Th,
} from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'

const CJK_TEXT_PATTERN = /[㐀-鿿豈-﫿]+/g

function slugifyTenantSlug(value: string) {
  return value
    .normalize('NFKD')
    .replace(/[̀-ͯ]/g, '')
    .replace(CJK_TEXT_PATTERN, (text) =>
      TinyPinyin.convertToPinyin(text, '-', true),
    )
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
  const setCurrentTenantId = useTenantStore(
    (state) => state.setCurrentTenantId,
  )
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [isSlugManuallyEdited, setIsSlugManuallyEdited] = useState(false)
  const [errors, setErrors] = useState<FieldErrors>({})

  const resetCreateTenantForm = () => {
    setName('')
    setSlug('')
    setIsSlugManuallyEdited(false)
    setErrors({})
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
    const result = validateForm(createTenantInputSchema, {
      name: name.trim(),
      slug: slug.trim(),
    })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    createTenant.mutate(result.data, {
      onSuccess: () => {
        resetCreateTenantForm()
        setIsCreateOpen(false)
      },
    })
  }

  if (tenants.isLoading) return <LoadingView label="加载组织列表..." />

  return (
    <div className="space-y-6">
      {tenants.error ? (
        <ErrorView error={tenants.error} title="组织列表加载失败" />
      ) : null}
      {createTenant.error ? (
        <ErrorView error={createTenant.error} title="组织创建失败" />
      ) : null}

      <div className="flex justify-end">
        <Button
          leftIcon={<Plus className="size-4" />}
          onClick={() => setIsCreateOpen(true)}
        >
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
            error={errors.name}
          />
          <div className="space-y-2">
            <Input
              label="Slug"
              value={slug}
              onChange={(event) => handleSlugChange(event.target.value)}
              placeholder="acme-demo"
              error={errors.slug}
              hint="用于组织 URL/API 标识；中文名称会自动转为拼音，可手动修改。"
            />
            <div className="flex justify-end">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={regenerateSlug}
              >
                按名称生成
              </Button>
            </div>
          </div>
          <Button
            type="submit"
            isLoading={createTenant.isPending}
            leftIcon={<Plus className="size-4" />}
          >
            添加并切换
          </Button>
        </form>
      </Dialog>

      <section>
        <Card>
          <CardHeader
            title="我的组织"
            description="选择一个组织后，成员、邀请、API Key 和审计功能会使用该上下文。"
          />
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
                      <div className="font-bold text-ink">
                        {tenant.tenant_name}
                      </div>
                      <div className="mt-1 text-xs text-subtle">
                        {tenant.tenant_slug}
                      </div>
                    </Td>
                    <Td>
                      <Badge
                        tone={
                          tenant.role === 'owner'
                            ? 'green'
                            : tenant.role === 'viewer'
                              ? 'gray'
                              : 'purple'
                        }
                      >
                        {tenant.role}
                      </Badge>
                    </Td>
                    <Td>
                      <Badge
                        tone={
                          tenant.tenant_status === 'active' ? 'green' : 'orange'
                        }
                      >
                        {tenant.tenant_status}
                      </Badge>
                    </Td>
                    <Td>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant={
                            currentTenantId === tenant.tenant_id
                              ? 'primary'
                              : 'secondary'
                          }
                          onClick={() => setCurrentTenantId(tenant.tenant_id)}
                        >
                          {currentTenantId === tenant.tenant_id
                            ? '当前'
                            : '切换'}
                        </Button>
                        <Link
                          to={`/tenant/settings?tenantId=${tenant.tenant_id}`}
                        >
                          <Button
                            size="sm"
                            variant="ghost"
                            leftIcon={<Building2 className="size-3" />}
                          >
                            设置
                          </Button>
                        </Link>
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
