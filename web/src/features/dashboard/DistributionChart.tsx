import { useTranslation } from 'react-i18next'
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { Card, CardHeader } from '@/shared/ui/Card'
import type { DistributionItem } from './dashboard-types'

interface DistributionChartProps {
  title: string
  data: DistributionItem[]
  layout?: 'horizontal' | 'vertical'
  height?: number
  loading?: boolean
  emptyLabel?: string
}

export function DistributionChart({
  title,
  data,
  layout = 'horizontal',
  height = 200,
  loading,
  emptyLabel,
}: DistributionChartProps) {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader title={title} />
      {loading ? (
        <div className="flex items-center justify-center animate-pulse" style={{ height }}>
          <div className="h-32 w-full rounded bg-surface-soft" />
        </div>
      ) : data.length === 0 ? (
        <div className="flex items-center justify-center text-sm text-muted" style={{ height }}>
          {emptyLabel ?? t('common.notAvailable')}
        </div>
      ) : layout === 'horizontal' ? (
        <ResponsiveContainer width="100%" height={height}>
          <BarChart data={data} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
            <XAxis dataKey="label" tick={{ fontSize: 10 }} stroke="#94a3b8" />
            <YAxis tick={{ fontSize: 10 }} stroke="#94a3b8" allowDecimals={false} />
            <Tooltip
              contentStyle={{ borderRadius: 8, border: '1px solid #e2e8f0', fontSize: 12 }}
            />
            <Bar dataKey="count" radius={[4, 4, 0, 0]} fill="#1761ff" />
          </BarChart>
        </ResponsiveContainer>
      ) : (
        <ResponsiveContainer width="100%" height={height}>
          <BarChart data={data} layout="vertical" margin={{ top: 4, right: 4, left: 40, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
            <XAxis type="number" tick={{ fontSize: 10 }} stroke="#94a3b8" allowDecimals={false} />
            <YAxis type="category" dataKey="label" tick={{ fontSize: 10 }} stroke="#94a3b8" width={80} />
            <Tooltip
              contentStyle={{ borderRadius: 8, border: '1px solid #e2e8f0', fontSize: 12 }}
            />
            <Bar dataKey="count" radius={[0, 4, 4, 0]} fill="#1761ff" />
          </BarChart>
        </ResponsiveContainer>
      )}
    </Card>
  )
}
