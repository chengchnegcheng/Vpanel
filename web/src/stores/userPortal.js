/**
 * 用户前台门户 Store
 * 管理用户认证状态和个人信息
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { auth as authApi } from '@/api/modules/portal'

export const useUserPortalStore = defineStore('userPortal', () => {
  // 状态
  const user = ref(null)
  const token = ref(localStorage.getItem('userToken') || null)
  const loading = ref(false)
  const error = ref(null)

  // 计算属性
  const isAuthenticated = computed(() => !!token.value)
  const username = computed(() => user.value?.username || '')
  const email = computed(() => user.value?.email || '')
  const status = computed(() => user.value?.status || 'unknown')
  const trafficUsed = computed(() => user.value?.traffic_used || 0)
  const trafficLimit = computed(() => user.value?.traffic_limit || 0)
  const trafficPercent = computed(() => {
    if (!trafficLimit.value) return 0
    return Math.min(100, Math.round((trafficUsed.value / trafficLimit.value) * 100))
  })
  const expiresAt = computed(() => user.value?.expires_at || null)
  const isExpired = computed(() => {
    if (!expiresAt.value) return false
    return new Date(expiresAt.value) < new Date()
  })
  const daysUntilExpiry = computed(() => {
    if (!expiresAt.value) return null
    const diff = new Date(expiresAt.value) - new Date()
    return Math.ceil(diff / (1000 * 60 * 60 * 24))
  })
  const twoFactorEnabled = computed(() => user.value?.two_factor_enabled || false)

  // 方法
  async function login(credentials) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.login(credentials)
      token.value = response.token
      user.value = response.user
      localStorage.setItem('userToken', response.token)
      localStorage.setItem('userInfo', JSON.stringify(response.user))
      
      // 检查用户角色，如果是管理员，同时设置管理员 token
      if (response.user && response.user.role === 'admin') {
        localStorage.setItem('token', response.token)
        localStorage.setItem('userRole', 'admin')
      } else {
        localStorage.setItem('userRole', 'user')
      }
      
      return response
    } catch (err) {
      error.value = err.message || '登录失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function register(data) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.register(data)
      return response
    } catch (err) {
      error.value = err.message || '注册失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      token.value = null
      user.value = null
      localStorage.removeItem('userToken')
      localStorage.removeItem('userInfo')
    }
  }

  async function fetchProfile() {
    if (!token.value) return
    loading.value = true
    error.value = null
    try {
      const response = await authApi.getProfile()
      user.value = response
      localStorage.setItem('userInfo', JSON.stringify(response))
      return response
    } catch (err) {
      error.value = err.message || '获取用户信息失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(data) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.updateProfile(data)
      user.value = { ...user.value, ...response }
      localStorage.setItem('userInfo', JSON.stringify(user.value))
      return response
    } catch (err) {
      error.value = err.message || '更新资料失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function changePassword(data) {
    loading.value = true
    error.value = null
    try {
      await authApi.changePassword(data)
    } catch (err) {
      error.value = err.message || '修改密码失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function forgotPassword(email) {
    loading.value = true
    error.value = null
    try {
      await authApi.forgotPassword({ email })
    } catch (err) {
      error.value = err.message || '发送重置邮件失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function resetPassword(data) {
    loading.value = true
    error.value = null
    try {
      await authApi.resetPassword(data)
    } catch (err) {
      error.value = err.message || '重置密码失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  // 初始化：从本地存储恢复用户信息
  function init() {
    const savedUser = localStorage.getItem('userInfo')
    if (savedUser) {
      try {
        user.value = JSON.parse(savedUser)
      } catch (e) {
        console.error('Failed to parse saved user info:', e)
      }
    }
  }

  init()

  return {
    // 状态
    user,
    token,
    loading,
    error,
    // 计算属性
    isAuthenticated,
    username,
    email,
    status,
    trafficUsed,
    trafficLimit,
    trafficPercent,
    expiresAt,
    isExpired,
    daysUntilExpiry,
    twoFactorEnabled,
    // 方法
    login,
    register,
    logout,
    fetchProfile,
    updateProfile,
    changePassword,
    forgotPassword,
    resetPassword
  }
})
