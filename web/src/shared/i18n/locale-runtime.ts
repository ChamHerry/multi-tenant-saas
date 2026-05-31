import { readStorage } from '@/shared/lib/storage'
import {
  LANGUAGE_STORAGE_KEY,
  normalizeLanguage,
  resolveInitialLanguage,
  type SupportedLanguage,
} from './languages'

export const LOCALE_CHANGE_EVENT = 'saas-template:locale-change'

export function readBrowserLanguages(): readonly string[] {
  if (typeof navigator === 'undefined') return []
  try {
    return navigator.languages?.length ? navigator.languages : [navigator.language]
  } catch {
    return []
  }
}

function readDocumentLanguage(): string | undefined {
  if (typeof document === 'undefined') return undefined
  return document.documentElement.lang || undefined
}

export function readStoredLanguage(): SupportedLanguage | undefined {
  return normalizeLanguage(readStorage(LANGUAGE_STORAGE_KEY))
}

export function readRuntimeLanguage({ includeDocument = false }: { includeDocument?: boolean } = {}): SupportedLanguage {
  return resolveInitialLanguage({
    storedLanguage: readStorage(LANGUAGE_STORAGE_KEY),
    browserLanguages: [
      ...(includeDocument ? [readDocumentLanguage()] : []),
      ...readBrowserLanguages(),
    ],
  })
}
