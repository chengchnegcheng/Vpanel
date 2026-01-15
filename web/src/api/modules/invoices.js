/**
 * 发票相关 API
 * 处理发票列表、下载等操作
 */
import api from '../base'

export const invoicesApi = {
  /**
   * 获取发票列表
   * @param {Object} params - 查询参数
   * @param {number} params.page - 页码
   * @param {number} params.page_size - 每页数量
   * @returns {Promise<Object>} 发票列表
   */
  list: (params = {}) => api.get('/invoices', { params }),

  /**
   * 获取发票详情
   * @param {number} id - 发票ID
   * @returns {Promise<Object>} 发票详情
   */
  get: (id) => api.get(`/invoices/${id}`),

  /**
   * 下载发票PDF
   * @param {number} id - 发票ID
   * @returns {Promise<Blob>} PDF文件
   */
  download: (id) => api.get(`/invoices/${id}/download`, { responseType: 'blob' }),

  // 管理员 API
  admin: {
    /**
     * 获取所有发票列表（管理员）
     * @param {Object} params - 查询参数
     * @param {number} params.page - 页码
     * @param {number} params.page_size - 每页数量
     * @param {number} params.user_id - 用户ID过滤
     * @param {string} params.start_date - 开始日期
     * @param {string} params.end_date - 结束日期
     * @returns {Promise<Object>} 发票列表
     */
    list: (params = {}) => api.get('/admin/invoices', { params }),

    /**
     * 重新生成发票（管理员）
     * @param {number} id - 发票ID
     * @returns {Promise<Object>} 重新生成的发票
     */
    regenerate: (id) => api.post(`/admin/invoices/${id}/regenerate`)
  }
}

export default invoicesApi
