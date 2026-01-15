/**
 * 用户前台节点 API 模块
 */
import api from '@/api/base'

const BASE_URL = '/portal/nodes'

/**
 * 获取节点列表
 * @param {Object} [params] - 查询参数
 * @param {string} [params.region] - 地区筛选
 * @param {string} [params.protocol] - 协议筛选
 * @param {string} [params.sort] - 排序字段 (name, region, latency)
 * @param {string} [params.order] - 排序方向 (asc, desc)
 * @returns {Promise}
 */
export function getNodes(params = {}) {
  return api.get(BASE_URL, { params })
}

/**
 * 获取节点详情
 * @param {number} id - 节点 ID
 * @returns {Promise}
 */
export function getNode(id) {
  return api.get(`${BASE_URL}/${id}`)
}

/**
 * 测试节点延迟
 * @param {number} id - 节点 ID
 * @returns {Promise}
 */
export function testLatency(id) {
  return api.post(`${BASE_URL}/${id}/ping`)
}

/**
 * 批量测试节点延迟
 * @param {number[]} ids - 节点 ID 列表
 * @returns {Promise}
 */
export function batchTestLatency(ids) {
  return api.post(`${BASE_URL}/ping`, { ids })
}

/**
 * 获取节点地区列表
 * @returns {Promise}
 */
export function getRegions() {
  return api.get(`${BASE_URL}/regions`)
}

/**
 * 获取节点协议列表
 * @returns {Promise}
 */
export function getProtocols() {
  return api.get(`${BASE_URL}/protocols`)
}

export default {
  getNodes,
  getNode,
  testLatency,
  batchTestLatency,
  getRegions,
  getProtocols
}
