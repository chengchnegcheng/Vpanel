/**
 * 监控相关 API
 * 处理系统监控、网络监控、告警设置等操作
 */
import api from '../base'

export const monitorApi = {
  /**
   * 获取系统监控数据
   * @returns {Promise<Object>} 系统监控数据
   */
  getSystemStats: () => api.get('/monitor/system'),

  /**
   * 获取网络监控数据
   * @returns {Promise<Object>} 网络监控数据
   */
  getNetworkStats: () => api.get('/monitor/network'),

  /**
   * 获取进程监控数据
   * @returns {Promise<Object>} 进程监控数据
   */
  getProcessStats: () => api.get('/monitor/process'),

  /**
   * 获取告警设置
   * @returns {Promise<Object>} 告警设置
   */
  getAlertSettings: () => api.get('/monitor/alerts/settings'),

  /**
   * 更新告警设置
   * @param {Object} data - 告警设置数据
   * @returns {Promise<Object>} 更新后的告警设置
   */
  updateAlertSettings: (data) => api.put('/monitor/alerts/settings', data),

  /**
   * 获取告警历史
   * @returns {Promise<Array>} 告警历史列表
   */
  getAlertHistory: () => api.get('/monitor/alerts/history'),

  /**
   * 测试邮件通知
   * @param {Object} data - 邮件配置
   * @returns {Promise<void>}
   */
  testEmailNotification: (data) => api.post('/monitor/alerts/test-email', data),

  /**
   * 测试 Webhook 通知
   * @param {Object} data - Webhook 配置
   * @returns {Promise<void>}
   */
  testWebhookNotification: (data) => api.post('/monitor/alerts/test-webhook', data),

  /**
   * 清除连接
   * @returns {Promise<void>}
   */
  clearConnections: () => api.post('/monitor/connections/clear'),

  /**
   * 终止连接
   * @param {string} id - 连接 ID
   * @returns {Promise<void>}
   */
  killConnection: (id) => api.post(`/monitor/connections/${id}/kill`),

  /**
   * 获取流量历史
   * @param {Object} params - 查询参数
   * @returns {Promise<Array>} 流量历史数据
   */
  getTrafficHistory: (params) => api.get('/monitor/traffic/history', { params }),

  /**
   * 获取协议分布
   * @returns {Promise<Object>} 协议分布数据
   */
  getProtocolDistribution: () => api.get('/monitor/traffic/protocols'),

  /**
   * 获取连接列表
   * @returns {Promise<Array>} 连接列表
   */
  getConnectionList: () => api.get('/monitor/connections')
}

export default monitorApi
