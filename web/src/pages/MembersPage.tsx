import { FormEvent, useState } from 'react'
import { UserPlus } from 'lucide-react'
import { useCanTenant } from '@/features/access/access-hooks'
import { useAddMember, useMembers, useRemoveMember, useUpdateMember } from '@/features/members/member-hooks'
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

const roles: TenantRole[] = ['owner', 'admin', 'member', 'viewer']
const statuses = ['active', 'invited', 'suspended', 'removed']

export function MembersPage() {
  const tenantId = useTenantStore((state) => state.currentTenantId)
  const members = useMembers(tenantId)
  const addMember = useAddMember(tenantId)
  const updateMember = useUpdateMember(tenantId)
  const removeMember = useRemoveMember(tenantId)
  const canManageMembers = useCanTenant('member:manage', tenantId)
  const [userId, setUserId] = useState('')
  const [role, setRole] = useState<TenantRole>('viewer')
  const [status, setStatus] = useState('active')

  const submit = (event: FormEvent) => {
    event.preventDefault()
    addMember.mutate(
      { user_id: userId.trim(), role, status },
      {
        onSuccess: () => {
          setUserId('')
          setRole('viewer')
          setStatus('active')
        },
      },
    )
  }

  if (!tenantId) {
    return <EmptyState title="请先选择组织" description="成员管理需要明确的组织上下文。可以在顶部选择已有组织，或先到组织页创建。" />
  }

  const firstError = members.error ?? addMember.error ?? updateMember.error ?? removeMember.error

  return (
    <div className="space-y-6">
      <div>
        <Badge tone="green">Members</Badge>
        <h1 className="mt-3 text-3xl font-black text-ink">成员管理</h1>
        <p className="mt-2 text-sm leading-6 text-muted">覆盖成员列表、添加、角色/状态更新和移除能力。viewer 执行管理动作时会由后端 RBAC 返回 403。</p>
      </div>

      {firstError ? <ErrorView error={firstError} title="成员操作失败" /> : null}

      <section className="grid gap-5 lg:grid-cols-[0.8fr_1.2fr]">
        <Card>
          <CardHeader title="添加成员" description={canManageMembers ? '需要目标用户已存在于 public.users。' : '当前角色只有成员查看权限，不能添加或更新成员。'} />
          {canManageMembers ? (
            <form className="space-y-4" onSubmit={submit}>
              <Input label="User ID" value={userId} onChange={(event) => setUserId(event.target.value)} placeholder="用户 UUID" required />
              <Select label="角色" value={role} onChange={(event) => setRole(event.target.value as TenantRole)}>
                {roles.map((item) => <option key={item} value={item}>{item}</option>)}
              </Select>
              <Select label="状态" value={status} onChange={(event) => setStatus(event.target.value)}>
                {statuses.map((item) => <option key={item} value={item}>{item}</option>)}
              </Select>
              <Button type="submit" isLoading={addMember.isPending} leftIcon={<UserPlus className="size-4" />}>添加/更新成员</Button>
            </form>
          ) : (
            <EmptyState title="只读成员权限" description="后端仍会拒绝 member:manage 之外的写操作；这里仅做前端体验层过滤。" />
          )}
        </Card>

        <Card>
          <CardHeader title="成员列表" description="调用 GET /api/v1/tenants/{tenant}/members。" />
          {members.isLoading ? <LoadingView label="加载成员..." /> : (members.data?.members.length ?? 0) === 0 ? (
            <EmptyState title="暂无成员" description="组织创建者应至少是 owner。如果没有成员，请检查后端数据。" />
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <thead>
                  <tr>
                    <Th>用户</Th>
                    <Th>角色</Th>
                    <Th>状态</Th>
                    <Th>操作</Th>
                  </tr>
                </thead>
                <tbody>
                  {members.data?.members.map((member) => (
                    <tr key={member.user_id}>
                      <Td>
                        <div className="font-bold text-ink">{member.display_name || member.email || member.user_id}</div>
                        <div className="mt-1 text-xs text-subtle">{member.user_id}</div>
                      </Td>
                      <Td><Badge tone={member.role === 'owner' ? 'green' : member.role === 'viewer' ? 'gray' : 'purple'}>{member.role}</Badge></Td>
                      <Td><Badge tone={member.status === 'active' ? 'green' : 'orange'}>{member.status}</Badge></Td>
                      <Td>
                        <div className="flex flex-wrap gap-2">
                          <Select
                            value={member.role}
                            onChange={(event) => updateMember.mutate({ userId: member.user_id, input: { role: event.target.value as TenantRole } })}
                            disabled={!canManageMembers}
                            className="h-8 min-w-28 bg-white text-xs"
                            aria-label="更新角色"
                          >
                            {roles.map((item) => <option key={item} value={item}>{item}</option>)}
                          </Select>
                          <Select
                            value={member.status}
                            onChange={(event) => updateMember.mutate({ userId: member.user_id, input: { status: event.target.value } })}
                            disabled={!canManageMembers}
                            className="h-8 min-w-28 bg-white text-xs"
                            aria-label="更新状态"
                          >
                            {statuses.map((item) => <option key={item} value={item}>{item}</option>)}
                          </Select>
                          <Button size="sm" variant="danger" disabled={!canManageMembers} onClick={() => removeMember.mutate(member.user_id)} isLoading={removeMember.isPending}>移除</Button>
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
