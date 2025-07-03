<template>
  <el-config-provider :locale="locale">
    <div id="app" :class="{ dark: isDark }">
      <router-view />
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import { useThemeStore } from '@/stores/theme'

const { locale: i18nLocale } = useI18n()
const themeStore = useThemeStore()

const locale = computed(() => {
  return i18nLocale.value === 'zh-CN' ? zhCn : en
})

const isDark = computed(() => themeStore.isDark)
</script>

<style>
html,
body,
#app {
  height: 100%;
  margin: 0;
  padding: 0;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
}

#app {
  background-color: var(--el-bg-color-page);
  transition: background-color 0.3s;
}

.dark {
  color-scheme: dark;
}

/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: var(--el-fill-color-lighter);
}

::-webkit-scrollbar-thumb {
  background: var(--el-fill-color-dark);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--el-fill-color-darker);
}

/* 自定义动画 */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s ease;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(-30px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(30px);
}

/* 卡片样式 */
.card-container {
  background: var(--el-bg-color);
  border-radius: 8px;
  box-shadow: var(--el-box-shadow-light);
  padding: 20px;
  margin-bottom: 20px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .card-container {
    padding: 15px;
    margin-bottom: 15px;
  }
}
</style>