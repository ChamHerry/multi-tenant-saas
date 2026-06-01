import type { ReactNode } from 'react'

export interface KpiCardRowProps {
  children: ReactNode
}

export function KpiCardRow({ children }: KpiCardRowProps) {
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
      {children}
    </div>
  )
}
