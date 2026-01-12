/**
 * 客户端管理 API
 * 处理客户端的增删改查、配置生成、流量统计等操作
 */
import api from '../base'

export const clientsApi = {
  /**
   * 获取客户端列表
   * @param {Object} params - 查询参数
   * @returns {Promise<Object>} 客户端列表
   */
  list: (params) => api.get('/clients', { params }),

  /**
   * 获取单个客户端
   * @param {number|string} id - 客户端 ID
   * @returns {Promise<Object>} 客户端信息
   */
  get: (id) => api.get(`/clients/${id}`),

  /**
   * 创建客户端
   * @param {Object} data - 客户端数据
   * @returns {Promise<Object>} 创建的客户端
   */
  create: (data) => api.post('/clients', data),

  /**
   * 更新客户端
   * @param {number|string} id - 客户端 ID
   * @param {Object} data - 客户端数据
   * @returns {Promise<Object>} 更新后的客户端
   */
  update: (id, data) => api.put(`/clients/${id}`, data),

  /**
   * 删除客户端
   * @param {number|string} id - 客户端 ID
   * @returns {Promise<void>}
   */
  delete: (id) => api.delete(`/clients/${id}`),

  /**
   * 启用客户端
   * @param {number|string} id - 客户端 ID
   * @returns {Promise<void>}
   */
  enable: (id) => api.post(`/clients/${id}/enable`),

  /**
   * 禁用客户端
   * @param {number|string} id - 客户端 ID
   * @returns {Promise<void>}
   */
  disable: (id) => api.post(`/clients/${id}/disable`),

  /**
   * 重置客户端流量
   * @param {number|string} id - 客户端 ID
   * @returns {Promise<void>}
   */
  resetTraffic: (id) => api.post(`/clients/${id}/reset-traffic`),

  /**
   * 生成客户端配置文件
   * @param {number|string} id - 客户端 ID
   * @param {string} type - 配置类型
   * @returns {Promise<void>} 触发下载
   */
  generateConfig: async (id, type) => {
    const response = await api.get(`/clients/${id}/config/${type}`, {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([response]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `client_config_${id}_${type}.json`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },

  /**
   * 获取客户端流量统计
   * @param {number|string} id - 客户端 ID
   * @param {Object} params - 查询参数
   * @returns {Promise<Object>} 流量统计数据
   */
  getTrafficStats: (id, params) => api.get(`/clients/${id}/traffic`, { params }),

  /**
   * 获取客户端连接记录
   * @param {number|string} id - 客户端 ID
   * @param {Object} params - 查询参数
   * @returns {Promise<Array>} 连接记录列表
   */
  getConnections: (id, params) => api.get(`/clients/${id}/connections`, { params })
}

export default clientsApi
