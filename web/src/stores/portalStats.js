/**
 * 用户前台统计 Store
 * 管理流量统计和使用数据
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { stats as statsApi } from '@/api/modules/portal'

export const usePortalStatsStore = defineStore('portalStats', () => {
  // 状态
  const trafficStats = ref([])
  const usageStats = ref([])
  const dashboardStats = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // 时间周期
  const period = ref('day') // day, week, month

  // 计算属性
  const totalUpload = computed(() => 
    trafficStats.value.reduce((sum, item) => sum + (item.upload || 0), 0)
  )

  const totalDownload = computed(() => 
    trafficStats.value.reduce((sum, item) => sum + (item.download || 0), 0)
  )

  const totalTraffic = computed(() => totalUpload.value + totalDownload.value)

  const peakUsage = computed(() => {
    if (!trafficStats.value.length) return null
    return trafficStats.value.reduce((max, item) => {
      const total = (item.upload || 0) + (item.download || 0)
      return total > max.total ? { ...item, total } : max
    }, { total: 0 })
  })

  // 方法
  async function fetchTrafficStats(params = {}) {
    loading.value = true
    error.value = null
    try {
      const response = await statsApi.getTrafficStats({
        period: period.value,
        ...params
      })
      trafficStats.value = response.data || response
      return response
    } catch (err) {
      error.value = err.message || '获取流量统计失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchUsageStats(params = {}) {
    loading.value = true
    error.value = null
    try {
      const response = await statsApi.getUsageStats({
        period: period.value,
        ...params
      })
      usageStats.value = response.data || response
      return response
    } catch (err) {
      error.value = err.message || '获取使用统计失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchDashboardStats() {
    loading.value = true
    error.value = null
    try {
      const response = await statsApi.getDashboardStats()
      dashboardStats.value = response
      return response
    } catch (err) {
      error.value = err.message || '获取仪表板统计失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function exportStats(params = {}) {
    try {
      const blob = await statsApi.exportStats({
        period: period.value,
        format: 'csv',
        ...params
      })
      // 创建下载链接
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `traffic-stats-${new Date().toISOString().split('T')[0]}.csv`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    } catch (err) {
      error.value = err.message || '导出统计数据失败'
      throw err
    }
  }

  function setPeriod(newPeriod) {
    period.value = newPeriod
  }

  function formatBytes(bytes) {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  function getChartData() {
    return {
      labels: trafficStats.value.map(item => item.date || item.time),
      datasets: [
        {
          label: '上传',
          data: trafficStats.value.map(item => item.upload || 0),
          borderColor: '#67c23a',
          backgroundColor: 'rgba(103, 194, 58, 0.1)'
        },
        {
          label: '下载',
          data: trafficStats.value.map(item => item.download || 0),
          borderColor: '#409eff',
          backgroundColor: 'rgba(64, 158, 255, 0.1)'
        }
      ]
    }
  }

  return {
    // 状态
    trafficStats,
    usageStats,
    dashboardStats,
    loading,
    error,
    period,
    // 计算属性
    totalUpload,
    totalDownload,
    totalTraffic,
    peakUsage,
    // 方法
    fetchTrafficStats,
    fetchUsageStats,
    fetchDashboardStats,
    exportStats,
    setPeriod,
    formatBytes,
    getChartData
  }
})
