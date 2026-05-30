import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { LanguageSwitcher } from '@/shared/i18n'

const featureKeys = ['auth.shell.features.httpOnly', 'auth.shell.features.api', 'auth.shell.features.orgContext']

export function AuthShell({ children }: { children: ReactNode }) {
  const { t } = useTranslation()

  return (
    <main className="min-h-screen bg-page p-4 sm:p-8">
      <div className="fixed right-4 top-4 z-10">
        <LanguageSwitcher />
      </div>
      <div className="hero-gradient mx-auto grid min-h-[calc(100vh-4rem)] max-w-6xl items-center rounded-[28px] border border-white/80 p-6 shadow-brand sm:p-10 lg:grid-cols-[1.1fr_0.9fr]">
        <section className="max-w-xl">
          <span className="inline-flex rounded-full bg-brand-soft px-3 py-1 text-xs font-bold text-brand">
            {t('auth.shell.badge')}
          </span>
          <h1 className="mt-6 text-4xl font-black leading-tight text-ink sm:text-5xl">
            {t('auth.shell.titlePrefix')} <span className="text-gradient">{t('auth.shell.titleHighlight')}</span>
          </h1>
          <p className="mt-5 text-base leading-8 text-muted">{t('auth.shell.description')}</p>
          <div className="mt-8 grid gap-3 sm:grid-cols-3">
            {featureKeys.map((key) => (
              <div
                key={key}
                className="rounded-card border border-line bg-white/80 p-4 text-sm font-bold text-brand shadow-soft"
              >
                {t(key)}
              </div>
            ))}
          </div>
        </section>
        <section>{children}</section>
      </div>
    </main>
  )
}
