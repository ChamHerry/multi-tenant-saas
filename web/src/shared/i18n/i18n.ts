import i18next from 'i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import { initReactI18next } from 'react-i18next'
import { writeStorage } from '@/shared/lib/storage'
import { LANGUAGE_STORAGE_KEY, defaultLanguage, getHtmlLang, normalizeLanguage, type SupportedLanguage } from './languages'
import { LOCALE_CHANGE_EVENT, readRuntimeLanguage } from './locale-runtime'
import { resources } from './resources'

export function detectInitialLanguage(): SupportedLanguage {
  return readRuntimeLanguage()
}

export function syncDocumentLanguage(language: SupportedLanguage) {
  if (typeof document === 'undefined') return
  document.documentElement.lang = getHtmlLang(language)
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent(LOCALE_CHANGE_EVENT, { detail: { language } }))
  }
}

export function persistLanguage(language: SupportedLanguage) {
  writeStorage(LANGUAGE_STORAGE_KEY, language)
}

export function getCurrentLanguage(): SupportedLanguage {
  return normalizeLanguage(i18next.resolvedLanguage ?? i18next.language) ?? defaultLanguage
}

export async function changeAppLanguage(language: SupportedLanguage) {
  persistLanguage(language)
  await i18next.changeLanguage(language)
  syncDocumentLanguage(language)
}

const initialLanguage = detectInitialLanguage()

if (!i18next.isInitialized) {
  void i18next.use(LanguageDetector).use(initReactI18next).init({
    resources,
    fallbackLng: defaultLanguage,
    supportedLngs: Object.keys(resources),
    detection: {
      order: ['localStorage', 'navigator'],
      lookupLocalStorage: LANGUAGE_STORAGE_KEY,
      caches: [],
      convertDetectedLanguage: (language) => normalizeLanguage(language) ?? defaultLanguage,
    },
    interpolation: {
      escapeValue: false,
    },
    react: {
      useSuspense: false,
    },
    returnNull: false,
    initAsync: false,
  })
}

syncDocumentLanguage(initialLanguage)

export { i18next }
