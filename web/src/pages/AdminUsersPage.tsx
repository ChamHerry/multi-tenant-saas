import { FormEvent, useState } from 'react'
import { useCanPlatform } from '@/features/access/access-hooks'
import { useAdminUsers, usePlatformAdminMutations, usePlatformAdmins } from '@/features/admin/admin-hooks'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Dialog } from '@/shared/ui/Dialog'
import { EmptyState } from '@/shared/ui/EmptyState'
import { Input } from '@/shared/ui/Input'
import { Select } from '@/shared/ui/Select'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

const roles = ['super_admin', 'support', 'auditor']

export function AdminUsersPage() {
  const [query, setQuery] = useState('')
  const params = { query }
  const canManageUsers = useCanPlatform('platform:user:manage')
  const canManagePlatformAdmins = useCanPlatform('platform:admin:manage')
  const users = useAdminUsers(params)
  const platformAdmins = usePlatformAdmins(params, canManagePlatformAdmins)
  const mutations = usePlatformAdminMutations(params)
  const [userId, setUserId] = useState('')
  const [role, setRole] = useState('support')
  const [isGrantOpen, setIsGrantOpen] = useState(false)
  const firstError = users.error ?? platformAdmins.error ?? mutations.grant.error ?? mutations.revoke.error ?? mutations.updateUserStatus.error
  const submit = (event: FormEvent) => {
    event.preventDefault()
    mutations.grant.mutate(
      { userId: userId.trim(), role },
      {
        onSuccess: () => {
          setUserId('')
          setIsGrantOpen(false)
        },
      },
    )
  }
  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title="用户管理失败" /> : null}
      <div className="flex justify-end">
        <Button disabled={!canManagePlatformAdmins} onClick={() => setIsGrantOpen(true)}>授权平台管理员</Button>
      </div>

      <Dialog open={isGrantOpen} title="授权平台管理员" onClose={() => setIsGrantOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input label="User ID" value={userId} onChange={(event) => setUserId(event.target.value)} required />
          <Select label="角色" value={role} onChange={(event) => setRole(event.target.value)}>
            {roles.map((item) => <option key={item} value={item}>{item}</option>)}
          </Select>
          <Button type="submit" isLoading={mutations.grant.isPending}>授权</Button>
        </form>
      </Dialog>

      <Card><CardHeader title="搜索用户" /><Input label="搜索" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="email / name / id" /></Card>
      <Card><CardHeader title="平台管理员列表" description={`Total: ${platformAdmins.data?.total ?? 0}`} />{!canManagePlatformAdmins ? <EmptyState title="无平台管理员管理权限" description="需要 platform:admin:manage 才能查看和维护平台管理员列表。" /> : platformAdmins.isLoading ? <LoadingView label="加载平台管理员..." /> : <div className="overflow-x-auto"><Table><thead><tr><Th>User ID</Th><Th>角色</Th><Th>状态</Th><Th>授权时间</Th><Th>操作</Th></tr></thead><tbody>{platformAdmins.data?.items.map((admin) => <tr key={admin.user_id}><Td><code className="text-xs">{admin.user_id}</code></Td><Td><Badge tone="red">{admin.role}</Badge></Td><Td><Badge tone={admin.status === 'active' ? 'green' : 'orange'}>{admin.status}</Badge></Td><Td>{new Date(admin.created_at).toLocaleString()}</Td><Td><Button size="sm" variant="danger" disabled={!canManagePlatformAdmins} onClick={() => mutations.revoke.mutate(admin.user_id)} isLoading={mutations.revoke.isPending}>撤销平台权限</Button></Td></tr>)}</tbody></Table></div>}</Card>
      <Card><CardHeader title="用户列表" description={`Total: ${users.data?.total ?? 0}`} />{users.isLoading ? <LoadingView label="加载用户..." /> : <div className="overflow-x-auto"><Table><thead><tr><Th>用户</Th><Th>状态</Th><Th>最后登录</Th><Th>操作</Th></tr></thead><tbody>{users.data?.items.map((user) => <tr key={user.id}><Td><div className="font-bold text-ink">{user.display_name || user.email}</div><div className="text-xs text-subtle">{user.id}</div></Td><Td><Badge tone={user.status === 'active' ? 'green' : 'orange'}>{user.status}</Badge></Td><Td>{user.last_login_at ? new Date(user.last_login_at).toLocaleString() : '-'}</Td><Td><div className="flex flex-wrap gap-2"><Button size="sm" variant="secondary" disabled={!canManageUsers} onClick={() => mutations.updateUserStatus.mutate({ userId: user.id, status: user.status === 'active' ? 'disabled' : 'active' })} isLoading={mutations.updateUserStatus.isPending}>{user.status === 'active' ? '禁用' : '启用'}</Button><Button size="sm" variant="danger" disabled={!canManagePlatformAdmins} onClick={() => mutations.revoke.mutate(user.id)} isLoading={mutations.revoke.isPending}>撤销平台权限</Button></div></Td></tr>)}</tbody></Table></div>}</Card>
    </div>
  )
}
