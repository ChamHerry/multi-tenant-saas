import { FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { ShieldCheck } from 'lucide-react'
import { useSecurityEvents } from '@/features/audit/audit-hooks'
import { authKeys } from '@/features/auth/auth-hooks'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { disableTOTP, getTOTPStatus, regenerateBackupCodes } from '@/features/auth/totp-api'
import { useDateTimeFormatter } from '@/shared/i18n'
import { Badge, Button, Card, CardHeader, EmptyState, ErrorView, Input, LoadingView, Table, Td, Th, Toast } from '@/shared/ui'
import { errorMessage } from '@/shared/api/errors'

export function SecurityEventsPage() {
  const { t } = useTranslation()
  const { messages } = useAuthI18n()
  const totpCopy = messages.totp
  const dateTimeFormatter = useDateTimeFormatter()
  const queryClient = useQueryClient()
  const logs = useSecurityEvents({ limit: 100 })
  const totpStatus = useQuery({ queryKey: authKeys.totpStatus, queryFn: getTOTPStatus, retry: false })
  const disableMutation = useMutation({ mutationFn: disableTOTP })
  const regenerateMutation = useMutation({ mutationFn: regenerateBackupCodes })
  const [disablePassword, setDisablePassword] = useState('')
  const [disableCode, setDisableCode] = useState('')
  const [regenPassword, setRegenPassword] = useState('')
  const [regeneratedCodes, setRegeneratedCodes] = useState<string[]>([])

  const refreshStatus = async () => {
    await queryClient.invalidateQueries({ queryKey: authKeys.totpStatus })
  }

  const handleDisable = async (event: FormEvent) => {
    event.preventDefault()
    await disableMutation.mutateAsync({ password: disablePassword || undefined, code: disableCode || undefined })
    setDisablePassword('')
    setDisableCode('')
    setRegeneratedCodes([])
    await refreshStatus()
  }

  const handleRegenerate = async (event: FormEvent) => {
    event.preventDefault()
    const result = await regenerateMutation.mutateAsync({ password: regenPassword })
    setRegeneratedCodes(result.backup_codes)
    setRegenPassword('')
    await refreshStatus()
  }

  const enabledAt = totpStatus.data?.enabled_at ? dateTimeFormatter.format(new Date(totpStatus.data.enabled_at)) : t('common.notAvailable')

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader
          title={totpCopy.title}
          description={totpCopy.description}
          action={totpStatus.data?.enabled ? <Badge tone="green">{totpCopy.status.enabledBadge}</Badge> : <Badge>{totpCopy.status.disabledBadge}</Badge>}
        />
        {totpStatus.isLoading ? <LoadingView label={t('common.loading')} /> : null}
        {totpStatus.error ? <ErrorView error={totpStatus.error} title={totpCopy.status.loadFailed} /> : null}
        {totpStatus.data ? (
          <div className="space-y-5">
            {totpStatus.data.enabled ? (
              <div className="space-y-4">
                <div className="rounded-panel border border-success/30 bg-success-soft p-4 text-sm text-success-strong">
                  <div className="flex items-center gap-2 font-bold"><ShieldCheck className="size-4" />{totpCopy.status.enabled}</div>
                  <div className="mt-2 text-xs">
                    {totpCopy.status.enabledAt.replace('{time}', enabledAt)} · {totpCopy.status.backupCodesRemaining.replace('{count}', String(totpStatus.data.backup_codes_remaining))}
                  </div>
                </div>

                <form className="grid gap-3 rounded-panel border border-line bg-surface-soft p-4 md:grid-cols-[1fr_1fr_auto]" onSubmit={handleDisable}>
                  <Input
                    label={totpCopy.status.disablePasswordLabel}
                    type="password"
                    value={disablePassword}
                    onChange={(event) => setDisablePassword(event.target.value)}
                    autoComplete="current-password"
                  />
                  <Input
                    label={totpCopy.status.disableCodeLabel}
                    value={disableCode}
                    onChange={(event) => setDisableCode(event.target.value)}
                    autoComplete="one-time-code"
                  />
                  <div className="flex items-end">
                    <Button className="w-full" type="submit" variant="danger" isLoading={disableMutation.isPending}>
                      {totpCopy.status.disable}
                    </Button>
                  </div>
                  <div className="md:col-span-3">
                    <Toast tone="red" message={disableMutation.isError ? errorMessage(disableMutation.error) : undefined} />
                  </div>
                </form>

                <form className="grid gap-3 rounded-panel border border-line bg-surface-soft p-4 md:grid-cols-[1fr_auto]" onSubmit={handleRegenerate}>
                  <Input
                    label={totpCopy.backupCodes.regeneratePasswordLabel}
                    type="password"
                    value={regenPassword}
                    onChange={(event) => setRegenPassword(event.target.value)}
                    autoComplete="current-password"
                  />
                  <div className="flex items-end">
                    <Button className="w-full" type="submit" variant="secondary" isLoading={regenerateMutation.isPending}>
                      {totpCopy.backupCodes.regenerate}
                    </Button>
                  </div>
                  <div className="md:col-span-2">
                    <Toast tone="red" message={regenerateMutation.isError ? errorMessage(regenerateMutation.error) : undefined} />
                  </div>
                </form>

                {regeneratedCodes.length > 0 ? (
                  <div className="rounded-panel border border-brand-ring bg-brand-soft p-4">
                    <div className="font-bold text-brand">{totpCopy.backupCodes.title}</div>
                    <p className="mt-1 text-sm text-muted">{totpCopy.backupCodes.description}</p>
                    <div className="mt-3 grid gap-2 sm:grid-cols-2">
                      {regeneratedCodes.map((item) => (
                        <code key={item} className="rounded-control bg-white px-3 py-2 text-center font-bold tracking-widest text-ink">
                          {item}
                        </code>
                      ))}
                    </div>
                  </div>
                ) : null}
              </div>
            ) : (
              <div className="rounded-panel border border-line bg-surface-soft p-4">
                <div className="font-bold text-ink">{totpCopy.status.disabled}</div>
                <p className="mt-1 text-sm leading-6 text-muted">{totpCopy.status.disabledDescription}</p>
                <Link className="mt-4 inline-flex" to="/me/security/totp/setup">
                  <Button leftIcon={<ShieldCheck className="size-4" />}>{totpCopy.status.enable}</Button>
                </Link>
              </div>
            )}
          </div>
        ) : null}
      </Card>

      {logs.error ? <ErrorView error={logs.error} title={t('audit.security.loadFailed')} /> : null}
      <Card>
        <CardHeader title={t('audit.security.title')} />
        {logs.isLoading ? <LoadingView label={t('audit.security.loading')} /> : (logs.data?.logs.length ?? 0) === 0 ? <EmptyState title={t('audit.security.empty')} /> : (
          <div className="overflow-x-auto">
            <Table>
              <thead><tr><Th>{t('audit.common.headers.time')}</Th><Th>{t('audit.common.headers.action')}</Th><Th>{t('audit.common.headers.ip')}</Th><Th>{t('audit.common.headers.userAgent')}</Th></tr></thead>
              <tbody>{logs.data?.logs.map((item) => <tr key={item.id}><Td>{dateTimeFormatter.format(new Date(item.created_at))}</Td><Td><Badge>{item.action}</Badge></Td><Td>{item.ip}</Td><Td className="max-w-xl break-all">{item.user_agent}</Td></tr>)}</tbody>
            </Table>
          </div>
        )}
      </Card>
    </div>
  )
}
