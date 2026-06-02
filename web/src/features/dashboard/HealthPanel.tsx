import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { cn } from '@/shared/lib/cn'
import type { SystemHealth } from './dashboard-types'

interface QuickAction {
  label: string
  to: string
  icon?: ReactNode
}

interface HealthPanelProps {
  health: SystemHealth
  quickActions: QuickAction[]
  loading?: boolean
}

function StatusDot({ status }: { status: string }) {
  return (
    <span
      className={cn(
        'inline-block size-2 rounded-full',
        status === 'ok' && 'bg-[#10b981]',
        status === 'error' && 'bg-danger',
        status !== 'ok' && status !== 'error' && 'bg-[#ea580c]',
      )}
    />
  )
}

export function HealthPanel({ health, quickActions, loading }: HealthPanelProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-3">
      <div className="rounded-card border border-line bg-surface p-5 shadow-soft">
        <h3 className="mb-3 text-sm font-black text-ink">{t('dashboard.health.title')}</h3>
        {loading ? (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-5 animate-pulse rounded bg-surface-soft" />
            ))}
          </div>
        ) : (
          <div className="space-y-2 text-xs">
            <div className="flex items-center justify-between">
              <span className="text-muted">{t('dashboard.health.api')}</span>
              <span className="flex items-center gap-1.5 font-semibold">
                <StatusDot status={health.api} />
                {health.api === 'ok' ? t('dashboard.health.ok') : health.api}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted">{t('dashboard.health.database')}</span>
              <span className="flex items-center gap-1.5 font-semibold">
                <StatusDot status={health.database} />
                {health.database === 'ok' ? t('dashboard.health.ok') : health.database}
              </span>
            </div>
            {health.migration_version > 0 ? (
              <div className="flex items-center justify-between">
                <span className="text-muted">{t('dashboard.health.migration')}</span>
                <span className="font-semibold text-ink">{t('dashboard.health.migrationVersionValue', { version: health.migration_version })}</span>
              </div>
            ) : null}
          </div>
        )}
      </div>

      <div className="rounded-card border border-line bg-surface p-5 shadow-soft">
        <h3 className="mb-3 text-sm font-black text-ink">{t('dashboard.quickActions.title')}</h3>
        <div className="flex flex-col gap-2">
          {quickActions.map((action) => (
            <Link
              key={action.to}
              to={action.to}
              className="flex items-center gap-2 rounded-panel border border-line bg-surface-soft px-3 py-2 text-xs font-bold text-ink transition hover:border-brand-ring hover:bg-brand-soft hover:text-brand"
            >
              {action.icon}
              {action.label}
            </Link>
          ))}
        </div>
      </div>
    </div>
  )
}
