<template>
  <div class="node-dashboard-page">
    <div class="page-header">
      <h1 class="page-title">节点集群概览</h1>
      <div class="header-actions">
        <el-button @click="refreshData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="$router.push('/admin/nodes')">
          管理节点
        </el-button>
      </div>
    </div>

    <!-- 集群健康状态 -->
    <el-row :gutter="20" class="health-row">
      <el-col :span="6">
        <el-card shadow="never" class="health-card" :class="{ healthy: healthStatus === 'healthy' }">
          <div class="health-content">
            <el-icon class="health-icon" :class="healthStatus">
              <CircleCheck v-if="healthStatus === 'healthy'" />
              <Warning v-else-if="healthStatus === 'warning'" />
              <CircleClose v-else />
            </el-icon>
            <div class="health-info">
              <div class="health-status">{{ getHealthStatusText(healthStatus) }}</div>
              <div class="health-label">集群状态</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ nodeStore.nodeCount }}</div>
            <div class="stat-label">总节点数</div>
          </div>
          <div class="stat-chart">
            <el-progress
              type="dashboard"
              :percentage="nodeStore.healthyPercentage"
              :width="60"
              :stroke-width="6"
              :color="getProgressColor"
            />
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ nodeStore.totalUsers }}</div>
            <div class="stat-label">在线用户</div>
          </div>
          <el-icon class="stat-icon"><User /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ nodeStore.averageLatency }}ms</div>
            <div class="stat-label">平均延迟</div>
          </div>
          <el-icon class="stat-icon"><Timer /></el-icon>
        </el-card>
      </el-col>
    </el-row>

    <!-- 节点状态分布 -->
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <span>节点状态分布</span>
          </template>
          <div class="status-distribution">
            <div class="status-item online">
              <div class="status-count">{{ nodeStore.onlineCount }}</div>
              <div class="status-label">在线</div>
              <div class="status-bar">
                <div class="bar-fill" :style="{ width: getStatusPercentage('online') + '%' }"></div>
              </div>
            </div>
            <div class="status-item offline">
              <div class="status-count">{{ nodeStore.offlineNodes.length }}</div>
              <div class="status-label">离线</div>
              <div class="status-bar">
                <div class="bar-fill" :style="{ width: getStatusPercentage('offline') + '%' }"></div>
              </div>
            </div>
            <div class="status-item unhealthy">
              <div class="status-count">{{ nodeStore.unhealthyNodes.length }}</div>
              <div class="status-label">不健康</div>
              <div class="status-bar">
                <div class="bar-fill" :style="{ width: getStatusPercentage('unhealthy') + '%' }"></div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 节点列表快览 -->
        <el-card shadow="never" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>节点快览</span>
              <el-radio-group v-model="nodeFilter" size="small">
                <el-radio-button label="all">全部</el-radio-button>
                <el-radio-button label="online">在线</el-radio-button>
                <el-radio-button label="issues">有问题</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <el-table :data="filteredNodes" size="small" style="width: 100%" max-height="300">
            <el-table-column prop="name" label="名称" min-width="120">
              <template #default="{ row }">
                <el-link type="primary" @click="$router.push(`/admin/nodes/${row.id}`)">
                  {{ row.name }}
                </el-link>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="region" label="地区" width="80" />
            <el-table-column label="负载" width="100">
              <template #default="{ row }">
                {{ row.current_users }}/{{ row.max_users || '∞' }}
              </template>
            </el-table-column>
            <el-table-column prop="latency" label="延迟" width="80">
              <template #default="{ row }">
                <span :class="getLatencyClass(row.latency)">{{ row.latency }}ms</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <!-- 分组概览 -->
        <el-card shadow="never" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>分组概览</span>
              <el-button link @click="$router.push('/admin/node-groups')">查看全部</el-button>
            </div>
          </template>
          <div class="groups-list">
            <div
              v-for="group in groupStore.groups.slice(0, 5)"
              :key="group.id"
              class="group-item"
              @click="$router.push('/admin/node-groups')"
            >
              <div class="group-info">
                <span class="group-name">{{ group.name }}</span>
                <span class="group-region">{{ group.region }}</span>
              </div>
              <div class="group-stats">
                <span class="nodes-count">
                  <el-icon><Monitor /></el-icon>
                  {{ group.total_nodes || 0 }}
                </span>
                <span class="users-count">
                  <el-icon><User /></el-icon>
                  {{ group.total_users || 0 }}
                </span>
              </div>
            </div>
            <el-empty v-if="groupStore.groups.length === 0" description="暂无分组" :image-size="60" />
          </div>
        </el-card>

        <!-- 告警信息 -->
        <el-card shadow="never" class="chart-card alerts-card">
          <template #header>
            <span>告警信息</span>
          </template>
          <div class="alerts-list">
            <div v-for="alert in alerts" :key="alert.id" class="alert-item" :class="alert.level">
              <el-icon class="alert-icon">
                <Warning v-if="alert.level === 'warning'" />
                <CircleClose v-else />
              </el-icon>
              <div class="alert-content">
                <div class="alert-message">{{ alert.message }}</div>
                <div class="alert-time">{{ formatTime(alert.time) }}</div>
              </div>
            </div>
            <el-empty v-if="alerts.length === 0" description="暂无告警" :image-size="60" />
          </div>
        </el-card>

        <!-- 流量统计 -->
        <el-card shadow="never" class="chart-card">
          <template #header>
            <span>今日流量</span>
          </template>
          <div class="traffic-stats">
            <div class="traffic-item">
              <div class="traffic-value">{{ formatBytes(trafficStats.upload) }}</div>
              <div class="traffic-label">上传</div>
            </div>
            <div class="traffic-item">
              <div class="traffic-value">{{ formatBytes(trafficStats.download) }}</div>
              <div class="traffic-label">下载</div>
            </div>
            <div class="traffic-item total">
              <div class="traffic-value">{{ formatBytes(trafficStats.upload + trafficStats.download) }}</div>
              <div class="traffic-label">总计</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Refresh, CircleCheck, Warning, CircleClose, User, Timer, Monitor
} from '@element-plus/icons-vue'
import { useNodeStore } from '@/stores/node'
import { useNodeGroupStore } from '@/stores/nodeGroup'

