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
      {firstError ? <ErrorView error={firstError} title={t('myInvitations.errorTitle', '邀请处理失败')} /> : null}
      <div className="flex justify-end">
        <Button leftIcon={<Inbox className="size-4" />} onClick={() => setIsAcceptTokenOpen(true)}>{t('myInvitations.acceptByToken', '通过 Token 接受邀请')}</Button>
      </div>

      <Dialog open={isAcceptTokenOpen} title={t('myInvitations.dialogTitle', '通过 Token 接受邀请')} onClose={() => setIsAcceptTokenOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input value={token} onChange={(event) => setToken(event.target.value)} placeholder={t('myInvitations.tokenPlaceholder', '邀请 token')} error={errors.token} />
          <Button type="submit" isLoading={acceptInvitation.isPending} leftIcon={<Inbox className="size-4" />}>{t('myInvitations.submit', '接受邀请')}</Button>
        </form>
      </Dialog>

      <Card>
        <CardHeader title={t('myInvitations.pendingTitle', '待处理邀请')} />
        {invitations.isLoading ? <LoadingView label={t('myInvitations.loading', '加载我的邀请...')} /> : (invitations.data?.invitations.length ?? 0) === 0 ? (
          <EmptyState title={t('myInvitations.emptyTitle', '暂无待处理邀请')} description={t('myInvitations.emptyDescription', '被邀请邮箱与当前账号邮箱匹配时会显示在这里。')} />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <thead><tr><Th>{t('myInvitations.columns.organization', '组织')}</Th><Th>{t('myInvitations.columns.email', '邮箱')}</Th><Th>{t('myInvitations.columns.role', '角色')}</Th><Th>{t('myInvitations.columns.expiresAt', '过期时间')}</Th><Th>{t('myInvitations.columns.actions', '操作')}</Th></tr></thead>
              <tbody>
                {invitations.data?.invitations.map((item) => (
                  <tr key={item.id}>
                    <Td><div className="font-bold text-ink">{item.tenant_name || item.tenant_id}</div><div className="text-xs text-subtle">{item.tenant_slug}</div></Td>
                    <Td>{item.invitee_email}</Td>
                    <Td><Badge tone="purple">{formatRole(item.role)}</Badge></Td>
                    <Td>{formatDateTime(item.expires_at)}</Td>
                    <Td><div className="flex gap-2"><Button size="sm" onClick={() => acceptInvitation.mutate('', { onError: () => undefined })} disabled>{t('myInvitations.tokenRequired', '需要 Token')}</Button><Button size="sm" variant="danger" onClick={() => declineInvitation.mutate(item.id)} isLoading={declineInvitation.isPending}>{t('myInvitations.decline', '拒绝')}</Button></div></Td>
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
