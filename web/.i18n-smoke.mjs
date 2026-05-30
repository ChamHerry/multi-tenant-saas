import { createServer } from 'vite'

const store = new Map()
globalThis.window = {
  localStorage: {
    getItem: (key) => (store.has(key) ? store.get(key) : null),
    setItem: (key, value) => { store.set(key, value) },
    removeItem: (key) => { store.delete(key) },
  },
}
Object.defineProperty(globalThis, 'navigator', {
  value: { languages: ['en-US'], language: 'en-US' },
  configurable: true,
})
globalThis.document = { documentElement: { lang: '' } }

const server = await createServer({
  root: process.cwd(),
  configFile: 'vite.config.ts',
  server: { middlewareMode: true },
  appType: 'custom',
  logLevel: 'silent',
})
try {
  const mod = await server.ssrLoadModule('/src/shared/i18n/index.ts')
  const initial = mod.getCurrentLanguage()
  const initiallyPersisted = window.localStorage.getItem(mod.LANGUAGE_STORAGE_KEY)
  const initialHtmlLang = document.documentElement.lang
  await mod.changeAppLanguage('zh-CN')
  const changed = mod.getCurrentLanguage()
  const persisted = window.localStorage.getItem(mod.LANGUAGE_STORAGE_KEY)
  const htmlLang = document.documentElement.lang
  const fallback = mod.resolveInitialLanguage({ storedLanguage: undefined, browserLanguages: ['fr-FR'] })

  if (mod.LANGUAGE_STORAGE_KEY !== 'saas-template.locale') throw new Error(`wrong storage key: ${mod.LANGUAGE_STORAGE_KEY}`)
  if (initial !== 'en-US') throw new Error(`browser language should be detected when no stored override exists: ${initial}`)
  if (initiallyPersisted !== null) throw new Error(`browser-detected language should not be persisted: ${initiallyPersisted}`)
  if (initialHtmlLang !== 'en') throw new Error(`initial document lang was not synced from browser detection: ${initialHtmlLang}`)
  if (changed !== 'zh-CN') throw new Error(`changeAppLanguage did not update i18next: ${changed}`)
  if (persisted !== 'zh-CN') throw new Error(`manual language switch was not persisted: ${persisted}`)
  if (htmlLang !== 'zh-CN') throw new Error(`document lang was not synced: ${htmlLang}`)
  if (fallback !== 'zh-CN') throw new Error(`unsupported browser language should fall back to zh-CN: ${fallback}`)
  console.log(JSON.stringify({ key: mod.LANGUAGE_STORAGE_KEY, initial, initiallyPersisted, initialHtmlLang, changed, persisted, htmlLang, fallback }))
} finally {
  await server.close()
}
