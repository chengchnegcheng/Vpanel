/**
 * 事件管理 API
 * 处理系统事件的查询和清理
 */
import api from '../base'

export const eventsApi = {
  /**
   * 获取事件列表
   * @param {Object} params - 查询参数
   * @param {number} params.page - 页码
   * @param {number} params.pageSize - 每页数量
   * @param {string} params.type - 事件类型
   * @returns {Promise<Object>} 事件列表
   */
  list: (params) => api.get('/events', { params }),

  /**
   * 获取单个事件
   * @param {number|string} id - 事件 ID
   * @returns {Promise<Object>} 事件详情
   */
  get: (id) => api.get(`/events/${id}`),

  /**
   * 清除所有事件
   * @returns {Promise<void>}
   */
  clear: () => api.delete('/events')
}

export default eventsApi
