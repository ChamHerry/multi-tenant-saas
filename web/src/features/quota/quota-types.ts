export type QuotaMetric = 'api_key.count' | 'member.count'

export type QuotaUsage = {
  metric: QuotaMetric
  used: number
  reserved: number
  limit?: number
  remaining?: number
}

export type TenantQuotaView = {
  tenant_id: string
  plan: string
  items: QuotaUsage[]
}
