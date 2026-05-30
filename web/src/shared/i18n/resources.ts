export const resources = {
  'zh-CN': {
    translation: {
      app: {
        offline: '当前处于离线状态，连接恢复后会自动同步。',
      },
      common: {
        close: '关闭',
        loading: '加载中...',
        requestFailed: '请求失败',
      },
      language: {
        switcherLabel: '界面语言',
        options: {
          zhCN: '简体中文',
          enUS: 'English',
        },
      },
      appShell: {
        tenant: {
          selectLabel: '选择组织',
          currentBadge: '当前',
          emptyTitle: '未选择可用组织。',
          emptyDescription: '可先创建或加入组织，再访问组织管理菜单。',
          manageCta: '管理组织',
        },
        user: {
          devUser: '开发用户',
          logout: '退出',
          logoutPending: '退出中…',
        },
        errors: {
          meLoad: '当前用户认证失败',
          accessLoad: '权限上下文加载失败',
        },
      },
      auth: {
        shell: {
          badge: 'SaaS Template Console',
          titlePrefix: '多租户 SaaS 模板的',
          titleHighlight: '安全控制台',
          description:
            '使用正式账号登录后，可以创建组织、切换组织上下文、管理成员和 API Key。浏览器登录态由 HttpOnly Cookie 承载，机器访问继续使用 API Key。',
          features: {
            httpOnly: 'HttpOnly Cookie',
            api: 'GoFrame API',
            orgContext: 'Organization Context',
          },
        },
        guard: {
          checkingSession: '正在检查登录状态...',
          sessionCheckFailed: '登录状态检查失败',
        },
      },
      access: {
        guard: {
          loading: '正在加载权限上下文...',
          loadFailed: '权限上下文加载失败',
          noPlatformPermissionTitle: '无平台权限',
          noPlatformPermissionDescription: '当前账号没有访问该平台管理页面所需的权限。真实 API 仍由后端授权保护。',
          selectTenantTitle: '请先选择组织',
          selectTenantDescription: '该页面需要组织上下文。可以在左侧顶部选择已有组织，或先创建组织。',
          noTenantPermissionTitle: '无组织权限',
          noTenantPermissionDescription: '当前组织角色没有访问该页面所需的权限。真实 API 仍由后端 RBAC 保护。',
        },
      },
      routes: {
        dashboard: {
          title: '组织概览',
          welcome: '欢迎，{{name}}',
        },
        tenants: {
          title: '我的组织',
        },
        tenantSettings: {
          title: '组织设置',
          dynamic: '{{name}}',
        },
        tenantAccess: {
          title: '用户与访问',
        },
        tenantApiKeys: {
          title: '组织 API Keys',
        },
        tenantAudit: {
          title: '组织审计',
        },
        myInvitations: {
          title: '我的邀请',
        },
        securityEvents: {
          title: '安全事件',
        },
        platformTenants: {
          title: '平台组织',
        },
        platformUsers: {
          title: '平台用户与管理员',
        },
        platformAudit: {
          title: '平台审计',
        },
        notFound: {
          title: '页面不存在',
        },
      },
      nav: {
        sections: {
          account: '我的账户',
          tenant: '组织管理',
          platform: '平台管理',
        },
        items: {
          dashboard: '组织概览',
          tenants: '我的组织',
          tenantSettings: '组织设置',
          tenantAccess: '用户与访问',
          tenantApiKeys: '组织 API Keys',
          tenantAudit: '组织审计',
          myInvitations: '我的邀请',
          securityEvents: '安全事件',
          platformTenants: '平台组织',
          platformUsers: '平台用户与管理员',
          platformAudit: '平台审计',
        },
      },
    },
  },
  'en-US': {
    translation: {
      app: {
        offline: 'You are offline. Changes will sync when the connection is restored.',
      },
      common: {
        close: 'Close',
        loading: 'Loading...',
        requestFailed: 'Request failed',
      },
      language: {
        switcherLabel: 'Interface language',
        options: {
          zhCN: '简体中文',
          enUS: 'English',
        },
      },
      appShell: {
        tenant: {
          selectLabel: 'Select organization',
          currentBadge: 'Current',
          emptyTitle: 'No available organization selected.',
          emptyDescription: 'Create or join an organization before opening organization management.',
          manageCta: 'Manage organizations',
        },
        user: {
          devUser: 'Dev User',
          logout: 'Log out',
          logoutPending: 'Logging out…',
        },
        errors: {
          meLoad: 'Current user authentication failed',
          accessLoad: 'Failed to load permission context',
        },
      },
      auth: {
        shell: {
          badge: 'SaaS Template Console',
          titlePrefix: 'Secure console for a',
          titleHighlight: 'multi-tenant SaaS template',
          description:
            'Sign in with a user account to create organizations, switch organization context, manage members, and manage API keys. Browser sessions use HttpOnly cookies while machine access continues to use API keys.',
          features: {
            httpOnly: 'HttpOnly Cookie',
            api: 'GoFrame API',
            orgContext: 'Organization Context',
          },
        },
        guard: {
          checkingSession: 'Checking sign-in status...',
          sessionCheckFailed: 'Failed to check sign-in status',
        },
      },
      access: {
        guard: {
          loading: 'Loading permission context...',
          loadFailed: 'Failed to load permission context',
          noPlatformPermissionTitle: 'No platform permission',
          noPlatformPermissionDescription:
            'This account does not have permission to access this platform admin page. The real API remains protected by backend authorization.',
          selectTenantTitle: 'Select an organization first',
          selectTenantDescription:
            'This page requires organization context. Select an existing organization from the sidebar or create one first.',
          noTenantPermissionTitle: 'No organization permission',
          noTenantPermissionDescription:
            'The current organization role does not have permission to access this page. The real API remains protected by backend RBAC.',
        },
      },
      routes: {
        dashboard: {
          title: 'Organization overview',
          welcome: 'Welcome, {{name}}',
        },
        tenants: {
          title: 'My organizations',
        },
        tenantSettings: {
          title: 'Organization settings',
          dynamic: '{{name}}',
        },
        tenantAccess: {
          title: 'Users and access',
        },
        tenantApiKeys: {
          title: 'Organization API keys',
        },
        tenantAudit: {
          title: 'Organization audit',
        },
        myInvitations: {
          title: 'My invitations',
        },
        securityEvents: {
          title: 'Security events',
        },
        platformTenants: {
          title: 'Platform organizations',
        },
        platformUsers: {
          title: 'Platform users and admins',
        },
        platformAudit: {
          title: 'Platform audit',
        },
        notFound: {
          title: 'Page not found',
        },
      },
      nav: {
        sections: {
          account: 'My account',
          tenant: 'Organization management',
          platform: 'Platform management',
        },
        items: {
          dashboard: 'Organization overview',
          tenants: 'My organizations',
          tenantSettings: 'Organization settings',
          tenantAccess: 'Users and access',
          tenantApiKeys: 'Organization API keys',
          tenantAudit: 'Organization audit',
          myInvitations: 'My invitations',
          securityEvents: 'Security events',
          platformTenants: 'Platform organizations',
          platformUsers: 'Platform users and admins',
          platformAudit: 'Platform audit',
        },
      },
    },
  },
} as const
