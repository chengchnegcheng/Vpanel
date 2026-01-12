/**
 * SSL 证书管理 API
 * 处理证书的申请、上传、续期、验证等操作
 */
import api from '../base'

export const certificatesApi = {
  /**
   * 获取证书列表
   * @returns {Promise<Array>} 证书列表
   */
  list: () => api.get('/certificates'),

  /**
   * 申请证书
   * @param {Object} data - 申请数据
   * @param {string} data.domain - 域名
   * @returns {Promise<Object>} 申请结果
   */
  apply: (data) => api.post('/certificates/apply', data),

  /**
   * 上传证书
   * @param {Object} data - 证书数据
   * @param {string} data.domain - 域名
   * @param {File} data.certFile - 证书文件
   * @param {File} data.keyFile - 私钥文件
   * @returns {Promise<Object>} 上传结果
   */
  upload: (data) => {
    const formData = new FormData()
    formData.append('domain', data.domain)
    formData.append('cert', data.certFile)
    formData.append('key', data.keyFile)
    return api.post('/certificates/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  /**
   * 续期证书
   * @param {number|string} id - 证书 ID
   * @returns {Promise<Object>} 续期结果
   */
  renew: (id) => api.post(`/certificates/${id}/renew`),

  /**
   * 验证证书
   * @param {number|string} id - 证书 ID
   * @returns {Promise<Object>} 验证结果
   */
  validate: (id) => api.get(`/certificates/${id}/validate`),

  /**
   * 备份证书
   * @param {number|string} id - 证书 ID
   * @returns {Promise<void>} 触发下载
   */
  backup: async (id) => {
    const response = await api.get(`/certificates/${id}/backup`, {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([response]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `certificate_${id}_${Date.now()}.zip`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },

  /**
   * 删除证书
   * @param {number|string} id - 证书 ID
   * @returns {Promise<void>}
   */
  delete: (id) => api.delete(`/certificates/${id}`),

  /**
   * 更新自动续期设置
   * @param {number|string} id - 证书 ID
   * @param {boolean} autoRenew - 是否自动续期
   * @returns {Promise<void>}
   */
  updateAutoRenew: (id, autoRenew) => api.put(`/certificates/${id}/auto-renew`, { autoRenew })
}

export default certificatesApi
