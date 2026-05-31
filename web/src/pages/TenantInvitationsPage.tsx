import { FormEvent, useMemo, useState } from 'react'
import { Copy, MailPlus } from 'lucide-react'
import { useCreateInvitation, useResendInvitation, useRevokeInvitation, useTenantInvitations } from '@/features/invitations/invitation-hooks'
import { createCreateInvitationSchema } from '@/features/invitations/invitation-types'
import { useTenantI18n } from '@/features/tenants/tenant-i18n'
import { tenantRoleSchema, type TenantRole } from '@/features/tenants/tenant-types'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge, Button, Card, CardHeader, Dialog, EmptyState, Input, Select, ErrorView, LoadingView, Table, Td, Th } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'

const roles = tenantRoleSchema.options

export function TenantInvitationsPage() {
  const tenantId = useTenantStore((state) => state.currentTenantId)
  const invitations = useTenantInvitations(tenantId, '')
  const createInvitation = useCreateInvitation(tenantId)
  const resendInvitation = useResendInvitation(tenantId)
  const revokeInvitation = useRevokeInvitation(tenantId)
  const [email, setEmail] = useState('')
  const [role, setRole] = useState<TenantRole>('viewer')
  const [message, setMessage] = useState('')
  const [latestToken, setLatestToken] = useState('')
  const [isInviteOpen, setIsInviteOpen] = useState(false)
  const [errors, setErrors] = useState<FieldErrors>({})
  const { schemaTranslator, t, formatRole, formatInvitationStatus, formatDateTime } = useTenantI18n()
  const createInvitationInputSchema = useMemo(() => createCreateInvitationSchema(schemaTranslator), [schemaTranslator])

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(createInvitationInputSchema, {
      invitee_email: email.trim(),
      role,
      message: message.trim() || undefined,
    })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    createInvitation.mutate(result.data, {
      onSuccess: (data) => {
        setLatestToken(data.accept_url || data.token || '')
        setEmail('')
        setMessage('')
        setRole('viewer')
        setIsInviteOpen(false)
      },
    })
  }

  if (!tenantId) return <EmptyState title={t('tenantInvitations.noTenantTitle', 'Select an organization first')} description={t('tenantInvitations.noTenantDescription', 'Inviting members requires the current organization context.')} />
  const firstError = invitations.error ?? createInvitation.error ?? resendInvitation.error ?? revokeInvitation.error

  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title={t('tenantInvitations.errorTitle', 'Invitation action failed')} /> : null}
      {latestToken ? (
        <Card className="border-success/30 bg-success-soft">
          <CardHeader title={t('tenantInvitations.tokenTitle', 'Invitation link or token')} description={t('tenantInvitations.tokenDescription', 'Copy this manually for the invited user when email delivery is not connected.')} />
          <div className="flex flex-col gap-3 sm:flex-row">
            <code className="min-w-0 flex-1 overflow-x-auto rounded-panel bg-white p-3 text-sm font-bold text-success-strong">{latestToken}</code>
            <Button variant="secondary" leftIcon={<Copy className="size-4" />} onClick={() => void navigator.clipboard?.writeText(latestToken)}>{t('tenantInvitations.copy', 'Copy')}</Button>
          </div>
        </Card>
      ) : null}
      <div className="flex justify-end">
        <Button leftIcon={<MailPlus className="size-4" />} onClick={() => setIsInviteOpen(true)}>{t('tenantInvitations.inviteButton', 'Invite member')}</Button>
      </div>

      <Dialog open={isInviteOpen} title={t('tenantInvitations.dialogTitle', 'Invite member')} onClose={() => setIsInviteOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input label={t('tenantInvitations.emailLabel', 'Email')} type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder={t('tenantInvitations.emailPlaceholder', 'member@example.com')} error={errors.invitee_email} />
          <Select label={t('tenantInvitations.roleLabel', 'Role')} value={role} onChange={(event) => setRole(event.target.value as TenantRole)}>
            {roles.map((item) => <option key={item} value={item}>{formatRole(item)}</option>)}
          </Select>
          <Input label={t('tenantInvitations.messageLabel', 'Message, optional')} value={message} onChange={(event) => setMessage(event.target.value)} placeholder={t('tenantInvitations.messagePlaceholder', 'Welcome to our organization')} error={errors.message} />
          <Button type="submit" isLoading={createInvitation.isPending} leftIcon={<MailPlus className="size-4" />}>{t('tenantInvitations.submit', 'Send invitation')}</Button>
        </form>
      </Dialog>

      <section>
        <Card>
          <CardHeader title={t('tenantInvitations.listTitle', 'Invitations')} description={t('tenantInvitations.listDescription', 'Pending invitations can be resent or revoked. Accepted, expired, and revoked invitations are retained for audit context.')} />
          {invitations.isLoading ? <LoadingView label={t('tenantInvitations.loading', 'Loading invitations...')} /> : (invitations.data?.invitations.length ?? 0) === 0 ? (
            <EmptyState title={t('tenantInvitations.emptyTitle', 'No invitations yet')} description={t('tenantInvitations.emptyDescription', 'Your first sent invitation will appear here.')} />
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <thead><tr><Th>{t('tenantInvitations.columns.email', 'Email')}</Th><Th>{t('tenantInvitations.columns.role', 'Role')}</Th><Th>{t('tenantInvitations.columns.status', 'Status')}</Th><Th>{t('tenantInvitations.columns.expiresAt', 'Expires at')}</Th><Th>{t('tenantInvitations.columns.actions', 'Actions')}</Th></tr></thead>
                <tbody>
                  {invitations.data?.invitations.map((item) => (
                    <tr key={item.id}>
                      <Td><div className="font-bold text-ink">{item.invitee_email}</div><div className="mt-1 text-xs text-subtle">{item.id}</div></Td>
                      <Td><Badge tone="purple">{formatRole(item.role)}</Badge></Td>
                      <Td><Badge tone={item.status === 'pending' ? 'orange' : item.status === 'accepted' ? 'green' : 'gray'}>{formatInvitationStatus(item.status)}</Badge></Td>
                      <Td>{formatDateTime(item.expires_at)}</Td>
                      <Td>
                        <div className="flex gap-2">
                          <Button size="sm" variant="secondary" disabled={item.status !== 'pending'} isLoading={resendInvitation.isPending} onClick={() => resendInvitation.mutate(item.id, { onSuccess: (data) => setLatestToken(data.accept_url || data.token || '') })}>{t('tenantInvitations.resend', 'Resend')}</Button>
                          <Button size="sm" variant="danger" disabled={item.status !== 'pending'} isLoading={revokeInvitation.isPending} onClick={() => revokeInvitation.mutate(item.id)}>{t('tenantInvitations.revoke', 'Revoke')}</Button>
                        </div>
                      </Td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </div>
          )}
        </Card>
      </section>
    </div>
  )
}
