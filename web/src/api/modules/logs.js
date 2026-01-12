/**
 * 日志管理 API
 * 处理日志查询、分析、导出、清理等操作
 */
import api from '../base'

export const logsApi = {
  /**
   * 获取日志列表
   * @param {Object} params - 查询参数
   * @param {number} params.page - 页码
   * @param {number} params.pageSize - 每页数量
   * @param {string} params.level - 日志级别
   * @param {string} params.search - 搜索关键词
   * @returns {Promise<Object>} 日志列表
   */
  list: (params) => api.get('/logs', { params }),

  /**
   * 搜索日志
   * @param {Object} params - 搜索参数
   * @returns {Promise<Object>} 搜索结果
   */
  search: (params) => api.get('/logs/search', { params }),

  /**
   * 获取日志分析数据
   * @returns {Promise<Object>} 日志分析数据
   */
  getAnalysis: () => api.get('/logs/analysis'),

  /**
   * 获取日志统计
   * @returns {Promise<Object>} 日志统计数据
   */
  getStats: () => api.get('/logs/stats'),

  /**
   * 获取日志级别统计
   * @returns {Promise<Object>} 级别统计数据
   */
  getLevelStats: () => api.get('/logs/levels'),

  /**
   * 获取日志来源统计
   * @returns {Promise<Object>} 来源统计数据
   */
  getSourceStats: () => api.get('/logs/sources'),

  /**
   * 获取日志时间分布
   * @param {Object} params - 查询参数
   * @returns {Promise<Array>} 时间分布数据
   */
  getTimeDistribution: (params) => api.get('/logs/time-distribution', { params }),

  /**
   * 获取日志轮转配置
   * @returns {Promise<Object>} 轮转配置
   */
  getRotationConfig: () => api.get('/logs/rotation'),

  /**
   * 更新日志轮转配置
   * @param {Object} data - 轮转配置数据
   * @returns {Promise<Object>} 更新后的配置
   */
  updateRotationConfig: (data) => api.put('/logs/rotation', data),

  /**
   * 导出日志
   * @param {Object} params - 导出参数
   * @returns {Promise<Blob>} 日志文件
   */
  export: (params) => api.get('/logs/export', {
    params,
    responseType: 'blob'
  }),

  /**
   * 清理日志
   * @returns {Promise<void>}
   */
  clear: () => api.delete('/logs'),

  /**
   * 批量删除日志
   * @param {Array<number|string>} ids - 日志 ID 列表
   * @returns {Promise<void>}
   */
  batchDelete: (ids) => api.delete('/logs/batch', { data: { ids } })
}

export default logsApi
