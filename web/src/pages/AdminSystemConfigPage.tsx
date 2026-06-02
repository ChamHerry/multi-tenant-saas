import { FormEvent, useMemo, useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useCanPlatform } from '@/features/access/access-hooks'
import { useAdminSystemConfigMutations, useAdminSystemConfigs } from '@/features/admin/admin-hooks'
import type { SystemConfigItem, SystemConfigValueType } from '@/features/admin/admin-types'
import { useDateTimeFormatter } from '@/shared/i18n'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Dialog } from '@/shared/ui/Dialog'
import { EmptyState } from '@/shared/ui/EmptyState'
import { Input } from '@/shared/ui/Input'
import { Select } from '@/shared/ui/Select'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

const valueTypes: SystemConfigValueType[] = ['string', 'number', 'bool', 'json', 'secret']

function emptyForm() {
  return {
    key: '',
    value: '',
    valueType: 'string' as SystemConfigValueType,
    description: '',
    isExisting: false,
    isSecretWithValue: false,
  }
}

export function AdminSystemConfigPage() {
  const { t } = useTranslation()
  const dateTimeFormatter = useDateTimeFormatter()
  const [query, setQuery] = useState('')
  const [category, setCategory] = useState('')
  const params = { query, category, limit: 100 }
  const configs = useAdminSystemConfigs(params)
  const mutations = useAdminSystemConfigMutations(params)
  const canManageConfig = useCanPlatform('platform:config:manage')
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [form, setForm] = useState(emptyForm())
  const firstError = configs.error ?? mutations.upsert.error ?? mutations.deleteConfig.error

  const categories = useMemo(() => {
    const values = new Set(configs.data?.items.map((item) => item.category).filter(Boolean))
    return Array.from(values).sort()
  }, [configs.data?.items])

  const openCreate = () => {
    setForm(emptyForm())
    setIsDialogOpen(true)
  }

  const openEdit = (item: SystemConfigItem) => {
    setForm({
      key: item.key,
      value: item.is_secret ? '' : item.value,
      valueType: item.value_type,
      description: item.description,
      isExisting: true,
      isSecretWithValue: item.is_secret && item.has_value,
    })
    setIsDialogOpen(true)
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const trimmedKey = form.key.trim()
    const valueProvided = !(form.valueType === 'secret' && form.isExisting && form.value.trim() === '')
    mutations.upsert.mutate(
      {
        key: trimmedKey,
        payload: {
          value: form.value,
          value_provided: valueProvided,
          value_type: form.valueType,
          description: form.description,
        },
      },
      {
        onSuccess: () => {
          setIsDialogOpen(false)
          setForm(emptyForm())
        },
      },
    )
  }

  const deleteConfig = (item: SystemConfigItem) => {
    const confirmed = window.confirm(t('admin.systemConfig.deleteConfirm', { key: item.key }))
    if (confirmed) mutations.deleteConfig.mutate(item.key)
  }

  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title={t('admin.systemConfig.failed')} /> : null}

      <div className="flex justify-end">
        <Button disabled={!canManageConfig} leftIcon={<Plus className="size-4" />} onClick={openCreate}>
          {t('admin.systemConfig.create')}
        </Button>
      </div>

      <Card>
        <CardHeader title={t('admin.systemConfig.filtersTitle')} description={t('admin.systemConfig.filtersDescription')} />
        <div className="grid gap-4 md:grid-cols-2">
          <Input
            label={t('admin.common.search')}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={t('admin.systemConfig.searchPlaceholder')}
          />
          <Select label={t('admin.systemConfig.category')} value={category} onChange={(event) => setCategory(event.target.value)}>
            <option value="">{t('admin.systemConfig.allCategories')}</option>
            {categories.map((item) => (
              <option key={item} value={item}>
                {item}
              </option>
            ))}
          </Select>
        </div>
      </Card>

      <Card>
        <CardHeader title={t('admin.systemConfig.listTitle')} description={t('common.total', { count: configs.data?.total ?? 0 })} />
        {configs.isLoading ? (
          <LoadingView label={t('admin.systemConfig.loading')} />
        ) : !configs.data?.items.length ? (
          <EmptyState title={t('admin.systemConfig.emptyTitle')} description={t('admin.systemConfig.emptyDescription')} />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <Th>{t('admin.systemConfig.headers.key')}</Th>
                  <Th>{t('admin.systemConfig.headers.value')}</Th>
                  <Th>{t('admin.systemConfig.headers.type')}</Th>
                  <Th>{t('admin.systemConfig.headers.description')}</Th>
                  <Th>{t('admin.systemConfig.headers.updatedAt')}</Th>
                  <Th>{t('common.fields.actions')}</Th>
                </tr>
              </thead>
              <tbody>
                {configs.data.items.map((item) => (
                  <tr key={item.key}>
                    <Td>
                      <div className="font-bold text-ink">{item.key}</div>
                      <div className="text-xs text-subtle">{item.category}</div>
                    </Td>
                    <Td>
                      {item.is_secret ? (
                        <div className="space-y-1">
                          <Badge tone={item.has_value ? 'orange' : 'gray'}>{item.has_value ? item.masked_value : t('common.notAvailable')}</Badge>
                          <div className="text-xs text-subtle">{t('admin.systemConfig.secretMasked')}</div>
                        </div>
                      ) : (
                        <code className="break-all text-xs text-ink">{item.value || t('common.notAvailable')}</code>
                      )}
                    </Td>
                    <Td>
                      <Badge tone={item.is_secret ? 'orange' : 'blue'}>{t(`admin.systemConfig.valueTypes.${item.value_type}`)}</Badge>
                    </Td>
                    <Td>{item.description || t('common.notAvailable')}</Td>
                    <Td>{dateTimeFormatter.format(new Date(item.updated_at))}</Td>
                    <Td>
                      <div className="flex flex-wrap gap-2">
                        <Button size="sm" variant="secondary" disabled={!canManageConfig} onClick={() => openEdit(item)}>
                          {t('admin.systemConfig.edit')}
                        </Button>
                        <Button
                          size="sm"
                          variant="danger"
                          disabled={!canManageConfig || mutations.deleteConfig.isPending}
                          leftIcon={<Trash2 className="size-3" />}
                          onClick={() => deleteConfig(item)}
                        >
                          {t('admin.systemConfig.delete')}
                        </Button>
                      </div>
                    </Td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        )}
      </Card>

      <Dialog open={isDialogOpen} title={form.isExisting ? t('admin.systemConfig.editTitle') : t('admin.systemConfig.createTitle')} onClose={() => setIsDialogOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input
            label={t('admin.systemConfig.key')}
            value={form.key}
            onChange={(event) => setForm((current) => ({ ...current, key: event.target.value }))}
            placeholder={t('admin.systemConfig.keyPlaceholder')}
            disabled={form.isExisting}
            required
          />
          <Select
            label={t('admin.systemConfig.type')}
            value={form.valueType}
            onChange={(event) => setForm((current) => ({ ...current, valueType: event.target.value as SystemConfigValueType, value: '' }))}
          >
            {valueTypes.map((item) => (
              <option key={item} value={item}>
                {t(`admin.systemConfig.valueTypes.${item}`)}
              </option>
            ))}
          </Select>
          <label className="block">
            <span className="mb-1.5 block text-xs font-bold uppercase tracking-wide text-muted">{t('admin.systemConfig.value')}</span>
            <textarea
              className="min-h-28 w-full rounded-control border border-line bg-surface-soft px-3 py-2 text-sm text-ink outline-none transition placeholder:text-subtle focus:border-brand focus:bg-white focus:ring-3 focus:ring-brand/10"
              value={form.value}
              onChange={(event) => setForm((current) => ({ ...current, value: event.target.value }))}
              placeholder={form.valueType === 'secret' && form.isSecretWithValue ? t('admin.systemConfig.secretPlaceholder') : t('admin.systemConfig.valuePlaceholder')}
              required={!(form.valueType === 'secret' && form.isExisting)}
            />
            {form.valueType === 'secret' && form.isExisting ? (
              <span className="mt-1 block text-xs text-subtle">{t('admin.systemConfig.secretHint')}</span>
            ) : null}
          </label>
          <Input
            label={t('admin.systemConfig.description')}
            value={form.description}
            onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))}
            placeholder={t('admin.systemConfig.descriptionPlaceholder')}
          />
          <Button type="submit" isLoading={mutations.upsert.isPending} disabled={!canManageConfig}>
            {t('admin.systemConfig.save')}
          </Button>
        </form>
      </Dialog>
    </div>
  )
}
