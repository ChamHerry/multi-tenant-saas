import { useMemo, useSyncExternalStore } from 'react'
import tenantEnUS from './locales/en-US.json'
import tenantZhCN from './locales/zh-CN.json'

export type SchemaMessageValue = string | number | boolean | null | undefined
export type SchemaMessageParams = Record<string, SchemaMessageValue>
export type SchemaTranslator = (key: string, fallback: string, params?: SchemaMessageParams) => string

export const TENANT_LOCALE_STORAGE_KEY = 'saas-template.locale'
export const TENANT_LOCALE_CHANGE_EVENT = 'saas-template:locale-change'

type LocaleShape<T> = {
  readonly [K in keyof T]: T[K] extends string ? string : LocaleShape<T[K]>
}

export type TenantLocale = 'en-US' | 'zh-CN'
export type TenantMessages = LocaleShape<typeof tenantEnUS>

const tenantLocaleResources: Record<TenantLocale, TenantMessages> = {
  'en-US': tenantEnUS,
  'zh-CN': tenantZhCN,
}

export type TenantValidationKey = keyof TenantMessages['validation']
export type TenantRoleKey = keyof TenantMessages['roles']
export type TenantStatusKey = keyof TenantMessages['tenantStatuses']
export type MemberStatusKey = keyof TenantMessages['memberStatuses']
export type InvitationStatusKey = keyof TenantMessages['invitationStatuses']

export const defaultTenantLocale: TenantLocale = 'zh-CN'
export const defaultTenantMessages = tenantZhCN
export const tenantValidationFallback = defaultTenantMessages.validation

export function formatSchemaMessage(template: string, params: SchemaMessageParams = {}) {
  return template.replace(/\{([A-Za-z0-9_]+)\}/g, (match, name: string) => {
    const value = params[name]
    return value === undefined || value === null ? match : String(value)
  })
}

export const defaultTenantSchemaTranslator: SchemaTranslator = (_key, fallback, params) => formatSchemaMessage(fallback, params)

function resolveTenantLocale(value?: string | null): TenantLocale | undefined {
  const normalized = value?.trim().toLowerCase()
  if (!normalized) return undefined
  if (normalized.startsWith('zh')) return 'zh-CN'
  if (normalized.startsWith('en')) return 'en-US'
  return undefined
}

function readBrowserLocales(): readonly string[] {
  if (typeof navigator === 'undefined') return []
  return navigator.languages?.length ? navigator.languages : [navigator.language]
}

function readTenantLocale(): TenantLocale {
  if (typeof window === 'undefined') return defaultTenantLocale
  const storedLocale = resolveTenantLocale(window.localStorage.getItem(TENANT_LOCALE_STORAGE_KEY))
  if (storedLocale) return storedLocale

  const htmlLocale = resolveTenantLocale(document.documentElement.lang)
  if (htmlLocale) return htmlLocale

  for (const locale of readBrowserLocales()) {
    const browserLocale = resolveTenantLocale(locale)
    if (browserLocale) return browserLocale
  }

  return defaultTenantLocale
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
