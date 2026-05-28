import type { HTMLAttributes } from 'react'
import { cn } from '@/shared/lib/cn'

type BadgeTone = 'blue' | 'purple' | 'green' | 'orange' | 'red' | 'gray'

export function Badge({ className, tone = 'blue', ...props }: HTMLAttributes<HTMLSpanElement> & { tone?: BadgeTone }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-1 text-xs font-bold',
        tone === 'blue' && 'bg-brand-soft text-brand',
        tone === 'purple' && 'bg-accent-purple-soft text-accent-purple',
        tone === 'green' && 'bg-success-soft text-success-strong',
        tone === 'orange' && 'bg-warning-soft text-warning',
        tone === 'red' && 'bg-red-50 text-danger',
        tone === 'gray' && 'bg-surface-soft text-muted',
        className,
      )}
      {...props}
    />
  )
}
