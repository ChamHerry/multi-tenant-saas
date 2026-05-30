import { FormEvent, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Building2, Plus } from 'lucide-react'
import TinyPinyin from 'tiny-pinyin'
import { useMyTenants } from '@/features/auth/auth-hooks'
import { useCreateTenant } from '@/features/tenants/tenant-hooks'
import { useTenantI18n } from '@/features/tenants/tenant-i18n'
import { createCreateTenantInputSchema, TENANT_SLUG_MAX_LENGTH } from '@/features/tenants/tenant-types'
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

const CJK_TEXT_PATTERN = /[\u3400-\u9fff\uf900-\ufaff]+/g

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
  const { schemaTranslator, t, formatRole, formatTenantStatus } = useTenantI18n()
  const createTenantSchema = useMemo(() => createCreateTenantInputSchema(schemaTranslator), [schemaTranslator])

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
    const result = validateForm(createTenantSchema, {
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

  if (tenants.isLoading) return <LoadingView label={t('tenants.loading', 'Loading organizations...')} />

  return (
    <div className="space-y-6">
      {tenants.error ? (
        <ErrorView error={tenants.error} title={t('tenants.listErrorTitle', 'Organization list failed to load')} />
      ) : null}
      {createTenant.error ? (
        <ErrorView error={createTenant.error} title={t('tenants.createErrorTitle', 'Organization creation failed')} />
      ) : null}

      <div className="flex justify-end">
        <Button
          leftIcon={<Plus className="size-4" />}
          onClick={() => setIsCreateOpen(true)}
        >
          {t('tenants.addButton', 'Add organization')}
        </Button>
      </div>

      <Dialog open={isCreateOpen} title={t('tenants.dialogTitle', 'Add organization')} onClose={closeCreateDialog}>
        <form className="space-y-4" onSubmit={submit}>
          <Input
            label={t('tenants.nameLabel', 'Organization name')}
            value={name}
            onChange={(event) => handleNameChange(event.target.value)}
            placeholder={t('tenants.namePlaceholder', 'Acme Demo')}
            error={errors.name}
          />
          <div className="space-y-2">
            <Input
              label={t('tenants.slugLabel', 'Slug')}
              value={slug}
              onChange={(event) => handleSlugChange(event.target.value)}
              placeholder={t('tenants.slugPlaceholder', 'acme-demo')}
              error={errors.slug}
              hint={t('tenants.slugHint', 'Used for organization URL/API identifiers. Chinese names are converted to pinyin automatically and can be edited manually.')}
            />
            <div className="flex justify-end">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={regenerateSlug}
              >
                {t('tenants.regenerateSlug', 'Generate from name')}
              </Button>
            </div>
          </div>
          <Button
            type="submit"
            isLoading={createTenant.isPending}
            leftIcon={<Plus className="size-4" />}
          >
            {t('tenants.submit', 'Add and switch')}
          </Button>
        </form>
      </Dialog>

      <section>
        <Card>
          <CardHeader
            title={t('tenants.listTitle', 'My organizations')}
            description={t('tenants.listDescription', 'After selecting an organization, member, invitation, API key, and audit features use that context.')}
          />
          <div className="overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <Th>{t('tenants.columns.organization', 'Organization')}</Th>
                  <Th>{t('tenants.columns.role', 'Role')}</Th>
                  <Th>{t('tenants.columns.status', 'Status')}</Th>
                  <Th>{t('tenants.columns.actions', 'Actions')}</Th>
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
                        {formatRole(tenant.role)}
                      </Badge>
                    </Td>
                    <Td>
                      <Badge
                        tone={
                          tenant.tenant_status === 'active' ? 'green' : 'orange'
                        }
                      >
                        {formatTenantStatus(tenant.tenant_status)}
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
                            ? t('tenants.current', 'Current')
                            : t('tenants.switch', 'Switch')}
                        </Button>
                        <Link
                          to={`/tenant/settings?tenantId=${tenant.tenant_id}`}
                        >
                          <Button
                            size="sm"
                            variant="ghost"
                            leftIcon={<Building2 className="size-3" />}
                          >
                            {t('tenants.settings', 'Settings')}
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
