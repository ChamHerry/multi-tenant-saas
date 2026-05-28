import { FormEvent, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Inbox } from 'lucide-react'
import { useAcceptInvitation, useDeclineInvitation, useMyInvitations } from '@/features/invitations/invitation-hooks'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { EmptyState } from '@/shared/ui/EmptyState'
import { Input } from '@/shared/ui/Input'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function MyInvitationsPage() {
  const [searchParams] = useSearchParams()
  const [token, setToken] = useState(searchParams.get('token') ?? '')
  const invitations = useMyInvitations('pending')
  const acceptInvitation = useAcceptInvitation()
  const declineInvitation = useDeclineInvitation()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    acceptInvitation.mutate(token.trim(), { onSuccess: () => setToken('') })
  }

  const firstError = invitations.error ?? acceptInvitation.error ?? declineInvitation.error
  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title="邀请处理失败" /> : null}
      <Card>
        <CardHeader title="通过 Token 接受邀请" description="如果你拿到的是邀请链接或 token，可在这里提交。" />
        <form className="flex flex-col gap-3 sm:flex-row" onSubmit={submit}>
          <Input className="flex-1" value={token} onChange={(event) => setToken(event.target.value)} placeholder="邀请 token" required />
          <Button type="submit" isLoading={acceptInvitation.isPending} leftIcon={<Inbox className="size-4" />}>接受邀请</Button>
        </form>
      </Card>
      <Card>
        <CardHeader title="待处理邀请" />
        {invitations.isLoading ? <LoadingView label="加载我的邀请..." /> : (invitations.data?.invitations.length ?? 0) === 0 ? (
          <EmptyState title="暂无待处理邀请" description="被邀请邮箱与当前账号邮箱匹配时会显示在这里。" />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <thead><tr><Th>组织</Th><Th>邮箱</Th><Th>角色</Th><Th>过期时间</Th><Th>操作</Th></tr></thead>
              <tbody>
                {invitations.data?.invitations.map((item) => (
                  <tr key={item.id}>
                    <Td><div className="font-bold text-ink">{item.tenant_name || item.tenant_id}</div><div className="text-xs text-subtle">{item.tenant_slug}</div></Td>
                    <Td>{item.invitee_email}</Td>
                    <Td><Badge tone="purple">{item.role}</Badge></Td>
                    <Td>{new Date(item.expires_at).toLocaleString()}</Td>
                    <Td><div className="flex gap-2"><Button size="sm" onClick={() => acceptInvitation.mutate('', { onError: () => undefined })} disabled>需要 Token</Button><Button size="sm" variant="danger" onClick={() => declineInvitation.mutate(item.id)} isLoading={declineInvitation.isPending}>拒绝</Button></div></Td>
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
