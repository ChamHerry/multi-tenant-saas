import { useState, useEffect, useCallback, type ReactNode } from 'react'
import { QueryClient, QueryClientProvider, onlineManager } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { I18nProvider } from '@/shared/i18n'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 15_000,
      refetchOnWindowFocus: false,
      networkMode: 'offlineFirst',
    },
    mutations: {
      networkMode: 'offlineFirst',
    },
  },
})

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <I18nProvider>
        <OnlineStatusHandler />
        {children}
      </I18nProvider>
    </QueryClientProvider>
  )
}

function OnlineStatusHandler() {
  const { t } = useTranslation()
  const [isOnline, setIsOnline] = useState(navigator.onLine)

  const handleOnline = useCallback(() => {
    setIsOnline(true)
    onlineManager.setOnline(true)
    queryClient.invalidateQueries()
  }, [])

  const handleOffline = useCallback(() => {
    setIsOnline(false)
    onlineManager.setOnline(false)
  }, [])

  useEffect(() => {
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [handleOnline, handleOffline])

  if (!isOnline) {
    return (
      <div className="fixed bottom-4 right-4 z-50 rounded-lg bg-yellow-500 px-4 py-2 text-sm font-medium text-white shadow-lg">
        {t('app.offline')}
      </div>
    )
  }

  return null
}
