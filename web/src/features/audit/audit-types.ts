export type AuditLog = {
  id: string
  tenant_id?: string
  user_id?: string
  action: string
  resource_type: string
  resource_id?: string
  ip?: string
  user_agent?: string
  metadata: Record<string, unknown>
  created_at: string
}

export type AuditLogListResponse = {
  logs: AuditLog[]
  total: number
}

export type AuditExportJob = {
  id: string
  status: string
  format?: string
  content_type?: string
  file_name?: string
  content?: string
}

export type AuditExportResponse = {
  job: AuditExportJob
}
