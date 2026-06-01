import { useTranslation } from 'react-i18next'
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { Card, CardHeader } from '@/shared/ui/Card'
interface TrendChartProps {
  title: string
  data: ReadonlyArray<unknown>
  series: Array<{ key: string; color: string; name: string }>
  height?: number
  loading?: boolean
  emptyLabel?: string
}

export function TrendChart({ title, data, series, height = 200, loading, emptyLabel }: TrendChartProps) {
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
      ) : (
        <ResponsiveContainer width="100%" height={height}>
          <AreaChart data={data} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
            <XAxis dataKey="date" tick={{ fontSize: 10 }} stroke="#94a3b8" />
            <YAxis tick={{ fontSize: 10 }} stroke="#94a3b8" allowDecimals={false} />
            <Tooltip
              contentStyle={{ borderRadius: 8, border: '1px solid #e2e8f0', fontSize: 12 }}
            />
            {series.map((s) => (
              <Area
                key={s.key}
                type="monotone"
                dataKey={s.key}
                name={s.name}
                stroke={s.color}
                fill={s.color}
                fillOpacity={0.1}
                strokeWidth={2}
              />
            ))}
          </AreaChart>
        </ResponsiveContainer>
      )}
    </Card>
  )
}
