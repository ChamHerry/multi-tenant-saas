# 前端开发规范

> 本文档定义了 `web/` 前端项目的架构约定、文件组织、命名规则和编码模式。
> 所有新增代码必须遵循此规范，除非有充分理由并在 PR 中说明偏差。

---

## 1. 技术栈

| 职责 | 选型 | 版本策略 |
|------|------|----------|
| UI 框架 | React 19 + TypeScript (strict) | latest |
| 构建工具 | Vite 6 | latest |
| 样式 | Tailwind CSS v4 (`@tailwindcss/vite`) | latest |
| 路由 | React Router DOM v7 (`createBrowserRouter`) | latest |
| 服务端状态 | TanStack React Query v5 | latest |
| 客户端状态 | Zustand (仅限跨组件 UI 状态) | latest |
| Schema 校验 | Zod (表单校验 + API response validation) | latest |
| 图标 | Lucide React | latest |
| 工具函数 | `clsx` + `tailwind-merge` (`cn`) | latest |

---

## 2. 目录结构

```
src/
├── app/                → 应用入口层
│   ├── App.tsx         → 挂载 Providers + Router
│   ├── providers.tsx   → QueryClientProvider 等全局 Provider
│   ├── router.tsx      → createBrowserRouter 路由表
│   └── RequireAuth.tsx → 鉴权守卫组件
│
├── layouts/            → 页面壳布局
│   ├── AppShell.tsx    → 已登录布局（侧边栏 + 顶栏 + Outlet）
│   ├── AuthShell.tsx   → 未登录布局（居中卡片）
│   ├── menu-config.ts  → 导航菜单项定义
│   ├── scope-nav.ts    → 按权限过滤导航
│   └── page-title.ts   → 页面标题上下文
│
├── pages/              → 页面级组件（与路由一一对应）
│   ├── DashboardPage.tsx
│   ├── LoginPage.tsx
│   └── ...
│
├── features/           → 按业务领域垂直切分（核心）
│   ├── auth/
│   │   ├── auth-types.ts    ← Zod schema + z.infer 类型
│   │   ├── auth-api.ts      ← API 调用函数
│   │   ├── auth-hooks.ts    ← React Query hooks + queryKey 工厂
│   │   └── auth-store.ts    ← Zustand store（可选，按需）
│   ├── tenants/
│   └── ...
│
├── shared/             → 跨 feature 共享
│   ├── api/
│   │   ├── client.ts    ← apiRequest 封装（支持 responseSchema）
│   │   ├── envelope.ts  ← 后端统一响应 envelope 类型
│   │   ├── errors.ts    ← ApiError 类 + errorMessage 工具
│   │   └── request-id.ts
│   ├── lib/
│   │   ├── cn.ts        ← clsx + twMerge
│   │   ├── cookie.ts    ← Cookie 读写
│   │   └── validate.ts  ← validateForm 表单校验工具
│   └── ui/              → 通用 UI 原子组件（通过 index.ts 统一导出）
│       ├── index.ts     ← Barrel export
│       ├── Button.tsx
│       ├── Card.tsx
│       ├── Input.tsx
│       └── ...
│
├── main.tsx            → 应用入口
└── styles.css          → Tailwind @theme 令牌 + 全局样式
```

### 依赖方向规则

```
pages → features → shared
pages → layouts → shared
layouts → features → shared
app → layouts, features, shared
```

- **features 之间不互相依赖**。如果两个 feature 需要共享逻辑，将共享部分提取到 `shared/`。
- **pages 只消费 features 的 hooks 和 shared 的组件**，不直接调用 `*-api.ts`。

---

## 3. Feature 模块约定

每个 feature 目录必须按以下模式组织（四层分离）：

### 3.1 `*-types.ts` — Zod Schema + 类型定义

- **必须使用 Zod schema 定义数据结构**，通过 `z.infer` 推导 TypeScript 类型
- Payload schema 命名规则为 `camelCase + Schema` 后缀（如 `loginPayloadSchema`）
- Response schema 按需定义，用于 API response validation
- 需要跨字段校验（如密码确认）使用 `.refine()`

```ts
// ✅ 正确：Zod schema + z.infer
import { z } from 'zod'

export const loginPayloadSchema = z.object({
  email: z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
  password: z.string().min(1, '请输入密码'),
})
export type LoginPayload = z.infer<typeof loginPayloadSchema>

// ✅ 跨字段校验
export const registerPayloadSchema = z
  .object({
    password: z.string().min(15, '密码至少 15 个字符'),
    confirmPassword: z.string().min(1, '请确认密码'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword'],
  })
export type RegisterPayload = z.infer<typeof registerPayloadSchema>
```

