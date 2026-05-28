import { createContext, useContext, useLayoutEffect } from 'react'

type PageTitleContextValue = {
  setPageTitle: (title?: string) => void
}

export const PageTitleContext = createContext<PageTitleContextValue | null>(null)

export function usePageTitle(title?: string) {
  const context = useContext(PageTitleContext)

  useLayoutEffect(() => {
    if (!context) return
    context.setPageTitle(title)
    return () => context.setPageTitle(undefined)
  }, [context, title])
}
