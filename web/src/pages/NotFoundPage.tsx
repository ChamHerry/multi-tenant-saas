import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/ui/Button'
import { Card } from '@/shared/ui/Card'

export function NotFoundPage() {
  const { t } = useTranslation()

  return (
    <Card className="mx-auto max-w-lg text-center">
      <div className="text-6xl font-black text-gradient">404</div>
      <p className="mt-2 text-sm text-muted">{t('notFound.description')}</p>
      <Link className="mt-5 inline-flex" to="/"><Button>{t('notFound.backHome')}</Button></Link>
    </Card>
  )
}
