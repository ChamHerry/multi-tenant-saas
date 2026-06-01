import { useTranslation } from 'react-i18next'
import { useMe } from '@/features/auth/auth-hooks'
import { usePageTitle } from '@/layouts/page-title'
import { pageTitle } from '@/layouts/page-title'
import { DashboardShell } from '@/features/dashboard/DashboardShell'

export function DashboardPage() {
  const { t } = useTranslation()
  const me = useMe()
  const displayName = me.data?.user.display_name || me.data?.user.email || t('dashboard.userFallback')

  usePageTitle(
    me.data ? pageTitle('dashboard.title.welcome', { name: displayName }) : pageTitle('dashboard.title.default'),
  )

  return <DashboardShell />
}
