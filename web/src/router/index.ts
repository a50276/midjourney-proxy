import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Layout from '@/layout/index.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: {
      title: '登录',
      hideInMenu: true
    }
  },
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: {
          title: '仪表板',
          icon: 'Monitor'
        }
      }
    ]
  },
  {
    path: '/tasks',
    component: Layout,
    redirect: '/tasks/list',
    meta: {
      title: '任务管理',
      icon: 'Document'
    },
    children: [
      {
        path: 'list',
        name: 'TaskList',
        component: () => import('@/views/tasks/List.vue'),
        meta: {
          title: '任务列表',
          icon: 'List'
        }
      },
      {
        path: 'test',
        name: 'TaskTest',
        component: () => import('@/views/tasks/Test.vue'),
        meta: {
          title: '绘图测试',
          icon: 'Picture'
        }
      }
    ]
  },
  {
    path: '/accounts',
    component: Layout,
    redirect: '/accounts/list',
    meta: {
      title: '账号管理',
      icon: 'User'
    },
    children: [
      {
        path: 'list',
        name: 'AccountList',
        component: () => import('@/views/accounts/List.vue'),
        meta: {
          title: '账号列表',
          icon: 'Avatar'
        }
      }
    ]
  },
  {
    path: '/users',
    component: Layout,
    redirect: '/users/list',
    meta: {
      title: '用户管理',
      icon: 'UserFilled',
      roles: ['admin']
    },
    children: [
      {
        path: 'list',
        name: 'UserList',
        component: () => import('@/views/users/List.vue'),
        meta: {
          title: '用户列表',
          icon: 'UserFilled'
        }
      }
    ]
  },
  {
    path: '/settings',
    component: Layout,
    redirect: '/settings/system',
    meta: {
      title: '系统设置',
      icon: 'Setting',
      roles: ['admin']
    },
    children: [
      {
        path: 'system',
        name: 'SystemSettings',
        component: () => import('@/views/settings/System.vue'),
        meta: {
          title: '系统配置',
          icon: 'Tools'
        }
      }
    ]
  },
  {
    path: '/stats',
    component: Layout,
    redirect: '/stats/overview',
    meta: {
      title: '统计监控',
      icon: 'DataAnalysis'
    },
    children: [
      {
        path: 'overview',
        name: 'StatsOverview',
        component: () => import('@/views/stats/Overview.vue'),
        meta: {
          title: '数据概览',
          icon: 'PieChart'
        }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/404.vue'),
    meta: {
      title: '页面不存在',
      hideInMenu: true
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    return { top: 0 }
  }
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta?.title) {
    document.title = `${to.meta.title} - Midjourney Proxy`
  }
  
  // 检查是否需要登录
  const token = localStorage.getItem('token')
  
  if (to.path === '/login') {
    if (token) {
      next('/')
    } else {
      next()
    }
    return
  }
  
  if (!token) {
    next('/login')
    return
  }
  
  // 检查权限
  const userRole = localStorage.getItem('userRole') || 'user'
  const requiredRoles = to.meta?.roles as string[] | undefined
  
  if (requiredRoles && !requiredRoles.includes(userRole)) {
    ElMessage.error('权限不足')
    next('/')
    return
  }
  
  next()
})

export default router