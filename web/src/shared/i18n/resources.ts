export const resources = {
  'zh-CN': {
    translation: {
      app: {
        offline: '当前处于离线状态，连接恢复后会自动同步。',
      },
      language: {
        switcherLabel: '界面语言',
        options: {
          zhCN: '简体中文',
          enUS: 'English',
        },
      },
    },
  },
  'en-US': {
    translation: {
      app: {
        offline: 'You are offline. Changes will sync when the connection is restored.',
      },
      language: {
        switcherLabel: 'Interface language',
        options: {
          zhCN: '简体中文',
          enUS: 'English',
        },
      },
    },
  },
} as const
