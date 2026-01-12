/**
 * 备份管理 API
 * 处理备份的创建、恢复、下载、配置等操作
 */
import api from '../base'

export const backupsApi = {
  /**
   * 获取备份列表
   * @returns {Promise<Array>} 备份列表
   */
  list: () => api.get('/backups'),

  /**
   * 获取单个备份
   * @param {number|string} id - 备份 ID
   * @returns {Promise<Object>} 备份信息
   */
  get: (id) => api.get(`/backups/${id}`),

  /**
   * 创建备份
   * @param {Object} data - 备份配置
   * @returns {Promise<Object>} 创建的备份
   */
  create: (data) => api.post('/backups', data),

  /**
   * 删除备份
   * @param {number|string} id - 备份 ID
   * @returns {Promise<void>}
   */
  delete: (id) => api.delete(`/backups/${id}`),

  /**
   * 下载备份
   * @param {number|string} id - 备份 ID
   * @returns {Promise<Blob>} 备份文件
   */
  download: (id) => api.get(`/backups/${id}/download`, {
    responseType: 'blob'
  }),

  /**
   * 验证备份
   * @param {number|string} id - 备份 ID
   * @returns {Promise<Object>} 验证结果
   */
  verify: (id) => api.post(`/backups/${id}/verify`),

  /**
   * 恢复备份
   * @param {Object} data - 恢复配置
   * @returns {Promise<void>}
   */
  restore: (data) => api.post('/backups/restore', data),

  /**
   * 上传备份文件
   * @param {File} file - 备份文件
   * @returns {Promise<Object>} 上传结果
   */
  upload: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/backups/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  /**
   * 获取备份配置
   * @returns {Promise<Object>} 备份配置
   */
  getConfig: () => api.get('/backups/config'),

  /**
   * 更新备份配置
   * @param {Object} data - 配置数据
   * @returns {Promise<Object>} 更新后的配置
   */
  updateConfig: (data) => api.put('/backups/config', data),

  /**
   * 测试存储连接
   * @param {Object} data - 存储配置
   * @returns {Promise<Object>} 测试结果
   */
  testStorage: (data) => api.post('/backups/test-storage', data),

  /**
   * 获取调度配置
   * @returns {Promise<Object>} 调度配置
   */
  getScheduleConfig: () => api.get('/backups/schedule'),

  /**
   * 更新调度配置
   * @param {Object} data - 调度配置数据
   * @returns {Promise<Object>} 更新后的配置
   */
  updateScheduleConfig: (data) => api.put('/backups/schedule', data),

  /**
   * 获取存储配置
   * @returns {Promise<Object>} 存储配置
   */
  getStorageConfig: () => api.get('/backups/storage'),

  /**
   * 更新存储配置
   * @param {Object} data - 存储配置数据
   * @returns {Promise<Object>} 更新后的配置
   */
  updateStorageConfig: (data) => api.put('/backups/storage', data),

  /**
   * 获取备份进度
   * @param {number|string} id - 备份 ID
   * @returns {Promise<Object>} 进度信息
   */
  getProgress: (id) => api.get(`/backups/${id}/progress`),

  /**
   * 取消备份
   * @param {number|string} id - 备份 ID
   * @returns {Promise<void>}
   */
  cancel: (id) => api.post(`/backups/${id}/cancel`),

  /**
   * 获取备份统计
   * @returns {Promise<Object>} 备份统计数据
   */
  getStats: () => api.get('/backups/stats'),

  /**
   * 清理过期备份
   * @returns {Promise<void>}
   */
  cleanup: () => api.post('/backups/cleanup')
}

export default backupsApi
