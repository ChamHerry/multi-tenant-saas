import { useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { changeAppLanguage, getCurrentLanguage } from './i18n'
import { normalizeLanguage, supportedLanguages, type SupportedLanguage } from './languages'

export function useLanguage() {
  const { i18n, t } = useTranslation()
  const language = normalizeLanguage(i18n.resolvedLanguage ?? i18n.language) ?? getCurrentLanguage()

  const setLanguage = useCallback(async (nextLanguage: SupportedLanguage) => {
    await changeAppLanguage(nextLanguage)
  }, [])

  return {
    language,
    languages: supportedLanguages,
    setLanguage,
    t,
  }
}
