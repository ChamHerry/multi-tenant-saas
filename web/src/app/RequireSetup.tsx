import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useSetupState } from '@/features/setup/setup-hooks'
import { ErrorView, LoadingView } from '@/shared/ui'

export function RequireSetup({ children, setupRoute = false }: { children: ReactNode; setupRoute?: boolean }) {
  const { t } = useTranslation()
  const location = useLocation()
  const setup = useSetupState()

  if (setup.isLoading) return <LoadingView label={t('setup.loading')} />
  if (setup.isError) return <ErrorView title={t('setup.failed')} error={setup.error} />

  if (setup.data?.requires_setup && !setupRoute) {
    return <Navigate to="/setup" state={{ from: location }} replace />
  }
  if (!setup.data?.requires_setup && setupRoute) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}
