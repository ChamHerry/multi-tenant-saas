import { Link } from 'react-router-dom'
import { Building2 } from 'lucide-react'
import { useMyTenants } from '@/features/auth/auth-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { Table, Td, Th } from '@/shared/ui/Table'

export function TenantsPage() {
  const tenants = useMyTenants()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const setCurrentTenantId = useTenantStore((state) => state.setCurrentTenantId)

  if (tenants.isLoading) return <LoadingView label="加载组织列表..." />

  return (
    <div className="space-y-6">
      {tenants.error ? <ErrorView error={tenants.error} title="组织列表加载失败" /> : null}

      <section>
        <Card>
          <CardHeader title="我的组织" description="选择一个组织后，成员、邀请、API Key、审计和配额功能会使用该上下文。" />
          <div className="overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <Th>组织</Th>
                  <Th>角色</Th>
                  <Th>状态</Th>
                  <Th>操作</Th>
                </tr>
              </thead>
              <tbody>
                {tenants.data?.tenants.map((tenant) => (
                  <tr key={tenant.tenant_id}>
                    <Td>
                      <div className="font-bold text-ink">{tenant.tenant_name}</div>
                      <div className="mt-1 text-xs text-subtle">{tenant.tenant_slug}</div>
                    </Td>
                    <Td><Badge tone={tenant.role === 'owner' ? 'green' : tenant.role === 'viewer' ? 'gray' : 'purple'}>{tenant.role}</Badge></Td>
                    <Td><Badge tone={tenant.tenant_status === 'active' ? 'green' : 'orange'}>{tenant.tenant_status}</Badge></Td>
                    <Td>
                      <div className="flex gap-2">
                        <Button size="sm" variant={currentTenantId === tenant.tenant_id ? 'primary' : 'secondary'} onClick={() => setCurrentTenantId(tenant.tenant_id)}>
                          {currentTenantId === tenant.tenant_id ? '当前' : '切换'}
                        </Button>
                        <Link to={`/tenants/${tenant.tenant_id}`}><Button size="sm" variant="ghost" leftIcon={<Building2 className="size-3" />}>详情</Button></Link>
                      </div>
                    </Td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        </Card>
      </section>
    </div>
  )
}
