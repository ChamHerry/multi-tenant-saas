import { useTranslation } from 'react-i18next'
import { errorMessage } from '@/shared/api/errors'

export function LoadingView({ label }: { label?: string }) {
  const { t } = useTranslation()
  return (
    <div className="rounded-card border border-line bg-white p-5 text-sm text-muted shadow-soft">
      {label ?? t('common.loading')}
    </div>
  )
}

export function ErrorView({ error, title }: { error: unknown; title?: string }) {
  const { t } = useTranslation()

  return (
    <div className="rounded-card border border-danger/20 bg-red-50 p-5 text-sm text-danger">
      <div className="font-bold">{title ?? t('common.requestFailed')}</div>
      <div className="mt-1 break-words">{errorMessage(error)}</div>
    </div>
  )
}
