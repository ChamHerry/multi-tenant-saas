import { useMemo } from 'react'
import { useLanguage } from './useLanguage'

export function useDateTimeFormatter(options?: Intl.DateTimeFormatOptions) {
  const { language } = useLanguage()
  return useMemo(() => new Intl.DateTimeFormat(language, options), [language, options])
}