**注意**：当 Zod schema 包含仅用于前端校验的字段（如 `confirmPassword`），API 函数应定义自己的 `Body` 类型（不含校验字段），在 `*-api.ts` 中单独声明。

### 3.2 `*-api.ts` — API 调用

- 每个函数调用 `apiRequest<T>(path, options)`，显式声明返回类型
- 函数名用动词开头：`get` / `create` / `update` / `delete` / `login` / `logout`
- 使用 `skipAuth: true` 标记无需鉴权的接口（如登录、注册）
- 使用 `skipTenant: true` 标记不绑定租户上下文的接口
- 可选传入 `responseSchema` 进行运行时响应校验

```ts
// ✅ 正确
export function login(payload: LoginPayload) {
  return apiRequest<MeResponse>('/api/v1/auth/login', {
    method: 'POST',
    body: payload,
    skipAuth: true,
    skipTenant: true,
  })
}
```

### 3.3 `*-hooks.ts` — React Query Hooks

- **必须**导出一个 `queryKey` 工厂对象，命名规则为 `<feature>Keys`
- Query hooks 使用 `useQuery`，Mutation hooks 使用 `useMutation`
- Query hooks 默认设置 `retry: false`（遵循项目现有约定）
- Mutation hooks 命名规则为 `use<Verb>Mutation`

```ts
// ✅ 正确
export const authKeys = {
  me: ['auth', 'me'] as const,
  session: ['auth', 'session'] as const,
  tenants: ['auth', 'tenants'] as const,
}

export function useMe() {
  return useQuery({ queryKey: authKeys.me, queryFn: getMe, retry: false })
}

export function useLoginMutation() {
  return useMutation({ mutationFn: login })
}
```

### 3.4 `*-store.ts` — Zustand Store（按需）

- **只存客户端 UI 状态**，不缓存服务端数据（那是 React Query 的职责）
- Store 命名规则为 `use<Feature>Store`
- 只在需要跨组件共享 UI 状态时才创建 store

```ts
// ✅ 正确：仅保存 UI 状态
export const useAuthStore = create<AuthState>((set) => ({
  lastLoginEmail: undefined,
  setLastLoginEmail: (email) => set({ lastLoginEmail: email }),
  clearAuthState: () => set({ lastLoginEmail: undefined }),
}))
```

### 3.5 Feature 内的 UI 组件

当 feature 有独立的 UI 片段（如面板、对话框），直接放在 feature 目录下：

```
features/tenant-access/
├── TenantMembersPanel.tsx
└── TenantInvitationsPanel.tsx
```

---

## 4. Pages 约定

- 每个页面一个文件，以 `Page` 后缀命名（如 `LoginPage.tsx`）
- 页面组件只做三件事：组装 hooks + 摆放 shared UI + 处理页面级交互
- 页面间导航使用 `react-router-dom` 的 `useNavigate` / `Link`
- 路由对应的页面标题通过 `handle: { title: '...' }` 定义在 `router.tsx` 中

---

## 5. 路由约定

路由定义集中在 `app/router.tsx`，使用 `createBrowserRouter`：

- 公开路由（`/login`、`/register`）放在顶层，无鉴权
- 需鉴权的路由放在 `RequireAuth` 的 `children` 中
- 需要权限控制的路由，用 `<RequirePermission>` 包裹
- `handle.title` 提供页面标题，AppShell 会自动读取并显示

```tsx
// ✅ 正确
{
  path: 'tenant/api-keys',
  handle: { title: '组织 API Keys' },
  element: (
    <RequirePermission tenant={['tenant:read']}>
      <TenantApiKeysPage />
    </RequirePermission>
  ),
}
```

---

## 6. API 客户端约定

### 6.1 `apiRequest` 使用规则

- 所有 API 调用必须通过 `apiRequest` 或 `apiRaw`
- 泛型参数 `<T>` 声明期望的响应 data 类型
- 不需要手动设置 `Content-Type`（`apiRequest` 自动处理）
- 不需要手动设置 CSRF Token（`apiRequest` 自动从 Cookie 读取）
- 可选传入 `responseSchema` 进行 Zod 运行时校验

### 6.2 API Response Validation

通过 `responseSchema` 选项可以对 API 响应进行 Zod 校验：

```ts
import { meResponseSchema } from './auth-types'

export function getMe() {
  return apiRequest<MeResponse>('/api/v1/me', {
    skipTenant: true,
    responseSchema: meResponseSchema,
  })
}
```

