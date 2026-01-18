/**
 * 余额相关 API
 * 处理余额查询、充值、交易历史等操作
 */
import api from '../base'

export const balanceApi = {
  /**
   * 获取当前用户余额
   * @returns {Promise<Object>} 余额信息
   */
  get: () => api.get('/balance'),

  /**
   * 创建充值订单
   * @param {Object} data - 充值数据
   * @param {number} data.amount - 充值金额（分）
   * @param {string} data.method - 支付方式
   * @returns {Promise<Object>} 充值订单信息
   */
  recharge: (data) => api.post('/balance/recharge', data),

  /**
   * 获取交易历史
   * @param {Object} params - 查询参数
   * @param {number} params.page - 页码
   * @param {number} params.page_size - 每页数量
   * @param {string} params.type - 交易类型过滤
   * @returns {Promise<Object>} 交易历史列表
   */
  getTransactions: (params = {}) => api.get('/balance/transactions', { params }),

  // 管理员 API
  admin: {
    /**
     * 获取用户余额（管理员）
     * @param {number} userId - 用户ID
     * @returns {Promise<Object>} 余额信息
     */
    getUserBalance: (userId) => api.get(`/admin/balance/${userId}`),

    /**
     * 调整用户余额（管理员）
     * @param {number} userId - 用户ID
     * @param {Object} data - 调整数据
     * @param {number} data.amount - 调整金额（正数增加，负数减少）
     * @param {string} data.reason - 调整原因
     * @returns {Promise<Object>} 调整后的余额
     */
    adjust: (userId, data) => api.post(`/admin/balance/${userId}/adjust`, data),

    /**
     * 获取用户交易历史（管理员）
     * @param {number} userId - 用户ID
     * @param {Object} params - 查询参数
     * @returns {Promise<Object>} 交易历史列表
     */
    getTransactions: (userId, params = {}) => api.get(`/admin/balance/${userId}/transactions`, { params }),

    /**
     * 获取余额统计（管理员）
     * @returns {Promise<Object>} 余额统计
     */
    getStats: () => api.get('/balance/stats')
  }
}

export default balanceApi
