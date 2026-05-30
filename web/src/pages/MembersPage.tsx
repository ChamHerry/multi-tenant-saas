import { FormEvent, useMemo, useState } from 'react'
import { UserPlus } from 'lucide-react'
import { useCanTenant } from '@/features/access/access-hooks'
import { useAddMember, useMembers, useRemoveMember, useUpdateMember } from '@/features/members/member-hooks'
import { createAddMemberInputSchema } from '@/features/members/member-types'
import { useTenantI18n } from '@/features/tenants/tenant-i18n'
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
  const { schemaTranslator, t, formatRole, formatMemberStatus } = useTenantI18n()
  const addMemberSchema = useMemo(() => createAddMemberInputSchema(schemaTranslator), [schemaTranslator])

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(addMemberSchema, {
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
    return <EmptyState title={t('members.noTenantTitle', '请先选择组织')} description={t('members.noTenantDescription', '成员管理需要明确的组织上下文。可以在顶部选择已有组织，或先到组织页创建。')} />
  }

  const firstError = members.error ?? addMember.error ?? updateMember.error ?? removeMember.error

  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title={t('members.errorTitle', '成员操作失败')} /> : null}

      <div className="flex justify-end">
        <Button disabled={!canManageMembers} leftIcon={<UserPlus className="size-4" />} onClick={() => setIsAddMemberOpen(true)}>
          {t('members.addButton', '添加成员')}
        </Button>
      </div>

      <Dialog open={isAddMemberOpen} title={t('members.dialogTitle', '添加成员')} onClose={() => setIsAddMemberOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input label={t('members.userIdLabel', 'User ID')} value={userId} onChange={(event) => setUserId(event.target.value)} placeholder={t('members.userIdPlaceholder', '用户 UUID')} error={errors.user_id} />
          <Select label={t('members.roleLabel', '角色')} value={role} onChange={(event) => setRole(event.target.value as TenantRole)}>
            {roles.map((item) => <option key={item} value={item}>{formatRole(item)}</option>)}
          </Select>
          <Select label={t('members.statusLabel', '状态')} value={status} onChange={(event) => setStatus(event.target.value)}>
            {statuses.map((item) => <option key={item} value={item}>{formatMemberStatus(item)}</option>)}
          </Select>
          <Button type="submit" isLoading={addMember.isPending} leftIcon={<UserPlus className="size-4" />}>{t('members.submit', '添加/更新成员')}</Button>
        </form>
      </Dialog>

      <section>
        <Card>
          <CardHeader title={t('members.tableTitle', '成员列表')} description={t('members.tableDescription', '调用 GET /api/v1/tenants/{tenant}/members。')} />
          {members.isLoading ? <LoadingView label={t('members.loading', '加载成员...')} /> : (members.data?.members.length ?? 0) === 0 ? (
            <EmptyState title={t('members.emptyTitle', '暂无成员')} description={t('members.emptyDescription', '组织创建者应至少是 owner。如果没有成员，请检查后端数据。')} />
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <thead>
                  <tr>
                    <Th>{t('members.columns.user', '用户')}</Th>
                    <Th>{t('members.columns.role', '角色')}</Th>
                    <Th>{t('members.columns.status', '状态')}</Th>
                    <Th>{t('members.columns.actions', '操作')}</Th>
                  </tr>
                </thead>
                <tbody>
                  {members.data?.members.map((member) => (
                    <tr key={member.user_id}>
                      <Td>
                        <div className="font-bold text-ink">{member.display_name || member.email || member.user_id}</div>
                        <div className="mt-1 text-xs text-subtle">{member.user_id}</div>
                      </Td>
                      <Td><Badge tone={member.role === 'owner' ? 'green' : member.role === 'viewer' ? 'gray' : 'purple'}>{formatRole(member.role)}</Badge></Td>
                      <Td><Badge tone={member.status === 'active' ? 'green' : 'orange'}>{formatMemberStatus(member.status)}</Badge></Td>
                      <Td>
                        <div className="flex flex-wrap gap-2">
                          <Select
                            value={member.role}
                            onChange={(event) => updateMember.mutate({ userId: member.user_id, input: { role: event.target.value as TenantRole } })}
                            disabled={!canManageMembers}
                            className="h-8 min-w-28 bg-white text-xs"
                            aria-label={t('members.updateRoleAria', '更新角色')}
                          >
                            {roles.map((item) => <option key={item} value={item}>{formatRole(item)}</option>)}
                          </Select>
                          <Select
                            value={member.status}
                            onChange={(event) => updateMember.mutate({ userId: member.user_id, input: { status: event.target.value } })}
                            disabled={!canManageMembers}
                            className="h-8 min-w-28 bg-white text-xs"
                            aria-label={t('members.updateStatusAria', '更新状态')}
                          >
                            {statuses.map((item) => <option key={item} value={item}>{formatMemberStatus(item)}</option>)}
                          </Select>
                          <Button size="sm" variant="danger" disabled={!canManageMembers} onClick={() => removeMember.mutate(member.user_id)} isLoading={removeMember.isPending}>{t('members.remove', '移除')}</Button>
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
