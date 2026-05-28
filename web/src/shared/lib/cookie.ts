export function readCookie(name: string): string | undefined {
  if (typeof document === 'undefined') return undefined
  const prefix = `${encodeURIComponent(name)}=`
  const item = document.cookie
    .split(';')
    .map((part) => part.trim())
    .find((part) => part.startsWith(prefix))
  if (!item) return undefined
  const value = item.slice(prefix.length)
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}
