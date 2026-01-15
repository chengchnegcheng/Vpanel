/**
 * 支付相关 API
 * 处理支付创建、状态查询等操作
 */
import api from '../base'

export const paymentsApi = {
  /**
   * 创建支付
   * @param {Object} data - 支付数据
   * @param {string} data.order_no - 订单号
   * @param {string} data.method - 支付方式 (alipay, wechat, balance)
   * @returns {Promise<Object>} 支付信息（包含支付URL或二维码）
   */
  create: (data) => api.post('/payments/create', data),

  /**
   * 查询支付状态
   * @param {string} orderNo - 订单号
   * @returns {Promise<Object>} 支付状态
   */
  getStatus: (orderNo) => api.get(`/payments/status/${orderNo}`),

  /**
   * 获取可用支付方式
   * @returns {Promise<Object>} 支付方式列表
   */
  getMethods: () => api.get('/payments/methods'),

  /**
   * 切换支付方式
   * @param {Object} data - 切换数据
   * @param {number} data.order_id - 订单ID
   * @param {string} data.method - 新支付方式
   * @returns {Promise<Object>} 操作结果
   */
  switchMethod: (data) => api.post('/payments/switch-method', data),

  /**
   * 重试支付
   * @param {Object} data - 重试数据
   * @param {number} data.order_id - 订单ID
   * @param {string} data.method - 支付方式（可选）
   * @returns {Promise<Object>} 操作结果
   */
  retry: (data) => api.post('/payments/retry', data),

  /**
   * 获取重试信息
   * @param {number} orderID - 订单ID
   * @returns {Promise<Object>} 重试信息
   */
  getRetryInfo: (orderID) => api.get(`/payments/retry/${orderID}`),

  /**
   * 轮询支付状态（用于前端轮询）
   * @param {string} orderNo - 订单号
   * @param {number} timeout - 超时时间（毫秒）
   * @param {number} interval - 轮询间隔（毫秒）
   * @returns {Promise<Object>} 支付结果
   */
  pollStatus: async (orderNo, timeout = 300000, interval = 3000) => {
    const startTime = Date.now()
    while (Date.now() - startTime < timeout) {
      const result = await paymentsApi.getStatus(orderNo)
      if (result.status === 'paid' || result.status === 'failed') {
        return result
      }
      await new Promise(resolve => setTimeout(resolve, interval))
    }
    throw new Error('支付超时')
  }
}

export default paymentsApi
