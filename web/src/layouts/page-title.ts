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

export function usePageTitle(title?: PageTitleDescriptor) {
  const context = useContext(PageTitleContext)

  useLayoutEffect(() => {
    if (!context) return
    context.setPageTitle(title)
    return () => context.setPageTitle(undefined)
  }, [context, title])
}
