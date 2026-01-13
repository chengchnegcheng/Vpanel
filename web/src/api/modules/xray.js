/**
 * Xray 管理 API
 * 处理 Xray 进程管理、配置管理、版本管理等操作
 */
import api from '../base'

export const xrayApi = {
  /**
   * 获取 Xray 状态
   * @returns {Promise<Object>} Xray 状态信息
   */
  getStatus: () => api.get('/xray/status'),

  /**
   * 启动 Xray
   * @returns {Promise<void>}
   */
  start: () => api.post('/xray/start'),

  /**
   * 停止 Xray
   * @returns {Promise<void>}
   */
  stop: () => api.post('/xray/stop'),

  /**
   * 重启 Xray
   * @returns {Promise<void>}
   */
  restart: () => api.post('/xray/restart'),

  /**
   * 获取 Xray 配置
   * @returns {Promise<Object>} Xray 配置
   */
  getConfig: () => api.get('/xray/config'),

  /**
   * 更新 Xray 配置
   * @param {Object} config - Xray 配置
   * @returns {Promise<void>}
   */
  updateConfig: (config) => api.put('/xray/config', config),

  /**
   * 验证 Xray 配置
   * @param {Object} config - Xray 配置
   * @returns {Promise<Object>} 验证结果
   */
  validateConfig: (config) => api.post('/xray/validate', config),

  /**
   * 获取 Xray 版本信息
   * @returns {Promise<Object>} 版本信息
   */
  getVersion: () => api.get('/xray/version'),

  /**
   * 获取可用版本列表
   * @returns {Promise<Object>} 版本列表
   */
  getVersions: () => api.get('/xray/versions'),

  /**
   * 从 GitHub 同步版本列表
   * @returns {Promise<Object>} 同步结果
   */
  syncVersions: () => api.post('/xray/sync-versions', {}, { timeout: 60000 }),

  /**
   * 切换 Xray 版本
   * @param {string} version - 目标版本
   * @returns {Promise<Object>} 切换结果
   */
  switchVersion: (version) => api.post('/xray/switch-version', { version }),

  /**
   * 更新 Xray 到指定版本
   * @param {string} version - 目标版本
   * @returns {Promise<void>}
   */
  update: (version) => api.post('/xray/update', { version })
}

export default xrayApi