当校验失败时，会抛出 `ApiError`（code 为 `RESPONSE_VALIDATION_ERROR`）。

### 6.3 后端响应 Envelope

后端统一返回：

```json
{ "code": 0, "message": "ok", "data": { ... } }
```

`apiRequest` 会自动解包 `data` 字段，hooks 拿到的就是 `data` 内的内容。

### 6.4 错误处理

- 使用 `errorMessage(error)` 统一提取错误信息（处理 `ApiError` / `Error` / 未知类型）
- Mutation 错误通过 `mutation.isError` 判断，配合 `<Toast>` 组件展示

```tsx
<Toast tone="red" message={loginMutation.isError ? errorMessage(loginMutation.error) : undefined} />
```

---

## 7. 表单校验约定

### 7.1 `validateForm` 工具

所有表单校验通过 `shared/lib/validate.ts` 的 `validateForm` 函数完成：

```ts
import { validateForm, type FieldErrors } from '@/shared/lib/validate'
import { loginPayloadSchema } from '@/features/auth/auth-types'

const [errors, setErrors] = useState<FieldErrors>({})

const submit = (event: FormEvent) => {
  event.preventDefault()
  const result = validateForm(loginPayloadSchema, { email: email.trim(), password })
  if (result.errors) {
    setErrors(result.errors)
    return
  }
  setErrors({})
  mutation.mutate(result.data)
}
```

### 7.2 字段级错误展示

每个 `<Input>` 通过 `error` prop 展示字段级校验错误：

```tsx
<Input
  label="邮箱"
  value={email}
  onChange={(event) => setEmail(event.target.value)}
  error={errors.email}
/>
```

### 7.3 表单校验 Checklist

- [ ] 在 `*-types.ts` 中定义 Zod schema（含中文错误消息）
- [ ] 在页面中用 `validateForm(schema, data)` 替代手写校验
- [ ] 用 `errors.fieldName` 在 `<Input error={...}>` 上展示字段错误
- [ ] 校验通过后用 `result.data` 传递给 mutation（不要重新从 state 拼装）

---

## 8. 共享 UI 组件约定

### 8.1 导入方式

所有 `shared/ui/` 组件通过 barrel `index.ts` 统一导出，使用单一导入路径：

```ts
// ✅ 正确
import { Badge, Button, Card, CardHeader, Input } from '@/shared/ui'

// ❌ 错误：不要直接导入单个文件
import { Button } from '@/shared/ui/Button'
```

### 8.2 组件设计

`shared/ui/` 下的组件遵循"无头 + 样式"模式：

- 通过 `variant` / `tone` / `size` props 控制变体
- 通过 `className` prop 支持外部样式覆盖（必须用 `cn()` 合并）
- 通过 `isLoading` / `leftIcon` 等 props 满足常见交互模式

### 8.3 `cn()` 样式合并

所有组件必须使用 `cn()` 合并 className，确保外部传入的样式能正确覆盖内部默认样式：

```tsx
// ✅ 正确
<button className={cn('base-styles', className)} {...props} />
```

### 8.4 新增 UI 组件 Checklist

- [ ] 放在 `shared/ui/` 目录下
- [ ] 文件名使用 PascalCase（如 `Dialog.tsx`）
- [ ] 使用 `cn()` 合并 className
- [ ] 导出组件自身，不使用 `default export`
- [ ] 在 `shared/ui/index.ts` 中添加 barrel export
- [ ] Props 类型内联定义或导出（如其他地方需要引用）

---

## 9. 样式约定

### 9.1 Design Token

所有颜色、圆角、阴影都定义在 `styles.css` 的 `@theme` 块中。使用 Tailwind 语义化 class：

| 语义 | Token | 用途 |
|------|-------|------|
| `bg-page` | `--color-page` | 页面背景 |
| `text-ink` | `--color-ink` | 主文本 |
| `text-muted` | `--color-muted` | 次要文本 |
| `text-subtle` | `--color-subtle` | 辅助文本 |
| `text-brand` | `--color-brand` | 品牌色文本 |
| `bg-surface` | `--color-surface` | 卡片/面板背景 |
| `bg-surface-soft` | `--color-surface-soft` | 柔和背景 |
| `bg-brand-soft` | `--color-brand-soft` | 品牌色柔和背景 |
| `border-line` | `--color-line` | 分隔线/边框 |
| `rounded-control` | `--radius-control` | 控件圆角 (6px) |
| `rounded-panel` | `--radius-panel` | 面板圆角 (10px) |
| `rounded-card` | `--radius-card` | 卡片圆角 (14px) |
| `shadow-soft` | `--shadow-soft` | 轻阴影 |
| `shadow-brand` | `--shadow-brand` | 品牌阴影 |

