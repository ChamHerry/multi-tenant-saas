import { FormEvent, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Inbox } from 'lucide-react'
import { useAcceptInvitation, useDeclineInvitation, useMyInvitations } from '@/features/invitations/invitation-hooks'
import { createAcceptInvitationTokenSchema } from '@/features/invitations/invitation-types'
import { useTenantI18n } from '@/features/tenants/tenant-i18n'
import { Badge, Button, Card, CardHeader, Dialog, EmptyState, Input, ErrorView, LoadingView, Table, Td, Th } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'

export function MyInvitationsPage() {
  const [searchParams] = useSearchParams()
  const [token, setToken] = useState(searchParams.get('token') ?? '')
  const [isAcceptTokenOpen, setIsAcceptTokenOpen] = useState(Boolean(searchParams.get('token')))
  const [errors, setErrors] = useState<FieldErrors>({})
  const invitations = useMyInvitations('pending')
  const acceptInvitation = useAcceptInvitation()
  const declineInvitation = useDeclineInvitation()
  const { schemaTranslator, t, formatRole, formatDateTime } = useTenantI18n()
  const acceptInvitationSchema = useMemo(() => createAcceptInvitationTokenSchema(schemaTranslator), [schemaTranslator])

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(acceptInvitationSchema, { token: token.trim() })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    acceptInvitation.mutate(result.data.token, {
      onSuccess: () => {
        setToken('')
        setIsAcceptTokenOpen(false)
      },
    })
  }

  const firstError = invitations.error ?? acceptInvitation.error ?? declineInvitation.error
  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title={t('myInvitations.errorTitle', 'Invitation handling failed')} /> : null}
      <div className="flex justify-end">
        <Button leftIcon={<Inbox className="size-4" />} onClick={() => setIsAcceptTokenOpen(true)}>{t('myInvitations.acceptByToken', 'Accept by token')}</Button>
      </div>

      <Dialog open={isAcceptTokenOpen} title={t('myInvitations.dialogTitle', 'Accept by token')} onClose={() => setIsAcceptTokenOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input value={token} onChange={(event) => setToken(event.target.value)} placeholder={t('myInvitations.tokenPlaceholder', 'Invitation token')} error={errors.token} />
          <Button type="submit" isLoading={acceptInvitation.isPending} leftIcon={<Inbox className="size-4" />}>{t('myInvitations.submit', 'Accept invitation')}</Button>
        </form>
      </Dialog>

      <Card>
        <CardHeader title={t('myInvitations.pendingTitle', 'Pending invitations')} />
        {invitations.isLoading ? <LoadingView label={t('myInvitations.loading', 'Loading my invitations...')} /> : (invitations.data?.invitations.length ?? 0) === 0 ? (
          <EmptyState title={t('myInvitations.emptyTitle', 'No pending invitations')} description={t('myInvitations.emptyDescription', 'Invitations matching your account email will appear here.')} />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <thead><tr><Th>{t('myInvitations.columns.organization', 'Organization')}</Th><Th>{t('myInvitations.columns.email', 'Email')}</Th><Th>{t('myInvitations.columns.role', 'Role')}</Th><Th>{t('myInvitations.columns.expiresAt', 'Expires at')}</Th><Th>{t('myInvitations.columns.actions', 'Actions')}</Th></tr></thead>
              <tbody>
                {invitations.data?.invitations.map((item) => (
                  <tr key={item.id}>
                    <Td><div className="font-bold text-ink">{item.tenant_name || item.tenant_id}</div><div className="text-xs text-subtle">{item.tenant_slug}</div></Td>
                    <Td>{item.invitee_email}</Td>
                    <Td><Badge tone="purple">{formatRole(item.role)}</Badge></Td>
                    <Td>{formatDateTime(item.expires_at)}</Td>
                    <Td><div className="flex gap-2"><Button size="sm" onClick={() => acceptInvitation.mutate('', { onError: () => undefined })} disabled>{t('myInvitations.tokenRequired', 'Token required')}</Button><Button size="sm" variant="danger" onClick={() => declineInvitation.mutate(item.id)} isLoading={declineInvitation.isPending}>{t('myInvitations.decline', 'Decline')}</Button></div></Td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        )}
      </Card>
    </div>
  )
}
