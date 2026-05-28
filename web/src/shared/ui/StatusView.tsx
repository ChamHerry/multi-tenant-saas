import { errorMessage } from '@/shared/api/errors'

export function LoadingView({ label = '加载中...' }: { label?: string }) {
  return <div className="rounded-card border border-line bg-white p-5 text-sm text-muted shadow-soft">{label}</div>
}

export function ErrorView({ error, title = '请求失败' }: { error: unknown; title?: string }) {
  return (
    <div className="rounded-card border border-danger/20 bg-red-50 p-5 text-sm text-danger">
      <div className="font-bold">{title}</div>
      <div className="mt-1 break-words">{errorMessage(error)}</div>
    </div>
  )
}
