import type { ReactNode } from 'react'
import { Button } from './Button'

export function Dialog({ open, title, children, onClose }: { open: boolean; title: ReactNode; children: ReactNode; onClose: () => void }) {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-ink/30 p-4">
      <div className="w-full max-w-lg rounded-card border border-line bg-white p-5 shadow-brand">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h3 className="text-lg font-bold text-ink">{title}</h3>
          <Button variant="ghost" size="sm" onClick={onClose}>关闭</Button>
        </div>
        {children}
      </div>
    </div>
  )
}
