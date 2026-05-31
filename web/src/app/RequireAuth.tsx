import { Navigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AppShell } from '@/layouts/AppShell'
import { useMe } from '@/features/auth/auth-hooks'
import { ApiError } from '@/shared/api/errors'
import { ErrorView, LoadingView } from '@/shared/ui'

export function RequireAuth() {
  const location = useLocation()
  const { t } = useTranslation()
  const me = useMe()

  if (me.isLoading) return <LoadingView label={t('auth.guard.checkingSession')} />
  if (me.isError) {
    if (me.error instanceof ApiError && me.error.status === 401) {
      return <Navigate to="/login" state={{ from: location }} replace />
    }
    return <ErrorView title={t('auth.guard.sessionCheckFailed')} error={me.error} />
  }
  return <AppShell />
}