const nodeStore = useNodeStore()
const groupStore = useNodeGroupStore()

const loading = ref(false)
const nodeFilter = ref('all')
const trafficStats = ref({ upload: 0, download: 0 })
const alerts = ref([])
let refreshInterval = null

const healthStatus = computed(() => {
  if (nodeStore.nodeCount === 0) return 'unknown'
  if (nodeStore.unhealthyNodes.length > 0) return 'error'
  if (nodeStore.offlineNodes.length > 0) return 'warning'
  return 'healthy'
})

const filteredNodes = computed(() => {
  if (nodeFilter.value === 'online') {
    return nodeStore.onlineNodes
  }
  if (nodeFilter.value === 'issues') {
    return [...nodeStore.offlineNodes, ...nodeStore.unhealthyNodes]
  }
  return nodeStore.nodes
})

const getProgressColor = (percentage) => {
  if (percentage >= 80) return '#67c23a'
  if (percentage >= 50) return '#e6a23c'
  return '#f56c6c'
}

const getHealthStatusText = (status) => {
  const texts = {
    healthy: '运行正常',
    warning: '部分异常',
    error: '存在故障',
    unknown: '未知'
  }
  return texts[status] || '未知'
}

const getStatusType = (status) => {
  const types = { online: 'success', offline: 'info', unhealthy: 'danger' }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = { online: '在线', offline: '离线', unhealthy: '不健康' }
  return texts[status] || status
}

const getStatusPercentage = (status) => {
  if (nodeStore.nodeCount === 0) return 0
  let count = 0
  if (status === 'online') count = nodeStore.onlineCount
  else if (status === 'offline') count = nodeStore.offlineNodes.length
  else if (status === 'unhealthy') count = nodeStore.unhealthyNodes.length
  return Math.round((count / nodeStore.nodeCount) * 100)
}

