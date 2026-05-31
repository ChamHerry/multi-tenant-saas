export type SchemaMessageValue = string | number | boolean | null | undefined
export type SchemaMessageParams = Record<string, SchemaMessageValue>

export type SchemaTranslator = (key: string, fallback: string, params?: SchemaMessageParams) => string

export function formatSchemaMessage(template: string, params: SchemaMessageParams = {}) {
  return template.replace(/\{([A-Za-z0-9_]+)\}/g, (match, name: string) => {
    const value = params[name]
    return value === undefined || value === null ? match : String(value)
  })
}

export const defaultSchemaTranslator: SchemaTranslator = (_key, fallback, params) => formatSchemaMessage(fallback, params)
