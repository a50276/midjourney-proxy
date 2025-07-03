import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createI18n } from 'vue-i18n'
import VChart, { THEME_KEY } from 'vue-echarts'

import App from './App.vue'
import router from './router'
import { useThemeStore } from './stores/theme'

// 国际化配置
const i18n = createI18n({
  legacy: false,
  locale: 'zh-CN',
  fallbackLocale: 'en',
  messages: {
    'zh-CN': {
      common: {
        confirm: '确认',
        cancel: '取消',
        save: '保存',
        delete: '删除',
        edit: '编辑',
        add: '添加',
        search: '搜索',
        reset: '重置',
        loading: '加载中...',
        success: '操作成功',
        error: '操作失败',
        warning: '警告',
        info: '提示'
      },
      nav: {
        dashboard: '仪表板',
        tasks: '任务管理',
        accounts: '账号管理',
        users: '用户管理',
        settings: '系统设置',
        stats: '统计信息'
      }
    },
    'en': {
      common: {
        confirm: 'Confirm',
        cancel: 'Cancel',
        save: 'Save',
        delete: 'Delete',
        edit: 'Edit',
        add: 'Add',
        search: 'Search',
        reset: 'Reset',
        loading: 'Loading...',
        success: 'Success',
        error: 'Error',
        warning: 'Warning',
        info: 'Info'
      },
      nav: {
        dashboard: 'Dashboard',
        tasks: 'Tasks',
        accounts: 'Accounts',
        users: 'Users',
        settings: 'Settings',
        stats: 'Statistics'
      }
    }
  }
})

const app = createApp(App)
const pinia = createPinia()

// 注册全局组件
app.component('v-chart', VChart)

// 注册Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(i18n)

// 主题配置
app.provide(THEME_KEY, 'light')

// 初始化主题
const themeStore = useThemeStore()
themeStore.initTheme()

app.mount('#app')