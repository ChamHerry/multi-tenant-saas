import { FormEvent, useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { UserPlus } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { AuthShell } from '@/layouts/AuthShell'
import { accessKeys } from '@/features/access/access-hooks'
import { authKeys, useRegisterMutation } from '@/features/auth/auth-hooks'
import { useAuthStore } from '@/features/auth/auth-store'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { Input } from '@/shared/ui/Input'
import { Toast } from '@/shared/ui/Toast'
import { errorMessage } from '@/shared/api/errors'

type RegisterLocationState = {
  from?: {
    pathname?: string
    search?: string
  }
}

export function RegisterPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const registerMutation = useRegisterMutation()
  const setLastLoginEmail = useAuthStore((state) => state.setLastLoginEmail)
  const [email, setEmail] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [clientError, setClientError] = useState<string>()
  const from = useMemo(() => {
    const state = location.state as RegisterLocationState | null
    const path = state?.from?.pathname && !['/login', '/register', '/'].includes(state.from.pathname) ? state.from.pathname : '/tenants'
    return `${path}${state?.from?.search ?? ''}`
  }, [location.state])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setClientError(undefined)
    const normalizedEmail = email.trim()
    const normalizedDisplayName = displayName.trim()
    if (password !== confirmPassword) {
      setClientError('两次输入的密码不一致')
      return
    }
    await registerMutation.mutateAsync({
      email: normalizedEmail,
      password,
      display_name: normalizedDisplayName || undefined,
    })
    setLastLoginEmail(normalizedEmail)
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
        <CardHeader title="注册 RepoMind" description="创建正式账号后，系统会自动准备一个默认组织；后端会在开始使用时按需完成初始化。" />
        <form className="space-y-4" onSubmit={submit}>
          <Input
            label="邮箱"
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="admin@example.com"
            autoComplete="email"
            autoFocus
            required
          />
          <Input
            label="显示名称"
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
            placeholder="Admin"
            autoComplete="name"
          />
          <Input
            label="密码"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder="至少 15 个字符"
            autoComplete="new-password"
            required
          />
          <Input
            label="确认密码"
            type="password"
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
            placeholder="再次输入密码"
            autoComplete="new-password"
            required
          />
          <Button className="w-full" type="submit" isLoading={registerMutation.isPending} leftIcon={<UserPlus className="size-4" />}>
            注册并进入控制台
          </Button>
          <Toast tone="red" message={clientError ?? (registerMutation.isError ? errorMessage(registerMutation.error) : undefined)} />
        </form>
        <div className="mt-5 rounded-panel border border-line bg-surface-soft p-3 text-xs leading-6 text-muted">
          <div className="font-bold text-ink">已有账号？</div>
          <Link className="font-bold text-brand hover:text-brand-hover" to="/login" state={location.state}>
            返回登录
          </Link>
          <p className="mt-2">如果后端关闭了公开注册，请在配置中开启 <code>auth.password.registrationEnabled</code> 或使用管理员 CLI 创建账号。</p>
        </div>
      </Card>
    </AuthShell>
  )
}
