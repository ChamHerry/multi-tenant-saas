import type { ReactNode } from 'react'

export function AuthShell({ children }: { children: ReactNode }) {
  return (
    <main className="min-h-screen bg-page p-4 sm:p-8">
      <div className="hero-gradient mx-auto grid min-h-[calc(100vh-4rem)] max-w-6xl items-center rounded-[28px] border border-white/80 p-6 shadow-brand sm:p-10 lg:grid-cols-[1.1fr_0.9fr]">
        <section className="max-w-xl">
          <span className="inline-flex rounded-full bg-brand-soft px-3 py-1 text-xs font-bold text-brand">SaaS Template Console</span>
          <h1 className="mt-6 text-4xl font-black leading-tight text-ink sm:text-5xl">
            多租户 SaaS 模板的 <span className="text-gradient">安全控制台</span>
          </h1>
          <p className="mt-5 text-base leading-8 text-muted">
            使用正式账号登录后，可以创建组织、切换组织上下文、管理成员和 API Key。浏览器登录态由 HttpOnly Cookie 承载，机器访问继续使用 API Key。
          </p>
          <div className="mt-8 grid gap-3 sm:grid-cols-3">
            {['HttpOnly Cookie', 'GoFrame API', 'Organization Context'].map((item) => (
              <div key={item} className="rounded-card border border-line bg-white/80 p-4 text-sm font-bold text-brand shadow-soft">
                {item}
              </div>
            ))}
          </div>
        </section>
        <section>{children}</section>
      </div>
    </main>
  )
}
