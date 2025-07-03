<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '240px'" class="sidebar">
      <div class="logo-container">
        <img v-if="!isCollapse" src="/logo.png" alt="Logo" class="logo" />
        <img v-else src="/logo-mini.png" alt="Logo" class="logo-mini" />
      </div>
      
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :unique-opened="true"
        background-color="var(--el-menu-bg-color)"
        text-color="var(--el-menu-text-color)"
        active-text-color="var(--el-color-primary)"
        @select="handleMenuSelect"
      >
        <template v-for="route in menuRoutes" :key="route.path">
          <el-sub-menu v-if="route.children?.length" :index="route.path">
            <template #title>
              <el-icon v-if="route.meta?.icon">
                <component :is="route.meta.icon" />
              </el-icon>
              <span>{{ route.meta?.title }}</span>
            </template>
            <el-menu-item
              v-for="child in route.children"
              :key="child.path"
              :index="route.path + '/' + child.path"
            >
              <el-icon v-if="child.meta?.icon">
                <component :is="child.meta.icon" />
              </el-icon>
              <span>{{ child.meta?.title }}</span>
            </el-menu-item>
          </el-sub-menu>
          
          <el-menu-item v-else :index="route.path">
            <el-icon v-if="route.meta?.icon">
              <component :is="route.meta.icon" />
            </el-icon>
            <span>{{ route.meta?.title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <el-button
            text
            @click="toggleSidebar"
            class="sidebar-toggle"
          >
            <el-icon>
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
          </el-button>
          
          <el-breadcrumb separator="/">
            <el-breadcrumb-item
              v-for="breadcrumb in breadcrumbs"
              :key="breadcrumb.path"
              :to="breadcrumb.path"
            >
              {{ breadcrumb.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <!-- 全屏切换 -->
          <el-tooltip content="全屏切换" placement="bottom">
            <el-button text @click="toggleFullscreen">
              <el-icon>
                <FullScreen />
              </el-icon>
            </el-button>
          </el-tooltip>

          <!-- 主题切换 -->
          <el-tooltip content="主题切换" placement="bottom">
            <el-button text @click="themeStore.toggleTheme()">
              <el-icon>
                <Sunny v-if="themeStore.isDark" />
                <Moon v-else />
              </el-icon>
            </el-button>
          </el-tooltip>

          <!-- 语言切换 -->
          <el-dropdown @command="changeLanguage">
            <el-button text>
              <el-icon>
                <Globe />
              </el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
                <el-dropdown-item command="en">English</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <!-- 用户菜单 -->
          <el-dropdown @command="handleUserMenuCommand">
            <div class="user-info">
              <el-avatar :size="32" :src="userInfo.avatar">
                <el-icon><User /></el-icon>
              </el-avatar>
              <span class="username">{{ userInfo.username }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="main-content">
        <transition name="fade-slide" mode="out-in">
          <router-view />
        </transition>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Fold,
  Expand,
  FullScreen,
  Sunny,
  Moon,
  Globe,
  User
} from '@element-plus/icons-vue'
import { useThemeStore } from '@/stores/theme'

const route = useRoute()
const router = useRouter()
const { locale } = useI18n()
const themeStore = useThemeStore()

// 侧边栏折叠状态
const isCollapse = ref(false)

// 用户信息
const userInfo = ref({
  username: localStorage.getItem('username') || 'Admin',
  avatar: ''
})

// 切换侧边栏
const toggleSidebar = () => {
  isCollapse.value = !isCollapse.value
}

// 全屏切换
const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
  } else {
    document.exitFullscreen()
  }
}

// 语言切换
const changeLanguage = (lang: string) => {
  locale.value = lang
  localStorage.setItem('language', lang)
  ElMessage.success('语言切换成功')
}

// 菜单路由
const menuRoutes = computed(() => {
  return router.getRoutes().filter(route => {
    return !route.meta?.hideInMenu && route.children?.length
  })
})

// 当前激活菜单
const activeMenu = computed(() => {
  return route.path
})

// 面包屑导航
const breadcrumbs = computed(() => {
  const matched = route.matched.filter(item => item.meta?.title)
  return matched.map(item => ({
    title: item.meta?.title,
    path: item.path
  }))
})

// 菜单选择处理
const handleMenuSelect = (index: string) => {
  if (index !== route.path) {
    router.push(index)
  }
}

// 用户菜单处理
const handleUserMenuCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      ElMessage.info('个人信息功能开发中')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        
        // 清除本地存储
        localStorage.removeItem('token')
        localStorage.removeItem('username')
        localStorage.removeItem('userRole')
        
        // 跳转到登录页
        router.push('/login')
        ElMessage.success('退出登录成功')
      } catch {
        // 用户取消操作
      }
      break
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background: var(--el-menu-bg-color);
  border-right: 1px solid var(--el-border-color-light);
  transition: width 0.3s;
  overflow: hidden;
}

.logo-container {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--el-border-color-light);
  padding: 0 20px;
}

.logo {
  height: 32px;
  width: auto;
}

.logo-mini {
  height: 32px;
  width: 32px;
}

.header {
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 60px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.sidebar-toggle {
  font-size: 18px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 5px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: var(--el-fill-color-light);
}

.username {
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.main-content {
  background: var(--el-bg-color-page);
  padding: 20px;
  overflow-y: auto;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header {
    padding: 0 15px;
  }
  
  .main-content {
    padding: 15px;
  }
  
  .header-left .el-breadcrumb {
    display: none;
  }
}
</style>