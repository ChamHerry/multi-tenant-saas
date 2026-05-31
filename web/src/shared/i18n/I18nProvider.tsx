import { useEffect, useState, type ReactNode } from 'react'
import { I18nextProvider } from 'react-i18next'
import { defaultLanguage, normalizeLanguage, type SupportedLanguage } from './languages'
import { getCurrentLanguage, i18next, syncDocumentLanguage } from './i18n'

export function I18nProvider({ children }: { children: ReactNode }) {
  return (
    <I18nextProvider i18n={i18next}>
      <LanguageSideEffects />
      {children}
    </I18nextProvider>
  )
}

function LanguageSideEffects() {
  const [language, setLanguage] = useState<SupportedLanguage>(getCurrentLanguage)

  useEffect(() => {
    syncDocumentLanguage(language)
  }, [language])

  useEffect(() => {
    const handleLanguageChanged = (nextLanguage: string) => {
      setLanguage(normalizeLanguage(nextLanguage) ?? defaultLanguage)
    }

    i18next.on('languageChanged', handleLanguageChanged)
    return () => {
      i18next.off('languageChanged', handleLanguageChanged)
    }
  }, [])

  return null
}
