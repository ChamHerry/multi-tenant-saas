import type { ReactNode, SelectHTMLAttributes } from 'react'
import { cn } from '@/shared/lib/cn'

type SelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  label?: ReactNode
  hint?: ReactNode
}

export function Select({ className, label, hint, id, children, ...props }: SelectProps) {
  return (
    <label className="block" htmlFor={id}>
      {label ? <span className="mb-1.5 block text-xs font-bold uppercase tracking-wide text-muted">{label}</span> : null}
      <select
        id={id}
        className={cn('h-10 w-full rounded-control border border-line bg-surface-soft px-3 text-sm text-ink outline-none transition focus:border-brand focus:bg-white focus:ring-3 focus:ring-brand/10', className)}
        {...props}
      >
        {children}
      </select>
      {hint ? <span className="mt-1 block text-xs text-subtle">{hint}</span> : null}
    </label>
  )
}
