import api from './index'

const systemApi = {
  /**
   * 获取系统监控数据
   * @returns {Promise} 返回系统监控数据
   */
  getSystemStatus() {
    return api.get('/system/status')
  },

  /**
   * 获取系统信息
   * @returns {Promise} 返回系统信息
   */
  getSystemInfo() {
    return api.get('/system/info')
  },

  /**
   * 获取CPU使用情况
   * @returns {Promise} 返回CPU使用情况
   */
  getCpuUsage() {
    return api.get('/system/cpu')
  },

  /**
   * 获取内存使用情况
   * @returns {Promise} 返回内存使用情况
   */
  getMemoryUsage() {
    return api.get('/system/memory')
  },

  /**
   * 获取磁盘使用情况
   * @returns {Promise} 返回磁盘使用情况
   */
  getDiskUsage() {
    return api.get('/system/disk')
  },

  /**
   * 获取进程列表
   * @returns {Promise} 返回进程列表
   */
  getProcessList() {
    return api.get('/system/processes')
  }
}

export default systemApi
