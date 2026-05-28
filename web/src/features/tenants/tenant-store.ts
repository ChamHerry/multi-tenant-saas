import { create } from 'zustand'
import { readStorage, writeStorage } from '@/shared/lib/storage'

export const SELECTED_TENANT_ID_KEY = 'saas-template.selectedTenantId'

export type TenantState = {
  currentTenantId?: string
  setCurrentTenantId: (tenantId?: string) => void
}

export const useTenantStore = create<TenantState>((set) => ({
  currentTenantId: readStorage(SELECTED_TENANT_ID_KEY),
  setCurrentTenantId: (tenantId) => {
    writeStorage(SELECTED_TENANT_ID_KEY, tenantId)
    set({ currentTenantId: tenantId?.trim() || undefined })
  },
}))

export function getCurrentTenantId() {
  return useTenantStore.getState().currentTenantId || readStorage(SELECTED_TENANT_ID_KEY)
}
