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

  if (!tenantId) return <EmptyState title={t('tenantInvitations.noTenantTitle', '请先选择组织')} description={t('tenantInvitations.noTenantDescription', '邀请成员需要当前组织上下文。')} />
  const firstError = invitations.error ?? createInvitation.error ?? resendInvitation.error ?? revokeInvitation.error

  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title={t('tenantInvitations.errorTitle', '邀请操作失败')} /> : null}
      {latestToken ? (
        <Card className="border-success/30 bg-success-soft">
          <CardHeader title={t('tenantInvitations.tokenTitle', '邀请链接/Token')} description={t('tenantInvitations.tokenDescription', '邮件发送未接入时可手动复制给被邀请用户。')} />
          <div className="flex flex-col gap-3 sm:flex-row">
            <code className="min-w-0 flex-1 overflow-x-auto rounded-panel bg-white p-3 text-sm font-bold text-success-strong">{latestToken}</code>
            <Button variant="secondary" leftIcon={<Copy className="size-4" />} onClick={() => void navigator.clipboard?.writeText(latestToken)}>{t('tenantInvitations.copy', '复制')}</Button>
          </div>
        </Card>
      ) : null}
      <div className="flex justify-end">
        <Button leftIcon={<MailPlus className="size-4" />} onClick={() => setIsInviteOpen(true)}>{t('tenantInvitations.inviteButton', '邀请成员')}</Button>
      </div>

      <Dialog open={isInviteOpen} title={t('tenantInvitations.dialogTitle', '邀请成员')} onClose={() => setIsInviteOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input label={t('tenantInvitations.emailLabel', '邮箱')} type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder={t('tenantInvitations.emailPlaceholder', 'member@example.com')} error={errors.invitee_email} />
          <Select label={t('tenantInvitations.roleLabel', '角色')} value={role} onChange={(event) => setRole(event.target.value as TenantRole)}>
            {roles.map((item) => <option key={item} value={item}>{formatRole(item)}</option>)}
          </Select>
          <Input label={t('tenantInvitations.messageLabel', '留言，可选')} value={message} onChange={(event) => setMessage(event.target.value)} placeholder={t('tenantInvitations.messagePlaceholder', '欢迎加入我们的组织')} error={errors.message} />
          <Button type="submit" isLoading={createInvitation.isPending} leftIcon={<MailPlus className="size-4" />}>{t('tenantInvitations.submit', '发送邀请')}</Button>
        </form>
      </Dialog>

      <section>
        <Card>
          <CardHeader title={t('tenantInvitations.listTitle', '邀请列表')} description={t('tenantInvitations.listDescription', '待处理邀请可重发或撤销；已接受、已过期、已撤销的邀请会保留审计线索。')} />
          {invitations.isLoading ? <LoadingView label={t('tenantInvitations.loading', '加载邀请...')} /> : (invitations.data?.invitations.length ?? 0) === 0 ? (
            <EmptyState title={t('tenantInvitations.emptyTitle', '暂无邀请')} description={t('tenantInvitations.emptyDescription', '发送第一封邀请后会显示在这里。')} />
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <thead><tr><Th>{t('tenantInvitations.columns.email', '邮箱')}</Th><Th>{t('tenantInvitations.columns.role', '角色')}</Th><Th>{t('tenantInvitations.columns.status', '状态')}</Th><Th>{t('tenantInvitations.columns.expiresAt', '过期时间')}</Th><Th>{t('tenantInvitations.columns.actions', '操作')}</Th></tr></thead>
                <tbody>
                  {invitations.data?.invitations.map((item) => (
                    <tr key={item.id}>
                      <Td><div className="font-bold text-ink">{item.invitee_email}</div><div className="mt-1 text-xs text-subtle">{item.id}</div></Td>
                      <Td><Badge tone="purple">{formatRole(item.role)}</Badge></Td>
                      <Td><Badge tone={item.status === 'pending' ? 'orange' : item.status === 'accepted' ? 'green' : 'gray'}>{formatInvitationStatus(item.status)}</Badge></Td>
                      <Td>{formatDateTime(item.expires_at)}</Td>
                      <Td>
                        <div className="flex gap-2">
                          <Button size="sm" variant="secondary" disabled={item.status !== 'pending'} isLoading={resendInvitation.isPending} onClick={() => resendInvitation.mutate(item.id, { onSuccess: (data) => setLatestToken(data.accept_url || data.token || '') })}>{t('tenantInvitations.resend', '重发')}</Button>
                          <Button size="sm" variant="danger" disabled={item.status !== 'pending'} isLoading={revokeInvitation.isPending} onClick={() => revokeInvitation.mutate(item.id)}>{t('tenantInvitations.revoke', '撤销')}</Button>
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
