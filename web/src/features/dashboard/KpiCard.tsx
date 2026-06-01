import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { cn } from '@/shared/lib/cn'

export interface KpiCardProps {
  icon: ReactNode
  label: string
  value: string | number
  trend?: string
  trendUp?: boolean
  tone?: 'default' | 'brand' | 'warning' | 'success'
  to?: string
  loading?: boolean
}

export function KpiCard({ icon, label, value, trend, trendUp, tone = 'default', to, loading }: KpiCardProps) {
  const toneBorder: Record<string, string> = {
    brand: 'border-l-brand',
    warning: 'border-l-[#ea580c]',
    success: 'border-l-[#10b981]',
    default: 'border-l-line',
  }

  const content = (
    <div
      className={cn(
        'rounded-card border border-line bg-surface p-5 shadow-soft transition hover:border-brand-ring hover:shadow-brand',
        tone !== 'default' && `border-l-4 ${toneBorder[tone]}`,
        to && 'cursor-pointer',
      )}
    >
      {loading ? (
        <div className="space-y-3 animate-pulse">
          <div className="h-4 w-8 rounded bg-surface-soft" />
          <div className="h-8 w-16 rounded bg-surface-soft" />
          <div className="h-3 w-20 rounded bg-surface-soft" />
        </div>
      ) : (
        <>
          <div className="mb-2 flex items-center gap-2">
            <span className="text-subtle">{icon}</span>
            <span className="text-xs font-bold uppercase tracking-wide text-subtle">{label}</span>
          </div>
          <div className="text-2xl font-black text-ink tabular-nums">{value}</div>
          <div
            className={cn(
              'mt-1 text-xs font-semibold min-h-[1em]',
              trend
                ? trendUp === true
                  ? 'text-success-strong'
                  : trendUp === false
                  ? 'text-danger'
                  : 'text-subtle'
                : 'invisible',
            )}
            aria-hidden={!trend}
          >
            {trend
              ? `${trendUp === true ? '↑' : trendUp === false ? '↓' : ''} ${trend}`
              : ' '}
          </div>
        </>
      )}
    </div>
  )

  if (to && !loading) {
    return <Link to={to}>{content}</Link>
  }
  return content
}