**禁止在 Tailwind class 中硬编码颜色值**，如 `bg-blue-500`。必须使用 design token。

### 9.2 圆角层级

```
rounded-control (6px) → 输入框、按钮、Badge
rounded-panel  (10px) → 内嵌面板、菜单、列表项
rounded-card   (14px) → 卡片、对话框、模态框
```

### 9.3 字重层级

```
font-normal (400) → 正文、描述
font-semibold (600) → 导航项、表单标签
font-bold (700) → 卡片标题、强调文本
font-black (900) → 页面标题、品牌文字
```

---

## 10. 命名规则

| 场景 | 规则 | 示例 |
|------|------|------|
| 文件名 | PascalCase | `LoginPage.tsx`, `auth-hooks.ts` |
| 组件 | PascalCase 函数 | `function LoginPage() { ... }` |
| Hooks | camelCase, `use` 前缀 | `useMe()`, `useLoginMutation()` |
| Query Key 工厂 | camelCase + `Keys` 后缀 | `authKeys` |
| Zod Schema | camelCase + `Schema` 后缀 | `loginPayloadSchema`, `userSchema` |
| API 函数 | camelCase 动词开头 | `getMe()`, `login()`, `createTenant()` |
| Types | PascalCase, `z.infer` 推导 | `User`, `LoginPayload`, `MeResponse` |
| Store | `use` 前缀 + PascalCase + `Store` | `useAuthStore`, `useTenantStore` |
| CSS Token | kebab-case | `--color-brand-soft` |
| Tailwind Class | 直接引用 token | `bg-brand-soft`, `text-ink` |
| 路由路径 | kebab-case | `/tenant/api-keys` |

---

## 11. 导入路径

- 使用 `@/` 路径别名指向 `src/`
- 导入顺序：React/第三方库 → `@/features/` → `@/layouts/` → `@/shared/` → `@/app/`
- `shared/ui` 组件从 barrel `index.ts` 统一导入

```ts
// ✅ 正确
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { LogIn } from 'lucide-react'

import { useLoginMutation } from '@/features/auth/auth-hooks'
import { useAuthStore } from '@/features/auth/auth-store'
import { Button, Card, CardHeader, Input } from '@/shared/ui'
import { errorMessage } from '@/shared/api/errors'
```

---

## 12. 状态管理决策树

```
需要存数据吗？
├── 来自服务端 → React Query (useQuery / useMutation)
├── 纯 UI 状态，仅组件内用 → useState / useReducer
└── 纯 UI 状态，跨组件共享 → Zustand store
```

**绝对不要**把服务端数据存进 Zustand store。

---

## 13. 权限控制

- 路由级权限：在 `router.tsx` 中用 `<RequirePermission>` 包裹页面
- 菜单级权限：在 `menu-config.ts` 中通过 `requiredTenantPermissions` / `requiredPlatformPermissions` 过滤
- 功能级权限：使用 `useAccess()` hook 获取当前权限快照

---

## 14. 国际化

当前项目 UI 文案使用**中文**。所有用户可见文本（页面标题、按钮、提示、错误信息）统一用中文，代码注释和变量名用英文。

---

## 15. 禁止事项

| ❌ 禁止 | 原因 |
|---------|------|
| 在 Tailwind class 中硬编码颜色 (`bg-blue-500`) | 绕过 design token，无法统一管理 |
| 在 `*-hooks.ts` 中写业务逻辑 | hooks 只做 React Query 封装 |
| feature 之间互相导入 | 违反依赖方向，造成循环依赖 |
| 使用 `default export` | 统一用 named export，便于 IDE 跳转 |
| 在 Zustand store 中缓存服务端数据 | 职责混乱，导致数据不一致 |
| 直接在 Page 中调用 `apiRequest` | 必须通过 `*-api.ts` → `*-hooks.ts` 分层调用 |
| 使用 `any` 类型 | TypeScript strict 模式，必须提供类型 |
| 创建未在 `@theme` 中定义的新 token | 所有样式 token 必须集中管理 |
| 手写表单校验逻辑 | 必须使用 Zod schema + `validateForm` |
| 直接导入 `@/shared/ui/X` 单文件 | 必须从 `@/shared/ui` barrel 导入 |
| 在 Zod schema 中不含中文错误消息 | 校验消息必须对用户友好 |
