import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/index'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => user.value?.username || '')
  const role = computed(() => user.value?.role || '')
  const userId = computed(() => user.value?.id || null)

  // 方法
  const setToken = (newToken) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUser = (userInfo) => {
    user.value = userInfo
    // 同步角色到 localStorage，供路由守卫使用
    if (userInfo?.role) {
      localStorage.setItem('userRole', userInfo.role)
    }
  }

  const clearAuth = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('userRole')
  }

  // API方法
  const login = async (credentials) => {
    loading.value = true
    error.value = null
    
    try {
      console.log('Login attempt:', credentials.username, '********')
      
      // 使用配置好的 api 实例
      const response = await api.post('/auth/login', credentials)
      
      console.log('Login response:', response)
      
      // 检查响应中是否包含错误信息
      if (response && response.error) {
        console.error('Login error from server:', response.error)
        error.value = response.error
        throw new Error(response.error)
      }
      
      // 检查响应是否包含必要的数据（api 拦截器已经返回 response.data）
      if (!response.token || !response.user) {
        console.error('Invalid login response format:', response)
        error.value = '服务器返回的数据格式不正确'
        throw new Error('服务器返回的数据格式不正确')
      }
      
      const { token: newToken, user: userInfo } = response
      console.log('Login successful:', userInfo)
      
      setToken(newToken)
      setUser(userInfo)
      return true
    } catch (err) {
      console.error('Login error:', err)
      
      // 根据错误类型返回友好的错误消息
      if (err.code === 'UNAUTHORIZED' || err.status === 401) {
        error.value = '用户名或密码错误'
      } else if (err.code === 'NETWORK_ERROR') {
        error.value = '网络连接失败，请检查网络'
      } else if (err.message) {
        error.value = err.message
      } else {
        error.value = '登录失败，请稍后重试'
      }
      
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const logout = () => {
    // 实际项目中可能需要调用登出API
    clearAuth()
  }

  const getUser = async () => {
    if (!token.value) return null
    
    loading.value = true
    error.value = null
    
    try {
      // 使用配置好的 api 实例（已包含认证 token）
      const response = await api.get('/auth/me')
      console.log('Get user response:', response)
      
      if (response && response.error) {
        error.value = response.error
        throw new Error(response.error)
      }
      
      // api 拦截器已经返回 response.data，直接使用 response
      if (!response.id) {
        error.value = '服务器返回的数据格式不正确'
        throw new Error('服务器返回的数据格式不正确')  
      }
      
      setUser(response)
      return user.value
    } catch (err) {
      console.error('Get user error:', err)
      
      if (err.response) {
        console.error('Error response:', err.response.status, err.response.data)
        error.value = err.response.data.error || '获取用户信息失败'
      } else if (err.request) {
        error.value = '网络错误，服务器未响应'
      } else {
        error.value = err.message || '获取用户信息失败'
      }
      
      throw error.value
    } finally {
      loading.value = false
    }
  }

  // 用户管理方法
  const fetchUsers = async (params) => {
    loading.value = true
    error.value = null
    
    try {
      console.log('Fetching users with params:', params)
      const response = await api.get('/users', { params })
      console.log('Users response:', response)
      
      // 适配不同的响应格式（api 拦截器已经返回 response.data）
      let users = [];
      let total = 0;
      
      if (Array.isArray(response)) {
        // 如果响应直接是数组
        users = response;
        total = response.length;
      } else if (response && response.users) {
        // 如果响应中有users字段
        users = response.users;
        total = response.total || users.length;
      } else if (response && response.data && Array.isArray(response.data)) {
        // 如果响应中有data字段
        users = response.data;
        total = response.total || users.length;
      } else {
        console.error('Unknown response format:', response);
        users = [];
        total = 0;
      }
      
      return {
        users,
        total
      }
    } catch (err) {
      console.error('Fetch users error:', err)
      
      if (err.response) {
        console.error('Error response:', err.response.status, err.response.data)
        error.value = err.response.data.error || '获取用户列表失败'
      } else if (err.request) {
        error.value = '网络错误，服务器未响应'
      } else {
        error.value = err.message || '获取用户列表失败'
      }
      
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const createUser = async (userData) => {
    loading.value = true
    error.value = null
    
    try {
      // TODO: 替换为真实 API 端点
      // const response = await axios.post('/api/users', userData)
      const newUser = {
        ...userData,
        id: Date.now(),
        created: new Date().toLocaleString(),
        lastLogin: '-',
        status: true
      }
      
      return newUser
    } catch (err) {
      error.value = err.response?.data?.message || '创建用户失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const updateUser = async (userId, userData) => {
    loading.value = true
    error.value = null
    
    try {
      // TODO: 替换为真实 API 端点
      // const response = await axios.put(`/api/users/${userId}`, userData)
      return { ...userData, id: userId }
    } catch (err) {
      error.value = err.response?.data?.message || '更新用户失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const deleteUser = async (userId) => {
    loading.value = true
    error.value = null
    
    try {
      // TODO: 替换为真实 API 端点
      // await axios.delete(`/api/users/${userId}`)
      return true
    } catch (err) {
      error.value = err.response?.data?.message || '删除用户失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const updateUserStatus = async (userId, status) => {
    loading.value = true
    error.value = null
    
    try {
      // TODO: 替换为真实 API 端点
      // await axios.patch(`/api/users/${userId}/status`, { status })
      return true
    } catch (err) {
      error.value = err.response?.data?.message || '更新用户状态失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }
  
  const updateUserProfile = async (profileData) => {
    loading.value = true
    error.value = null
    
    try {
      // TODO: 替换为真实 API 端点
      // const response = await axios.put('/api/users/profile', profileData)
      
      // 更新本地用户数据
      user.value = {
        ...user.value,
        ...profileData
      }
      
      return user.value
    } catch (err) {
      error.value = err.response?.data?.message || '更新个人资料失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }
  
  const changePassword = async (passwordData) => {
    loading.value = true
    error.value = null
    
    try {
      // TODO: 替换为真实 API 端点
      // await axios.post('/api/users/change-password', passwordData)
      
      return true
    } catch (err) {
      error.value = err.response?.data?.message || '修改密码失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  return {
    // 状态
    token,
    user,
    loading,
    error,
    
    // 计算属性
    isLoggedIn,
    username,
    role,
    userId,
    
    // 方法
    login,
    logout,
    getUser,
    fetchUsers,
    createUser,
    updateUser,
    deleteUser,
    updateUserStatus,
    updateUserProfile,
    changePassword
  }
}) 