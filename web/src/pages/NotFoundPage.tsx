import { Link } from 'react-router-dom'
import { Button } from '@/shared/ui/Button'
import { Card } from '@/shared/ui/Card'

export function NotFoundPage() {
  return (
    <Card className="mx-auto max-w-lg text-center">
      <div className="text-6xl font-black text-gradient">404</div>
      <h1 className="mt-4 text-2xl font-black text-ink">页面不存在</h1>
      <p className="mt-2 text-sm text-muted">该路径会由 GoFrame SPA fallback 返回前端入口，但前端路由没有匹配页面。</p>
      <Link className="mt-5 inline-flex" to="/"><Button>回到仪表盘</Button></Link>
    </Card>
  )
}
