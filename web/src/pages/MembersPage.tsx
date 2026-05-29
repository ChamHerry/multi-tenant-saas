import { FormEvent, useState } from 'react'
import { UserPlus } from 'lucide-react'
import { useCanTenant } from '@/features/access/access-hooks'
import { useAddMember, useMembers, useRemoveMember, useUpdateMember } from '@/features/members/member-hooks'
import { addMemberInputSchema } from '@/features/members/member-types'
import { tenantRoleSchema, type TenantRole } from '@/features/tenants/tenant-types'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge, Button, Card, CardHeader, Dialog, EmptyState, Input, Select, ErrorView, LoadingView, Table, Td, Th } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'

const roles = tenantRoleSchema.options
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
  const [isAddMemberOpen, setIsAddMemberOpen] = useState(false)
  const [errors, setErrors] = useState<FieldErrors>({})

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(addMemberInputSchema, {
      user_id: userId.trim(),
      role,
      status,
    })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    addMember.mutate(result.data, {
      onSuccess: () => {
        setUserId('')
        setRole('viewer')
        setStatus('active')
        setIsAddMemberOpen(false)
      },
    })
  }

  if (!tenantId) {
    return <EmptyState title="请先选择组织" description="成员管理需要明确的组织上下文。可以在顶部选择已有组织，或先到组织页创建。" />
  }

  const firstError = members.error ?? addMember.error ?? updateMember.error ?? removeMember.error

  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title="成员操作失败" /> : null}

      <div className="flex justify-end">
        <Button disabled={!canManageMembers} leftIcon={<UserPlus className="size-4" />} onClick={() => setIsAddMemberOpen(true)}>
          添加成员
        </Button>
      </div>

      <Dialog open={isAddMemberOpen} title="添加成员" onClose={() => setIsAddMemberOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input label="User ID" value={userId} onChange={(event) => setUserId(event.target.value)} placeholder="用户 UUID" error={errors.user_id} />
          <Select label="角色" value={role} onChange={(event) => setRole(event.target.value as TenantRole)}>
            {roles.map((item) => <option key={item} value={item}>{item}</option>)}
          </Select>
          <Select label="状态" value={status} onChange={(event) => setStatus(event.target.value)}>
            {statuses.map((item) => <option key={item} value={item}>{item}</option>)}
          </Select>
          <Button type="submit" isLoading={addMember.isPending} leftIcon={<UserPlus className="size-4" />}>添加/更新成员</Button>
        </form>
      </Dialog>

      <section>
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
