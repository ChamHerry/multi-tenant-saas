import type { InputHTMLAttributes, ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: ReactNode
  hint?: ReactNode
  error?: ReactNode
}

export function Input({ className, label, hint, error, id, ...props }: InputProps) {
  return (
    <label className="block" htmlFor={id}>
      {label ? <span className="mb-1.5 block text-xs font-bold uppercase tracking-wide text-muted">{label}</span> : null}
      <input
        id={id}
        className={cn(
          'h-10 w-full rounded-control border border-line bg-surface-soft px-3 text-sm text-ink outline-none transition placeholder:text-subtle focus:border-brand focus:bg-white focus:ring-3 focus:ring-brand/10',
          error && 'border-danger focus:border-danger focus:ring-danger/10',
          className,
        )}
        {...props}
      />
      {error ? <span className="mt-1 block text-xs text-danger">{error}</span> : hint ? <span className="mt-1 block text-xs text-subtle">{hint}</span> : null}
    </label>
  )
}
