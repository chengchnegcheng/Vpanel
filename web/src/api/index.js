/**
 * API 统一入口
 * 按领域组织所有 API 模块，提供统一的导出接口
 */

// 基础配置和工具
import api, { 
  generateErrorId, 
  getErrorMessage, 
  formatApiError 
} from './base'

// API 模块
import { authApi } from './modules/auth'
import { usersApi } from './modules/users'
import { proxiesApi } from './modules/proxies'
import { systemApi } from './modules/system'
import { settingsApi } from './modules/settings'
import { statsApi } from './modules/stats'
import { monitorApi } from './modules/monitor'
import { logsApi } from './modules/logs'
import { backupsApi } from './modules/backups'
import { certificatesApi } from './modules/certificates'
import { rolesApi } from './modules/roles'
import { clientsApi } from './modules/clients'
import { xrayApi } from './modules/xray'
import { eventsApi } from './modules/events'

// 导出所有 API 模块
export {
  // 基础
  api as default,
  generateErrorId,
  getErrorMessage,
  formatApiError,
  
  // 认证
  authApi,
  
  // 用户管理
  usersApi,
  
  // 代理管理
  proxiesApi,
  
  // 系统
  systemApi,
  
  // 设置
  settingsApi,
  
  // 统计
  statsApi,
  
  // 监控
  monitorApi,
  
  // 日志
  logsApi,
  
  // 备份
  backupsApi,
  
  // 证书
  certificatesApi,
  
  // 角色
  rolesApi,
  
  // 客户端
  clientsApi,
  
  // Xray
  xrayApi,
  
  // 事件
  eventsApi
}

// 兼容旧版导出（逐步废弃）
// @deprecated 请使用新的模块化导出
export const auth = authApi
export const users = usersApi
export const proxies = proxiesApi
export const roles = {
  getRoles: rolesApi.list,
  getRole: rolesApi.get,
  createRole: rolesApi.create,
  updateRole: rolesApi.update,
  deleteRole: rolesApi.delete
}
export const events = eventsApi
export const userApi = {
  getInfo: authApi.getProfile,
  updateInfo: authApi.updateProfile,
  changePassword: authApi.changePassword
}
