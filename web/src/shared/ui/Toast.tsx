import { cn } from '@/shared/lib/cn'

export function Toast({ message, tone = 'blue' }: { message?: string; tone?: 'blue' | 'green' | 'red' }) {
  if (!message) return null
  return (
    <div className={cn('rounded-panel border px-3 py-2 text-sm', tone === 'blue' && 'border-brand-ring bg-brand-soft text-brand', tone === 'green' && 'border-success/30 bg-success-soft text-success-strong', tone === 'red' && 'border-danger/30 bg-red-50 text-danger')}>
      {message}
    </div>
  )
}
