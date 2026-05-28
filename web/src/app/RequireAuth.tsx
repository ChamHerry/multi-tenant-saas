import { Navigate, useLocation } from 'react-router-dom'
import { AppShell } from '@/layouts/AppShell'
import { useMe } from '@/features/auth/auth-hooks'
import { ApiError } from '@/shared/api/errors'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'

export function RequireAuth() {
  const location = useLocation()
  const me = useMe()

  if (me.isLoading) return <LoadingView label="正在检查登录状态..." />
  if (me.isError) {
    if (me.error instanceof ApiError && me.error.status === 401) {
      return <Navigate to="/login" state={{ from: location }} replace />
    }
    return <ErrorView title="登录状态检查失败" error={me.error} />
  }
  return <AppShell />
}
