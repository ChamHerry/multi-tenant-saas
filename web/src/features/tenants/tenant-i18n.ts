import { useMemo, useSyncExternalStore } from 'react'
import { LANGUAGE_STORAGE_KEY, defaultLanguage, type SupportedLanguage } from '@/shared/i18n/languages'
import { LOCALE_CHANGE_EVENT, readRuntimeLanguage } from '@/shared/i18n/locale-runtime'
import {
  defaultSchemaTranslator,
  formatSchemaMessage,
  type SchemaMessageParams,
  type SchemaTranslator,
} from '@/shared/lib/schema-translator'
import tenantEnUS from './locales/en-US.json'
import tenantZhCN from './locales/zh-CN.json'

export const TENANT_LOCALE_STORAGE_KEY = LANGUAGE_STORAGE_KEY
export const TENANT_LOCALE_CHANGE_EVENT = LOCALE_CHANGE_EVENT
export type { SchemaMessageParams, SchemaTranslator } from '@/shared/lib/schema-translator'

type LocaleShape<T> = {
  readonly [K in keyof T]: T[K] extends string ? string : LocaleShape<T[K]>
}

export type TenantMessages = LocaleShape<typeof tenantEnUS>

const tenantLocaleResources = {
  'en-US': tenantEnUS,
  'zh-CN': tenantZhCN,
} as const satisfies Record<SupportedLanguage, TenantMessages>

export type TenantLocale = keyof typeof tenantLocaleResources

export type TenantValidationKey = keyof TenantMessages['validation']
export type TenantRoleKey = keyof TenantMessages['roles']
export type TenantStatusKey = keyof TenantMessages['tenantStatuses']
export type MemberStatusKey = keyof TenantMessages['memberStatuses']
export type InvitationStatusKey = keyof TenantMessages['invitationStatuses']

export const defaultTenantLocale: TenantLocale = defaultLanguage
export const defaultTenantMessages = getTenantMessages(defaultTenantLocale)
export const tenantValidationFallback = defaultTenantMessages.validation

export const defaultTenantSchemaTranslator = defaultSchemaTranslator

function toTenantLocale(language: SupportedLanguage): TenantLocale {
  return language in tenantLocaleResources ? language as TenantLocale : defaultTenantLocale
}

function readTenantLocale(): TenantLocale {
  return toTenantLocale(readRuntimeLanguage({ includeDocument: true }))
}

function subscribeTenantLocale(onStoreChange: () => void) {
  if (typeof window === 'undefined') return () => undefined
  window.addEventListener('storage', onStoreChange)
  window.addEventListener('languagechange', onStoreChange)
  window.addEventListener(TENANT_LOCALE_CHANGE_EVENT, onStoreChange)

  const observer = typeof MutationObserver === 'undefined'
    ? undefined
    : new MutationObserver(onStoreChange)
  observer?.observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] })

  return () => {
    window.removeEventListener('storage', onStoreChange)
    window.removeEventListener('languagechange', onStoreChange)
    window.removeEventListener(TENANT_LOCALE_CHANGE_EVENT, onStoreChange)
    observer?.disconnect()
  }
}

export function getTenantMessages(locale: TenantLocale = defaultTenantLocale): TenantMessages {
  return tenantLocaleResources[locale]
}

function lookupMessage(source: unknown, key: string): string | undefined {
  let current = source
  for (const part of key.split('.')) {
    if (!current || typeof current !== 'object' || !(part in current)) return undefined
    current = (current as Record<string, unknown>)[part]
  }
  return typeof current === 'string' ? current : undefined
}

export function createTenantTextTranslator(locale: TenantLocale = defaultTenantLocale) {
  const messages = getTenantMessages(locale)
  const fallbackMessages = getTenantMessages(defaultTenantLocale)
  return (key: string, fallback: string, params?: SchemaMessageParams) => {
    const message = lookupMessage(messages, key) ?? lookupMessage(fallbackMessages, key) ?? fallback
    return formatSchemaMessage(message, params)
  }
}

export function createTenantSchemaTranslator(locale: TenantLocale = defaultTenantLocale): SchemaTranslator {
  const messages = getTenantMessages(locale).validation
  const fallbackMessages = tenantValidationFallback
  return (key, fallback, params) => {
    const message = messages[key as TenantValidationKey] ?? fallbackMessages[key as TenantValidationKey] ?? fallback
    return formatSchemaMessage(message, params)
  }
}

export function useTenantI18n() {
  const locale = useSyncExternalStore(subscribeTenantLocale, readTenantLocale, () => defaultTenantLocale)
  return useMemo(() => {
    const messages = getTenantMessages(locale)
    const t = createTenantTextTranslator(locale)
    const formatRole = (role: string) => messages.roles[role as TenantRoleKey] ?? role
    const formatTenantStatus = (status: string) => messages.tenantStatuses[status as TenantStatusKey] ?? status
    const formatMemberStatus = (status: string) => messages.memberStatuses[status as MemberStatusKey] ?? status
    const formatInvitationStatus = (status: string) => messages.invitationStatuses[status as InvitationStatusKey] ?? status
    const formatDateTime = (value: string) => new Date(value).toLocaleString(locale)

    return {
      locale,
      messages,
      schemaTranslator: createTenantSchemaTranslator(locale),
      t,
      formatRole,
      formatTenantStatus,
      formatMemberStatus,
      formatInvitationStatus,
      formatDateTime,
    }
  }, [locale])
}
