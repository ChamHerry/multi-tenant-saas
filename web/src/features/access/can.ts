export function hasAll<T extends string>(owned: readonly T[] | undefined, required: readonly T[] = []) {
  return required.every((permission) => owned?.includes(permission))
}

export function hasAny<T extends string>(owned: readonly T[] | undefined, required: readonly T[] = []) {
  return required.length === 0 || required.some((permission) => owned?.includes(permission))
}
