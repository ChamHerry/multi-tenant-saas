import { useMemo, useSyncExternalStore } from 'react'
import { LOCALE_CHANGE_EVENT, readRuntimeLanguage } from '@/shared/i18n/locale-runtime'
import { defaultLanguage, type SupportedLanguage } from '@/shared/i18n/languages'
import { defaultSchemaTranslator, formatSchemaMessage, type SchemaTranslator } from '@/shared/lib/schema-translator'
import authEnUS from './locales/en-US.json'
import authZhCN from './locales/zh-CN.json'

const authLocaleResources = {
  'en-US': authEnUS,
  'zh-CN': authZhCN,
} as const

export type AuthLocale = keyof typeof authLocaleResources
type LocalizedMessages<T> = {
  readonly [Key in keyof T]: T[Key] extends string ? string : LocalizedMessages<T[Key]>
}

export type AuthMessages = LocalizedMessages<typeof authEnUS>
export type AuthValidationKey = keyof AuthMessages['validation']
export const defaultAuthLocale: AuthLocale = defaultLanguage

function toAuthLocale(language: SupportedLanguage): AuthLocale {
  return language in authLocaleResources ? language as AuthLocale : defaultAuthLocale
}

function readAuthLocale(): AuthLocale {
  return toAuthLocale(readRuntimeLanguage({ includeDocument: true }))
}

function subscribeAuthLocale(onStoreChange: () => void) {
  if (typeof window === 'undefined') return () => undefined
  window.addEventListener('storage', onStoreChange)
  window.addEventListener('languagechange', onStoreChange)
  window.addEventListener(LOCALE_CHANGE_EVENT, onStoreChange)
  return () => {
    window.removeEventListener('storage', onStoreChange)
    window.removeEventListener('languagechange', onStoreChange)
    window.removeEventListener(LOCALE_CHANGE_EVENT, onStoreChange)
  }
}

export function getAuthMessages(locale: AuthLocale = defaultAuthLocale): AuthMessages {
  return authLocaleResources[locale]
}

export function createAuthSchemaTranslator(locale: AuthLocale = defaultAuthLocale): SchemaTranslator {
  const messages = getAuthMessages(locale).validation
  const fallbacks = getAuthMessages(defaultAuthLocale).validation
  return (key, fallback, params) => {
    const message = messages[key as AuthValidationKey] ?? fallbacks[key as AuthValidationKey] ?? fallback
    return formatSchemaMessage(message, params)
  }
}

export function useAuthI18n() {
  const locale = useSyncExternalStore(subscribeAuthLocale, readAuthLocale, () => defaultAuthLocale)
  return useMemo(() => {
    const messages = getAuthMessages(locale)
    return {
      locale,
      messages,
      schemaTranslator: createAuthSchemaTranslator(locale),
    }
  }, [locale])
}

export const defaultAuthMessages = getAuthMessages(defaultAuthLocale)
export const defaultAuthSchemaTranslator = defaultSchemaTranslator
