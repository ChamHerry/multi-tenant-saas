import type { TableHTMLAttributes } from 'react'
import { cn } from '@/shared/lib/cn'

export function Table({ className, ...props }: TableHTMLAttributes<HTMLTableElement>) {
  return <table className={cn('w-full border-separate border-spacing-0 text-left text-sm', className)} {...props} />
}

export function Th({ className, ...props }: React.ThHTMLAttributes<HTMLTableCellElement>) {
  return <th className={cn('border-b border-line-soft px-3 py-2 text-xs font-bold uppercase tracking-wide text-subtle', className)} {...props} />
}

export function Td({ className, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) {
  return <td className={cn('border-b border-line-soft px-3 py-3 align-top text-muted', className)} {...props} />
}
