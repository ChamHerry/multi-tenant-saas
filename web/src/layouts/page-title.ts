import { createContext, useContext, useLayoutEffect } from 'react'
import type { TFunction } from 'i18next'

export type PageTitleDescriptor = {
  key: string
  values?: Record<string, unknown>
}

type PageTitleContextValue = {
  setPageTitle: (title?: PageTitleDescriptor) => void
}

export const PageTitleContext = createContext<PageTitleContextValue | null>(null)

export function pageTitle(key: string, values?: PageTitleDescriptor['values']): PageTitleDescriptor {
  return values ? { key, values } : { key }
}

export function resolvePageTitle(t: TFunction, descriptor: PageTitleDescriptor) {
  return t(descriptor.key, descriptor.values)
}

function stringifyTitleValues(values?: PageTitleDescriptor['values']) {
  return JSON.stringify(values ?? null)
}

function parseTitleValues(signature: string): PageTitleDescriptor['values'] {
  return signature === 'null' ? undefined : JSON.parse(signature)
}

export function usePageTitle(title?: PageTitleDescriptor) {
  const context = useContext(PageTitleContext)
  const titleKey = title?.key
  const titleValuesSignature = stringifyTitleValues(title?.values)

  useLayoutEffect(() => {
    if (!context) return
    context.setPageTitle(titleKey ? pageTitle(titleKey, parseTitleValues(titleValuesSignature)) : undefined)
    return () => context.setPageTitle(undefined)
  }, [context, titleKey, titleValuesSignature])
}
