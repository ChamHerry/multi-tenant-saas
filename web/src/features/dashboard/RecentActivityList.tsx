import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { Card, CardHeader } from '@/shared/ui/Card'
import { useDateTimeFormatter } from '@/shared/i18n/useDateTimeFormatter'
import type { ActivityItem } from './dashboard-types'

interface RecentActivityListProps {
  activities: ActivityItem[]
  auditLogLink: string
  loading?: boolean
}

export function RecentActivityList({ activities, auditLogLink, loading }: RecentActivityListProps) {
  const { t } = useTranslation()
  const formatDateTime = useDateTimeFormatter()

  return (
    <Card>
      <div className="flex items-center justify-between">
        <CardHeader title={t('dashboard.activity.title')} />
        <Link
          to={auditLogLink}
          className="mr-4 flex items-center gap-1 text-xs font-bold text-brand hover:underline"
        >
          {t('dashboard.activity.viewAll')} <ArrowRight className="size-3" />
        </Link>
      </div>
      {loading ? (
        <div className="space-y-3 px-4 pb-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-10 animate-pulse rounded bg-surface-soft" />
          ))}
        </div>
      ) : activities.length === 0 ? (
        <p className="px-4 pb-4 text-sm text-muted">{t('dashboard.activity.empty')}</p>
      ) : (
        <div className="space-y-1 px-2 pb-2">
          {activities.map((item, i) => (
            <div
              key={i}
              className="flex items-center justify-between rounded-panel px-3 py-2.5 text-sm transition hover:bg-surface-soft"
            >
              <div className="min-w-0 flex-1">
                <span className="font-semibold text-ink">
                  {item.user_display_name ?? t('common.notAvailable')}
                </span>
                <span className="ml-2 text-muted">{item.action}</span>
              </div>
              <time className="ml-3 shrink-0 text-xs text-subtle">
                {formatDateTime.format(new Date(item.created_at))}
              </time>
            </div>
          ))}
        </div>
      )}
    </Card>
  )
}
