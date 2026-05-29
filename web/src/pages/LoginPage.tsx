import { FormEvent, useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { LogIn } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { AuthShell } from '@/layouts/AuthShell'
import { accessKeys } from '@/features/access/access-hooks'
import { authKeys, useLoginMutation } from '@/features/auth/auth-hooks'
import { loginPayloadSchema } from '@/features/auth/auth-types'
import { useAuthStore } from '@/features/auth/auth-store'
import { Button, Card, CardHeader, Input, Toast } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'
import { errorMessage } from '@/shared/api/errors'

type LoginLocationState = {
  from?: {
    pathname?: string
    search?: string
  }
}

export function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const loginMutation = useLoginMutation()
  const lastLoginEmail = useAuthStore((state) => state.lastLoginEmail)
  const setLastLoginEmail = useAuthStore((state) => state.setLastLoginEmail)
  const [email, setEmail] = useState(lastLoginEmail ?? '')
  const [password, setPassword] = useState('')
  const [errors, setErrors] = useState<FieldErrors>({})
  const from = useMemo(() => {
    const state = location.state as LoginLocationState | null
    const path = state?.from?.pathname && state.from.pathname !== '/login' ? state.from.pathname : '/'
    return `${path}${state?.from?.search ?? ''}`
  }, [location.state])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(loginPayloadSchema, { email: email.trim(), password })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    setLastLoginEmail(result.data.email)
    await loginMutation.mutateAsync({ email: result.data.email, password: result.data.password })
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: authKeys.me }),
      queryClient.invalidateQueries({ queryKey: authKeys.tenants }),
      queryClient.invalidateQueries({ queryKey: authKeys.session }),
      queryClient.invalidateQueries({ queryKey: accessKeys.snapshot }),
    ])
    navigate(from, { replace: true })
  }

  return (
    <AuthShell>
      <Card className="mx-auto max-w-md bg-white/95 hover:border-line hover:shadow-soft">
        <CardHeader title="登录 SaaS 控制台" description="使用正式账号登录控制台。登录凭证由后端写入 HttpOnly Cookie，前端不会保存访问令牌。" />
        <form className="space-y-4" onSubmit={submit}>
          <Input
            label="邮箱"
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="admin@example.com"
            autoComplete="email"
            autoFocus
            error={errors.email}
          />
          <Input
            label="密码"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder="至少 15 个字符"
            autoComplete="current-password"
            error={errors.password}
          />
          <Button className="w-full" type="submit" isLoading={loginMutation.isPending} leftIcon={<LogIn className="size-4" />}>
            登录控制台
          </Button>
          <Toast tone="red" message={loginMutation.isError ? errorMessage(loginMutation.error) : undefined} />
        </form>
        <div className="mt-5 space-y-3 rounded-panel border border-line bg-surface-soft p-3 text-xs leading-6 text-muted">
          <div>
            <div className="font-bold text-ink">没有账号？</div>
            <Link className="font-bold text-brand hover:text-brand-hover" to="/register" state={location.state}>
              注册新账号
            </Link>
          </div>
          <div>
            <div className="font-bold text-ink">首次本地创建账号</div>
            <code className="mt-2 block overflow-x-auto rounded-control bg-white p-2 text-[11px] text-brand">
              go run . user-password-create --email admin@example.com --password '0123456789abcde' --name Admin --email-verified true
            </code>
          </div>
        </div>
      </Card>
    </AuthShell>
  )
}
