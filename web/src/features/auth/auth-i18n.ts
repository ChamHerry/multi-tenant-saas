import { useMemo, useSyncExternalStore } from 'react'
import { defaultSchemaTranslator, formatSchemaMessage, type SchemaTranslator } from '@/shared/lib/schema-translator'
import { authEnUS } from './locales/en-US'
import { authZhCN } from './locales/zh-CN'

const authLocaleResources = {
  'en-US': authEnUS,
  'zh-CN': authZhCN,
} as const

export type AuthLocale = keyof typeof authLocaleResources
export type AuthMessages = typeof authEnUS
export type AuthValidationKey = keyof AuthMessages['validation']
export const AUTH_LOCALE_STORAGE_KEY = 'saas-template.locale'

function resolveAuthLocale(value?: string | null): AuthLocale {
  if (value?.toLowerCase().startsWith('zh')) return 'zh-CN'
  return 'en-US'
}

function readAuthLocale(): AuthLocale {
  if (typeof window === 'undefined') return 'en-US'
  return resolveAuthLocale(window.localStorage.getItem(AUTH_LOCALE_STORAGE_KEY) ?? document.documentElement.lang ?? navigator.language)
}

function subscribeAuthLocale(onStoreChange: () => void) {
  if (typeof window === 'undefined') return () => undefined
  window.addEventListener('storage', onStoreChange)
  window.addEventListener('languagechange', onStoreChange)
  window.addEventListener('saas-template:locale-change', onStoreChange)
  return () => {
    window.removeEventListener('storage', onStoreChange)
    window.removeEventListener('languagechange', onStoreChange)
    window.removeEventListener('saas-template:locale-change', onStoreChange)
  }
}

export function getAuthMessages(locale: AuthLocale = 'en-US'): AuthMessages {
  return authLocaleResources[locale]
}

export function createAuthSchemaTranslator(locale: AuthLocale = 'en-US'): SchemaTranslator {
  const messages = getAuthMessages(locale).validation
  const fallbacks = authEnUS.validation
  return (key, fallback, params) => {
    const message = messages[key as AuthValidationKey] ?? fallbacks[key as AuthValidationKey] ?? fallback
    return formatSchemaMessage(message, params)
  }
}

export function useAuthI18n() {
  const locale = useSyncExternalStore(subscribeAuthLocale, readAuthLocale, () => 'en-US')
  return useMemo(() => {
    const messages = getAuthMessages(locale)
    return {
      locale,
      messages,
      schemaTranslator: createAuthSchemaTranslator(locale),
    }
  }, [locale])
}

export const defaultAuthMessages = authEnUS
export const defaultAuthSchemaTranslator = defaultSchemaTranslator
