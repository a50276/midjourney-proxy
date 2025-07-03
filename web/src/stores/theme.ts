import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const isDark = ref(false)
  
  // 初始化主题
  const initTheme = () => {
    const savedTheme = localStorage.getItem('theme')
    if (savedTheme) {
      isDark.value = savedTheme === 'dark'
    } else {
      // 检查系统主题偏好
      isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    updateTheme()
  }
  
  // 切换主题
  const toggleTheme = () => {
    isDark.value = !isDark.value
    updateTheme()
  }
  
  // 更新主题
  const updateTheme = () => {
    const html = document.documentElement
    if (isDark.value) {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  }
  
  // 监听主题变化
  watch(isDark, updateTheme)
  
  return {
    isDark,
    initTheme,
    toggleTheme
  }
})