import { useState } from 'react'
import { Shield } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAccess, useCanPlatform } from '@/features/access/access-hooks'
import { useAdminTenantActions, useAdminTenants } from '@/features/admin/admin-hooks'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Input } from '@/shared/ui/Input'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function AdminTenantsPage() {
  const { t } = useTranslation()
  const [query, setQuery] = useState('')
  const params = { query }
  const access = useAccess()
  const canManageTenants = useCanPlatform('platform:tenant:manage')
  const tenants = useAdminTenants(params)
  const actions = useAdminTenantActions(params)
  const firstError = access.error ?? tenants.error ?? actions.suspend.error ?? actions.restore.error
  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title={t('admin.tenants.unavailable')} /> : null}
      <Card><CardHeader title={t('admin.common.session')} />{access.isLoading ? <LoadingView label={t('admin.tenants.sessionLoading')} /> : <div className="text-sm text-muted">{t('admin.common.rolePrefix')}<Badge tone="red">{t(`admin.common.roles.${access.data?.platform_admin?.role ?? ''}`, { defaultValue: access.data?.platform_admin?.role })}</Badge></div>}</Card>
      <Card><CardHeader title={t('audit.common.filter')} /><Input label={t('admin.common.search')} value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t('admin.tenants.searchPlaceholder')} /></Card>
      <Card>
        <CardHeader title={t('admin.tenants.listTitle')} description={t('common.total', { count: tenants.data?.total ?? 0 })} />
        {tenants.isLoading ? <LoadingView label={t('admin.tenants.loading')} /> : <div className="overflow-x-auto"><Table><thead><tr><Th>{t('admin.tenants.headers.organization')}</Th><Th>{t('admin.tenants.headers.status')}</Th><Th>{t('admin.tenants.headers.actions')}</Th></tr></thead><tbody>{tenants.data?.items.map((tenant) => <tr key={tenant.id}><Td><div className="font-bold text-ink">{tenant.name}</div><div className="text-xs text-subtle">{tenant.id}</div></Td><Td><Badge tone={tenant.status === 'active' ? 'green' : 'orange'}>{t(`common.status.${tenant.status}`, { defaultValue: tenant.status })}</Badge></Td><Td><div className="flex flex-wrap gap-2"><Button size="sm" variant="secondary" leftIcon={<Shield className="size-3" />} onClick={() => actions.suspend.mutate(tenant.id)} disabled={!canManageTenants || tenant.status !== 'active'}>{t('common.actions.suspend')}</Button><Button size="sm" variant="secondary" onClick={() => actions.restore.mutate(tenant.id)} disabled={!canManageTenants || tenant.status === 'active'}>{t('common.actions.restore')}</Button></div></Td></tr>)}</tbody></Table></div>}
      </Card>
    </div>
  )
}
