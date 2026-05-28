import { FormEvent, useState } from 'react'
import { Copy, MailPlus } from 'lucide-react'
import { useCreateInvitation, useResendInvitation, useRevokeInvitation, useTenantInvitations } from '@/features/invitations/invitation-hooks'
import type { TenantRole } from '@/features/tenants/tenant-types'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { EmptyState } from '@/shared/ui/EmptyState'
import { Input } from '@/shared/ui/Input'
import { Select } from '@/shared/ui/Select'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

const roles: TenantRole[] = ['admin', 'member', 'viewer', 'owner']

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

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createInvitation.mutate(
      { invitee_email: email.trim(), role, message: message.trim() || undefined },
      {
        onSuccess: (data) => {
          setLatestToken(data.accept_url || data.token || '')
          setEmail('')
          setMessage('')
          setRole('viewer')
        },
      },
    )
  }

  if (!tenantId) return <EmptyState title="请先选择组织" description="邀请成员需要当前组织上下文。" />
  const firstError = invitations.error ?? createInvitation.error ?? resendInvitation.error ?? revokeInvitation.error

  return (
    <div className="space-y-6">
      <div>
        <Badge tone="purple">Invitations</Badge>
        <h1 className="mt-3 text-3xl font-black text-ink">组织邀请</h1>
        <p className="mt-2 text-sm leading-6 text-muted">通过邮箱发起 pending 邀请，用户登录后可在“我的邀请”中接受或拒绝。</p>
      </div>
      {firstError ? <ErrorView error={firstError} title="邀请操作失败" /> : null}
      {latestToken ? (
        <Card className="border-success/30 bg-success-soft">
          <CardHeader title="邀请链接/Token" description="邮件发送未接入时可手动复制给被邀请用户。" />
          <div className="flex flex-col gap-3 sm:flex-row">
            <code className="min-w-0 flex-1 overflow-x-auto rounded-panel bg-white p-3 text-sm font-bold text-success-strong">{latestToken}</code>
            <Button variant="secondary" leftIcon={<Copy className="size-4" />} onClick={() => void navigator.clipboard?.writeText(latestToken)}>复制</Button>
          </div>
        </Card>
      ) : null}
      <section className="grid gap-5 lg:grid-cols-[0.8fr_1.2fr]">
        <Card>
          <CardHeader title="邀请成员" description="owner/admin 可邀请邮箱加入当前组织。" />
          <form className="space-y-4" onSubmit={submit}>
            <Input label="邮箱" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required placeholder="member@example.com" />
            <Select label="角色" value={role} onChange={(event) => setRole(event.target.value as TenantRole)}>
              {roles.map((item) => <option key={item} value={item}>{item}</option>)}
            </Select>
            <Input label="留言，可选" value={message} onChange={(event) => setMessage(event.target.value)} placeholder="欢迎加入 RepoMind" />
            <Button type="submit" isLoading={createInvitation.isPending} leftIcon={<MailPlus className="size-4" />}>发送邀请</Button>
          </form>
        </Card>
        <Card>
          <CardHeader title="邀请列表" description="pending 可重发或撤销；accepted/expired/revoked 保留审计线索。" />
          {invitations.isLoading ? <LoadingView label="加载邀请..." /> : (invitations.data?.invitations.length ?? 0) === 0 ? (
            <EmptyState title="暂无邀请" description="发送第一封邀请后会显示在这里。" />
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <thead><tr><Th>邮箱</Th><Th>角色</Th><Th>状态</Th><Th>过期时间</Th><Th>操作</Th></tr></thead>
                <tbody>
                  {invitations.data?.invitations.map((item) => (
                    <tr key={item.id}>
                      <Td><div className="font-bold text-ink">{item.invitee_email}</div><div className="mt-1 text-xs text-subtle">{item.id}</div></Td>
                      <Td><Badge tone="purple">{item.role}</Badge></Td>
                      <Td><Badge tone={item.status === 'pending' ? 'orange' : item.status === 'accepted' ? 'green' : 'gray'}>{item.status}</Badge></Td>
                      <Td>{new Date(item.expires_at).toLocaleString()}</Td>
                      <Td>
                        <div className="flex gap-2">
                          <Button size="sm" variant="secondary" disabled={item.status !== 'pending'} isLoading={resendInvitation.isPending} onClick={() => resendInvitation.mutate(item.id, { onSuccess: (data) => setLatestToken(data.accept_url || data.token || '') })}>重发</Button>
                          <Button size="sm" variant="danger" disabled={item.status !== 'pending'} isLoading={revokeInvitation.isPending} onClick={() => revokeInvitation.mutate(item.id)}>撤销</Button>
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
