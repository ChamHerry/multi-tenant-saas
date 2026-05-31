import { useId } from 'react'
import { cn } from '@/shared/lib/cn'
import { useLanguage } from './useLanguage'
import type { SupportedLanguage } from './languages'

type LanguageSwitcherProps = {
  className?: string
}

export function LanguageSwitcher({ className }: LanguageSwitcherProps) {
  const id = useId()
  const { language, languages, setLanguage, t } = useLanguage()
  const label = t('language.switcherLabel')

  return (
    <label className={cn('inline-flex items-center gap-2 text-sm font-bold text-muted', className)} htmlFor={id}>
      <span className="sr-only">{label}</span>
      <select
        id={id}
        aria-label={label}
        className="h-9 rounded-control border border-line bg-white px-3 text-sm font-bold text-ink outline-none transition hover:border-brand focus:border-brand focus:ring-3 focus:ring-brand/10"
        value={language}
        onChange={(event) => void setLanguage(event.target.value as SupportedLanguage)}
      >
        {languages.map((item) => (
          <option key={item.code} value={item.code}>
            {t(item.labelKey)}
          </option>
        ))}
      </select>
    </label>
  )
}
