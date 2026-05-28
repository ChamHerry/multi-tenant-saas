export function readStorage(key: string): string | undefined {
  if (typeof window === 'undefined') return undefined
  const value = window.localStorage.getItem(key)
  return value === null || value.trim() === '' ? undefined : value
}

export function writeStorage(key: string, value?: string) {
  if (typeof window === 'undefined') return
  if (value && value.trim() !== '') {
    window.localStorage.setItem(key, value.trim())
    return
  }
  window.localStorage.removeItem(key)
}