const getLatencyClass = (latency) => {
  if (!latency) return ''
  if (latency < 100) return 'latency-good'
  if (latency < 300) return 'latency-medium'
  return 'latency-bad'
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(2)} ${units[i]}`
}

const generateAlerts = () => {
  const newAlerts = []
  
  nodeStore.unhealthyNodes.forEach(node => {
    newAlerts.push({
      id: `unhealthy-${node.id}`,
      level: 'error',
      message: `节点 ${node.name} 状态不健康`,
      time: new Date()
    })
  })
  
  nodeStore.offlineNodes.forEach(node => {
    newAlerts.push({
      id: `offline-${node.id}`,
      level: 'warning',
      message: `节点 ${node.name} 已离线`,
      time: new Date()
    })
  })
  
  alerts.value = newAlerts.slice(0, 5)
}

const fetchData = async () => {
  try {
    await Promise.all([
      nodeStore.fetchNodes(),
      groupStore.fetchGroupsWithStats()
    ])
    
    // 获取流量统计
    try {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      const res = await nodeStore.getTotalTraffic({
        start: start.toISOString(),
        end: now.toISOString()
      })
      trafficStats.value = res || { upload: 0, download: 0 }
    } catch (e) {
      console.error('获取流量统计失败:', e)
    }
    
    generateAlerts()
  } catch (e) {
    ElMessage.error(e.message || '获取数据失败')
  }
}

const refreshData = async () => {
  loading.value = true
  await fetchData()
  loading.value = false
}

onMounted(() => {
  fetchData()
  // 每 30 秒自动刷新
  refreshInterval = setInterval(fetchData, 30000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.node-dashboard-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.health-row {
  margin-bottom: 20px;
}

.health-card {
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e7ed 100%);
}

.health-card.healthy {
  background: linear-gradient(135deg, #f0f9eb 0%, #e1f3d8 100%);
}

.health-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.health-icon {
  font-size: 48px;
}

.health-icon.healthy { color: var(--el-color-success); }
.health-icon.warning { color: var(--el-color-warning); }
.health-icon.error { color: var(--el-color-danger); }

.health-status {
  font-size: 20px;
  font-weight: 600;
}

.health-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.stat-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-content .stat-value {
  font-size: 28px;
  font-weight: 600;
}

.stat-content .stat-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.stat-icon {
  font-size: 48px;
  color: var(--el-color-primary-light-5);
  opacity: 0.6;
}

.chart-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-distribution {
  display: flex;
  gap: 40px;
  padding: 20px 0;
}

.status-item {
  flex: 1;
  text-align: center;
}

.status-count {
  font-size: 32px;
  font-weight: 600;
}

.status-item.online .status-count { color: var(--el-color-success); }
.status-item.offline .status-count { color: var(--el-color-info); }
.status-item.unhealthy .status-count { color: var(--el-color-danger); }

.status-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin: 8px 0;
}

.status-bar {
  height: 8px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s;
}

.status-item.online .bar-fill { background: var(--el-color-success); }
.status-item.offline .bar-fill { background: var(--el-color-info); }
.status-item.unhealthy .bar-fill { background: var(--el-color-danger); }

.groups-list {
  max-height: 200px;
  overflow-y: auto;
}

.group-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
}

.group-item:hover {
  background: var(--el-fill-color-light);
  margin: 0 -20px;
  padding: 12px 20px;
}

.group-item:last-child {
  border-bottom: none;
}

.group-name {
  font-weight: 500;
}

.group-region {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-left: 8px;
}

.group-stats {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.group-stats span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.alerts-list {
  max-height: 200px;
  overflow-y: auto;
}

.alert-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.alert-item:last-child {
  border-bottom: none;
}

.alert-icon {
  font-size: 18px;
  margin-top: 2px;
}

.alert-item.warning .alert-icon { color: var(--el-color-warning); }
.alert-item.error .alert-icon { color: var(--el-color-danger); }

.alert-message {
  font-size: 14px;
}

.alert-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.traffic-stats {
  display: flex;
  justify-content: space-around;
  padding: 20px 0;
}

.traffic-item {
  text-align: center;
}

.traffic-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--el-color-primary);
}

.traffic-item.total .traffic-value {
  color: var(--el-color-success);
}

.traffic-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.latency-good { color: var(--el-color-success); }
.latency-medium { color: var(--el-color-warning); }
.latency-bad { color: var(--el-color-danger); }
</style>
