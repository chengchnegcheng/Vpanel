/**
 * 主题管理 Composable
 * 支持浅色、深色和跟随系统三种模式
 */
import { ref, computed, watch, onMounted } from 'vue'

// 主题模式
const THEME_MODES = {
  LIGHT: 'light',
  DARK: 'dark',
  AUTO: 'auto'
}

// 存储键
const STORAGE_KEY = 'userTheme'

// 全局状态
const themeMode = ref(THEME_MODES.AUTO)
const isDark = ref(false)

// 系统偏好
let mediaQuery = null

/**
 * 获取系统主题偏好
 */
function getSystemPreference() {
  if (typeof window === 'undefined') return false
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

/**
 * 应用主题到 DOM
 */
function applyTheme(dark) {
  isDark.value = dark
  
  if (dark) {
    document.documentElement.classList.add('dark')
    document.documentElement.setAttribute('data-theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    document.documentElement.setAttribute('data-theme', 'light')
  }
}

/**
 * 更新主题
 */
function updateTheme() {
  if (themeMode.value === THEME_MODES.AUTO) {
    applyTheme(getSystemPreference())
  } else {
    applyTheme(themeMode.value === THEME_MODES.DARK)
  }
}

/**
 * 监听系统主题变化
 */
function setupSystemThemeListener() {
  if (typeof window === 'undefined') return
  
  mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  
  const handler = (e) => {
    if (themeMode.value === THEME_MODES.AUTO) {
      applyTheme(e.matches)
    }
  }
  
  // 兼容旧版浏览器
  if (mediaQuery.addEventListener) {
    mediaQuery.addEventListener('change', handler)
  } else {
    mediaQuery.addListener(handler)
  }
}

/**
 * 主题管理 Hook
 */
export function useTheme() {
  // 初始化
  onMounted(() => {
    // 从本地存储加载主题设置
    const savedTheme = localStorage.getItem(STORAGE_KEY)
    if (savedTheme && Object.values(THEME_MODES).includes(savedTheme)) {
      themeMode.value = savedTheme
    }
    
    // 应用主题
    updateTheme()
    
    // 监听系统主题变化
    setupSystemThemeListener()
  })
  
  // 监听主题模式变化
  watch(themeMode, (newMode) => {
    localStorage.setItem(STORAGE_KEY, newMode)
    updateTheme()
  })
  
  // 计算属性
  const themeModeText = computed(() => {
    const texts = {
      [THEME_MODES.LIGHT]: '浅色',
      [THEME_MODES.DARK]: '深色',
      [THEME_MODES.AUTO]: '跟随系统'
    }
    return texts[themeMode.value]
  })
  
  // 方法
  function setTheme(mode) {
    if (Object.values(THEME_MODES).includes(mode)) {
      themeMode.value = mode
    }
  }
  
  function toggleTheme() {
    if (themeMode.value === THEME_MODES.LIGHT) {
      themeMode.value = THEME_MODES.DARK
    } else if (themeMode.value === THEME_MODES.DARK) {
      themeMode.value = THEME_MODES.AUTO
    } else {
      themeMode.value = THEME_MODES.LIGHT
    }
  }
  
  function toggleDarkMode() {
    if (isDark.value) {
      themeMode.value = THEME_MODES.LIGHT
    } else {
      themeMode.value = THEME_MODES.DARK
    }
  }
  
  return {
    // 状态
    themeMode,
    isDark,
    themeModeText,
    
    // 常量
    THEME_MODES,
    
    // 方法
    setTheme,
    toggleTheme,
    toggleDarkMode
  }
}

export default useTheme
