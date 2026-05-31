export const LANGUAGE_STORAGE_KEY = 'saas-template.locale'

export const defaultLanguage = 'zh-CN'

export const supportedLanguages = [
  { code: 'zh-CN', labelKey: 'language.options.zhCN', htmlLang: 'zh-CN' },
  { code: 'en-US', labelKey: 'language.options.enUS', htmlLang: 'en' },
] as const

export type SupportedLanguage = (typeof supportedLanguages)[number]['code']

type LanguageLike = string | null | undefined

const languageAliases: Record<string, SupportedLanguage> = {
  cn: 'zh-CN',
  en: 'en-US',
  'en-us': 'en-US',
  'en_us': 'en-US',
  zh: 'zh-CN',
  'zh-cn': 'zh-CN',
  'zh-hans': 'zh-CN',
  'zh-hans-cn': 'zh-CN',
  'zh_cn': 'zh-CN',
}

export function isSupportedLanguage(value: LanguageLike): value is SupportedLanguage {
  return supportedLanguages.some((language) => language.code === value)
}

export function normalizeLanguage(value: LanguageLike): SupportedLanguage | undefined {
  const normalized = value?.trim().toLowerCase()
  if (!normalized) return undefined

  const alias = languageAliases[normalized]
  if (alias) return alias
  if (normalized.startsWith('zh')) return 'zh-CN'
  if (normalized.startsWith('en')) return 'en-US'
  return undefined
}

export function getHtmlLang(language: SupportedLanguage) {
  return supportedLanguages.find((item) => item.code === language)?.htmlLang ?? language
}

export function resolveInitialLanguage(options: {
  storedLanguage?: LanguageLike
  browserLanguages?: readonly LanguageLike[]
}): SupportedLanguage {
  const storedLanguage = normalizeLanguage(options.storedLanguage)
  if (storedLanguage) return storedLanguage

  for (const language of options.browserLanguages ?? []) {
    const browserLanguage = normalizeLanguage(language)
    if (browserLanguage) return browserLanguage
  }

  return defaultLanguage
}
