import { FormEvent, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { Save, ShieldAlert } from 'lucide-react'
import { useCanTenant } from '@/features/access/access-hooks'
import { useTenantAction, useTenantDetail, useUpdateTenant } from '@/features/tenants/tenant-hooks'
import { useTenantI18n } from '@/features/tenants/tenant-i18n'
import { createUpdateTenantInputSchema } from '@/features/tenants/tenant-types'
import { useTenantStore } from '@/features/tenants/tenant-store'
import type { Tenant } from '@/features/tenants/tenant-types'
import { Button, Card, CardHeader, Dialog, EmptyState, Input, ErrorView, LoadingView } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'
import { pageTitle, usePageTitle } from '@/layouts/page-title'

function TenantEditForm({
  tenantId,
  tenant,
  canManage,
  onSaved,
}: {
  tenantId: string
  tenant: Tenant
  canManage: boolean
  onSaved?: () => void
}) {
  const updateTenant = useUpdateTenant(tenantId)
  const [name, setName] = useState(tenant.name)
  const [slug, setSlug] = useState(tenant.slug)
  const [errors, setErrors] = useState<FieldErrors>({})
  const { schemaTranslator, t } = useTenantI18n()
  const updateTenantSchema = useMemo(() => createUpdateTenantInputSchema(schemaTranslator), [schemaTranslator])

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(updateTenantSchema, {
      name: name.trim(),
      slug: slug.trim(),
    })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    updateTenant.mutate(result.data, { onSuccess: onSaved })
  }

  return (
    <>
      {updateTenant.error ? <ErrorView error={updateTenant.error} title={t('tenantDetail.updateErrorTitle', 'Organization update failed')} /> : null}
      <form className="grid gap-4 sm:grid-cols-2" onSubmit={submit}>
        <Input label={t('tenantDetail.nameLabel', 'Name')} value={name} onChange={(event) => setName(event.target.value)} disabled={!canManage} error={canManage ? errors.name : undefined} />
        <Input label={t('tenantDetail.slugLabel', 'Slug')} value={slug} onChange={(event) => setSlug(event.target.value)} disabled={!canManage} error={canManage ? errors.slug : undefined} />
        <div className="sm:col-span-2">
          <Button type="submit" disabled={!canManage} isLoading={updateTenant.isPending} leftIcon={<Save className="size-4" />}>{t('tenantDetail.save', 'Save changes')}</Button>
        </div>
      </form>
    </>
  )
}

export function TenantDetailPage() {
  const params = useParams()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const setCurrentTenantId = useTenantStore((state) => state.setCurrentTenantId)
  const tenantId = params.tenantId ?? searchParams.get('tenantId') ?? currentTenantId ?? ''
  const detail = useTenantDetail(tenantId)
  const actions = useTenantAction(tenantId)
  const { t } = useTenantI18n()
  const canManageTenant = useCanTenant('tenant:manage', tenantId)
  const tenant = detail.data?.tenant
  const tenantName = tenant?.name
  const [isEditOpen, setIsEditOpen] = useState(false)
  const tenantPageTitle = useMemo(
    () =>
      tenantName
        ? pageTitle('routes.tenantSettings.dynamic', { name: tenantName })
        : pageTitle('routes.tenantSettings.title'),
    [tenantName],
  )

  usePageTitle(tenantPageTitle)

  useEffect(() => {
    if (tenantId) setCurrentTenantId(tenantId)
  }, [setCurrentTenantId, tenantId])

  const firstError = useMemo(
    () => detail.error ?? actions.suspend.error ?? actions.restore.error ?? actions.remove.error,
    [actions.remove.error, actions.restore.error, actions.suspend.error, detail.error],
  )

  if (!tenantId) {
    return (
      <EmptyState
        title={t('tenantDetail.noTenantTitle', 'Select an organization first')}
        description={t('tenantDetail.noTenantDescription', 'Organization settings require the current organization context. Go to My organizations first to select one.')}
        action={<Link to="/tenants"><Button variant="secondary">{t('tenantDetail.goToOrganizations', 'Go to my organizations')}</Button></Link>}
      />
    )
  }
  if (detail.isLoading) return <LoadingView label={t('tenantDetail.loading', 'Loading organization details...')} />

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-end gap-2">
        <Button disabled={!tenant || !canManageTenant} onClick={() => setIsEditOpen(true)}>{t('tenantDetail.editBasics', 'Edit basics')}</Button>
        <Button variant="secondary" onClick={() => navigate('/tenants')}>{t('tenantDetail.backToOrganizations', 'Back to my organizations')}</Button>
      </div>

      {firstError ? <ErrorView error={firstError} title={t('tenantDetail.operationErrorTitle', 'Organization action failed')} /> : null}

      <Dialog open={isEditOpen} title={t('tenantDetail.editDialogTitle', 'Edit basics')} onClose={() => setIsEditOpen(false)}>
        {tenant ? <TenantEditForm key={tenant.id} tenantId={tenantId} tenant={tenant} canManage={canManageTenant} onSaved={() => setIsEditOpen(false)} /> : <ErrorView title={t('tenantDetail.notFoundTitle', 'Organization not found')} error={t('tenantDetail.notFoundDescription', 'The backend did not return organization data.')} />}
      </Dialog>

      <section>
        <Card>
          <CardHeader title={t('tenantDetail.capabilityTitle', 'Template capability boundary')} description={t('tenantDetail.capabilityDescription', 'This project keeps only generic multi-tenant SaaS administration capabilities and does not include domain-specific business models.')} />
          <div className="space-y-3 text-sm text-muted">
            <div className="rounded-panel bg-surface-soft p-3"><span className="font-bold text-ink">{t('tenantDetail.accessBoundaryTitle', 'Access boundary')}</span><p>{t('tenantDetail.accessBoundaryDescription', 'Member roles, organization API key scopes, and platform administrator permissions jointly control console access.')}</p></div>
            <div className="rounded-panel bg-surface-soft p-3"><span className="font-bold text-ink">{t('tenantDetail.extensionPointTitle', 'Extension point')}</span><p>{t('tenantDetail.extensionPointDescription', 'Future business tables should explicitly bind tenant_id and reuse the existing RBAC and audit services.')}</p></div>
          </div>
        </Card>
      </section>

      <Card>
        <CardHeader title={t('tenantDetail.dangerTitle', 'Danger zone')} description={t('tenantDetail.dangerDescription', 'These actions require tenant:manage permission and write audit logs.')} />
        <div className="flex flex-wrap gap-3">
          <Button variant="secondary" disabled={!canManageTenant} onClick={() => actions.suspend.mutate()} isLoading={actions.suspend.isPending} leftIcon={<ShieldAlert className="size-4" />}>{t('tenantDetail.suspend', 'Suspend organization')}</Button>
          <Button variant="secondary" disabled={!canManageTenant} onClick={() => actions.restore.mutate()} isLoading={actions.restore.isPending}>{t('tenantDetail.restore', 'Restore organization')}</Button>
          <Button variant="danger" disabled={!canManageTenant} onClick={() => actions.remove.mutate()} isLoading={actions.remove.isPending}>{t('tenantDetail.delete', 'Soft delete organization')}</Button>
        </div>
      </Card>
    </div>
  )
}
