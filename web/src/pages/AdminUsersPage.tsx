import { FormEvent, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useCanPlatform } from '@/features/access/access-hooks'
import { useAdminUsers, usePlatformAdminMutations, usePlatformAdmins } from '@/features/admin/admin-hooks'
import { useUnlockUserMutation } from '@/features/auth/auth-hooks'
import { useDateTimeFormatter } from '@/shared/i18n'
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
  const { t } = useTranslation()
  const dateTimeFormatter = useDateTimeFormatter()
  const [query, setQuery] = useState('')
  const params = { query }
  const canManageUsers = useCanPlatform('platform:user:manage')
  const canManagePlatformAdmins = useCanPlatform('platform:admin:manage')
  const users = useAdminUsers(params)
  const platformAdmins = usePlatformAdmins(params, canManagePlatformAdmins)
  const mutations = usePlatformAdminMutations(params)
  const unlockUser = useUnlockUserMutation()
  const [userId, setUserId] = useState('')
  const [role, setRole] = useState('support')
  const [isGrantOpen, setIsGrantOpen] = useState(false)
  const firstError = users.error ?? platformAdmins.error ?? mutations.grant.error ?? mutations.revoke.error ?? mutations.updateUserStatus.error ?? unlockUser.error

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
      {firstError ? <ErrorView error={firstError} title={t('admin.users.failed')} /> : null}

      <div className="flex justify-end">
        <Button disabled={!canManagePlatformAdmins} onClick={() => setIsGrantOpen(true)}>
          {t('admin.users.grantAdmin')}
        </Button>
      </div>

      <Dialog open={isGrantOpen} title={t('admin.users.dialogTitle')} onClose={() => setIsGrantOpen(false)}>
        <form className="space-y-4" onSubmit={submit}>
          <Input label={t('admin.users.userId')} value={userId} onChange={(event) => setUserId(event.target.value)} required />
          <Select label={t('common.fields.role')} value={role} onChange={(event) => setRole(event.target.value)}>
            {roles.map((item) => (
              <option key={item} value={item}>
                {t(`admin.common.roles.${item}`, { defaultValue: item })}
              </option>
            ))}
          </Select>
          <Button type="submit" isLoading={mutations.grant.isPending}>
            {t('admin.users.grantAdmin')}
          </Button>
        </form>
      </Dialog>

      <Card>
        <CardHeader title={t('admin.users.searchUsers')} />
        <Input
          label={t('admin.common.search')}
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder={t('admin.users.searchPlaceholder')}
        />
      </Card>

      <Card>
        <CardHeader title={t('admin.users.platformAdminsTitle')} description={t('common.total', { count: platformAdmins.data?.total ?? 0 })} />
        {!canManagePlatformAdmins ? (
          <EmptyState
            title={t('admin.users.noPlatformAdminPermissionTitle')}
            description={t('admin.users.noPlatformAdminPermissionDescription')}
          />
        ) : platformAdmins.isLoading ? (
          <LoadingView label={t('admin.users.loadingPlatformAdmins')} />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <Th>{t('admin.users.userId')}</Th>
                  <Th>{t('common.fields.role')}</Th>
                  <Th>{t('common.fields.status')}</Th>
                  <Th>{t('admin.users.grantedAt')}</Th>
                  <Th>{t('common.fields.actions')}</Th>
                </tr>
              </thead>
              <tbody>
                {platformAdmins.data?.items.map((admin) => (
                  <tr key={admin.user_id}>
                    <Td>
                      <code className="text-xs">{admin.user_id}</code>
                    </Td>
                    <Td>
                      <Badge tone="red">{t(`admin.common.roles.${admin.role}`, { defaultValue: admin.role })}</Badge>
                    </Td>
                    <Td>
                      <Badge tone={admin.status === 'active' ? 'green' : 'orange'}>
                        {t(`common.status.${admin.status}`, { defaultValue: admin.status })}
                      </Badge>
                    </Td>
                    <Td>{dateTimeFormatter.format(new Date(admin.created_at))}</Td>
                    <Td>
                      <Button
                        size="sm"
                        variant="danger"
                        disabled={!canManagePlatformAdmins}
                        onClick={() => mutations.revoke.mutate(admin.user_id)}
                        isLoading={mutations.revoke.isPending}
                      >
                        {t('admin.users.revokePlatformPermission')}
                      </Button>
                    </Td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        )}
      </Card>

      <Card>
        <CardHeader title={t('admin.users.usersTitle')} description={t('common.total', { count: users.data?.total ?? 0 })} />
        {users.isLoading ? (
          <LoadingView label={t('admin.users.loadingUsers')} />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <Th>{t('common.fields.user')}</Th>
                  <Th>{t('common.fields.status')}</Th>
                  <Th>{t('admin.users.lastLogin')}</Th>
                  <Th>{t('common.fields.actions')}</Th>
                </tr>
              </thead>
              <tbody>
                {users.data?.items.map((user) => (
                  <tr key={user.id}>
                    <Td>
                      <div className="font-bold text-ink">{user.display_name || user.email}</div>
                      <div className="text-xs text-subtle">{user.id}</div>
                    </Td>
                    <Td>
                      <Badge tone={user.status === 'active' ? 'green' : 'orange'}>
                        {t(`common.status.${user.status}`, { defaultValue: user.status })}
                      </Badge>
                    </Td>
                    <Td>{user.last_login_at ? dateTimeFormatter.format(new Date(user.last_login_at)) : t('common.notAvailable')}</Td>
                    <Td>
                      <div className="flex flex-wrap gap-2">
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={!canManageUsers}
                          onClick={() => mutations.updateUserStatus.mutate({ userId: user.id, status: user.status === 'active' ? 'disabled' : 'active' })}
                          isLoading={mutations.updateUserStatus.isPending}
                        >
                          {user.status === 'active' ? t('common.actions.disable') : t('common.actions.enable')}
                        </Button>
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={!canManageUsers}
                          onClick={() => unlockUser.mutate(user.id)}
                          isLoading={unlockUser.isPending && unlockUser.variables === user.id}
                        >
                          {t('admin.users.unlockAccount')}
                        </Button>
                        <Button
                          size="sm"
                          variant="danger"
                          disabled={!canManagePlatformAdmins}
                          onClick={() => mutations.revoke.mutate(user.id)}
                          isLoading={mutations.revoke.isPending}
                        >
                          {t('admin.users.revokePlatformPermission')}
                        </Button>
                      </div>
                    </Td>
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
