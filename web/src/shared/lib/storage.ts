export function readStorage(key: string): string | undefined {
  if (typeof window === 'undefined') return undefined
  try {
    const value = window.localStorage.getItem(key)
    return value === null || value.trim() === '' ? undefined : value
  } catch {
    return undefined
  }
}

export function writeStorage(key: string, value?: string) {
  if (typeof window === 'undefined') return
  try {
    if (value && value.trim() !== '') {
      window.localStorage.setItem(key, value.trim())
      return
    }
    window.localStorage.removeItem(key)
  } catch {
    // Ignore storage failures (private mode/quota/browser policy) and keep in-memory state authoritative.
  }
}
