<template>
  <div class="node-map-page">
    <div class="page-header">
      <h1 class="page-title">节点地理分布</h1>
      <div class="header-actions">
        <el-button @click="fetchNodes" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-row :gutter="20">
      <el-col :span="18">
        <el-card shadow="never" class="map-card">
          <div class="world-map">
            <!-- 简化的世界地图 SVG -->
            <svg viewBox="0 0 1000 500" class="map-svg">
              <!-- 背景 -->
              <rect width="1000" height="500" fill="var(--el-fill-color-light)" />
              
              <!-- 大陆轮廓（简化） -->
              <g class="continents" fill="var(--el-border-color)" stroke="var(--el-border-color-light)" stroke-width="1">
                <!-- 北美 -->
                <path d="M 100 80 Q 150 60 200 80 L 250 120 Q 280 180 250 220 L 180 250 Q 120 220 100 180 Z" />
                <!-- 南美 -->
                <path d="M 200 280 Q 230 260 260 280 L 280 350 Q 260 420 230 400 L 200 350 Z" />
                <!-- 欧洲 -->
                <path d="M 450 80 Q 500 60 550 80 L 580 120 Q 560 160 520 150 L 470 130 Z" />
                <!-- 非洲 -->
                <path d="M 480 180 Q 520 160 560 180 L 580 280 Q 540 350 500 320 L 480 250 Z" />
                <!-- 亚洲 -->
                <path d="M 580 60 Q 700 40 800 80 L 850 150 Q 820 220 750 200 L 650 180 Q 600 140 580 100 Z" />
                <!-- 澳洲 -->
                <path d="M 780 320 Q 830 300 880 320 L 900 380 Q 860 420 820 400 L 780 360 Z" />
              </g>
              
              <!-- 节点标记 -->
              <g class="node-markers">
                <g
                  v-for="node in nodesWithPosition"
                  :key="node.id"
                  :transform="`translate(${node.x}, ${node.y})`"
                  class="node-marker"
                  :class="node.status"
                  @click="selectNode(node)"
                  @mouseenter="hoveredNode = node"
                  @mouseleave="hoveredNode = null"
                >
                  <circle r="8" class="marker-bg" />
                  <circle r="5" class="marker-dot" />
                  <text y="-15" text-anchor="middle" class="marker-label">{{ node.name }}</text>
                </g>
              </g>
            </svg>
            
            <!-- 悬浮提示 -->
            <div
              v-if="hoveredNode"
              class="node-tooltip"
              :style="{ left: hoveredNode.x + 'px', top: hoveredNode.y + 'px' }"
            >
              <div class="tooltip-name">{{ hoveredNode.name }}</div>
              <div class="tooltip-info">
                <span>地区: {{ hoveredNode.region || '-' }}</span>
                <span>状态: {{ getStatusText(hoveredNode.status) }}</span>
                <span>用户: {{ hoveredNode.current_users }}</span>
                <span>延迟: {{ hoveredNode.latency }}ms</span>
              </div>
            </div>
          </div>
          
          <!-- 图例 -->
          <div class="map-legend">
            <div class="legend-item">
              <span class="legend-dot online"></span>
              <span>在线 ({{ nodeStore.onlineCount }})</span>
            </div>
            <div class="legend-item">
              <span class="legend-dot offline"></span>
              <span>离线 ({{ nodeStore.offlineNodes.length }})</span>
            </div>
            <div class="legend-item">
              <span class="legend-dot unhealthy"></span>
              <span>不健康 ({{ nodeStore.unhealthyNodes.length }})</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <!-- 按地区统计 -->
        <el-card shadow="never" class="region-card">
          <template #header>
            <span>按地区统计</span>
          </template>
          <div class="region-list">
            <div
              v-for="region in regionStats"
              :key="region.name"
              class="region-item"
              :class="{ active: selectedRegion === region.name }"
              @click="filterByRegion(region.name)"
            >
              <div class="region-info">
                <span class="region-name">{{ region.name }}</span>
                <span class="region-count">{{ region.count }} 节点</span>
              </div>
              <div class="region-status">
                <span class="status-dot online" v-if="region.online > 0">{{ region.online }}</span>
                <span class="status-dot offline" v-if="region.offline > 0">{{ region.offline }}</span>
                <span class="status-dot unhealthy" v-if="region.unhealthy > 0">{{ region.unhealthy }}</span>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 选中节点详情 -->
        <el-card shadow="never" class="detail-card" v-if="selectedNode">
          <template #header>
            <div class="card-header">
              <span>节点详情</span>
              <el-button link @click="selectedNode = null">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
          </template>
          <el-descriptions :column="1" size="small">
            <el-descriptions-item label="名称">{{ selectedNode.name }}</el-descriptions-item>
            <el-descriptions-item label="地址">{{ selectedNode.address }}</el-descriptions-item>
            <el-descriptions-item label="地区">{{ selectedNode.region || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="getStatusType(selectedNode.status)" size="small">
                {{ getStatusText(selectedNode.status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="用户数">
              {{ selectedNode.current_users }}/{{ selectedNode.max_users || '∞' }}
            </el-descriptions-item>
            <el-descriptions-item label="延迟">
              <span :class="getLatencyClass(selectedNode.latency)">
                {{ selectedNode.latency }}ms
              </span>
            </el-descriptions-item>
          </el-descriptions>
          <el-button
            type="primary"
            size="small"
            style="width: 100%; margin-top: 12px;"
            @click="$router.push(`/admin/nodes/${selectedNode.id}`)"
          >
            查看详情
          </el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Close } from '@element-plus/icons-vue'
import { useNodeStore } from '@/stores/node'

const nodeStore = useNodeStore()

const loading = ref(false)
const hoveredNode = ref(null)
const selectedNode = ref(null)
const selectedRegion = ref(null)

// 地区坐标映射（简化的世界地图坐标）
const regionCoordinates = {
  '香港': { x: 750, y: 180 },
  '日本': { x: 820, y: 140 },
  '新加坡': { x: 720, y: 260 },
  '美国': { x: 180, y: 140 },
  '韩国': { x: 790, y: 130 },
  '台湾': { x: 770, y: 170 },
  '德国': { x: 510, y: 100 },
  '英国': { x: 470, y: 90 },
  '法国': { x: 490, y: 110 },
  '荷兰': { x: 500, y: 95 },
  '俄罗斯': { x: 650, y: 80 },
  '澳大利亚': { x: 830, y: 360 },
  '加拿大': { x: 200, y: 100 },
  '巴西': { x: 280, y: 340 },
  '印度': { x: 680, y: 200 },
  '默认': { x: 500, y: 250 }
}

const nodesWithPosition = computed(() => {
  let nodes = nodeStore.nodes
  
  if (selectedRegion.value) {
    nodes = nodes.filter(n => n.region === selectedRegion.value)
  }
  
  return nodes.map((node, index) => {
    const coords = regionCoordinates[node.region] || regionCoordinates['默认']
    // 同一地区的节点稍微偏移，避免重叠
    const offset = index * 15
    return {
      ...node,
      x: coords.x + (offset % 30) - 15,
      y: coords.y + Math.floor(offset / 30) * 15
    }
  })
})

const regionStats = computed(() => {
  const stats = {}
  
  nodeStore.nodes.forEach(node => {
    const region = node.region || '未知'
    if (!stats[region]) {
      stats[region] = { name: region, count: 0, online: 0, offline: 0, unhealthy: 0 }
    }
    stats[region].count++
    if (node.status === 'online') stats[region].online++
    else if (node.status === 'offline') stats[region].offline++
    else if (node.status === 'unhealthy') stats[region].unhealthy++
  })
  
  return Object.values(stats).sort((a, b) => b.count - a.count)
})

const getStatusType = (status) => {
  const types = { online: 'success', offline: 'info', unhealthy: 'danger' }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = { online: '在线', offline: '离线', unhealthy: '不健康' }
  return texts[status] || status
}

const getLatencyClass = (latency) => {
  if (!latency) return ''
  if (latency < 100) return 'latency-good'
  if (latency < 300) return 'latency-medium'
  return 'latency-bad'
}

const selectNode = (node) => {
  selectedNode.value = node
}

const filterByRegion = (region) => {
  if (selectedRegion.value === region) {
    selectedRegion.value = null
  } else {
    selectedRegion.value = region
  }
}

const fetchNodes = async () => {
  loading.value = true
  try {
    await nodeStore.fetchNodes()
  } catch (e) {
    ElMessage.error(e.message || '获取节点列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchNodes)
</script>

<style scoped>
.node-map-page {
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

.map-card {
  min-height: 500px;
}

.world-map {
  position: relative;
  width: 100%;
  padding-bottom: 50%;
}

.map-svg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.continents {
  opacity: 0.3;
}

.node-marker {
  cursor: pointer;
  transition: transform 0.2s;
}

.node-marker:hover {
  transform: scale(1.2);
}

.marker-bg {
  fill: var(--el-bg-color);
  stroke: currentColor;
  stroke-width: 2;
}

.marker-dot {
  fill: currentColor;
}

.node-marker.online { color: var(--el-color-success); }
.node-marker.offline { color: var(--el-color-info); }
.node-marker.unhealthy { color: var(--el-color-danger); }

.marker-label {
  font-size: 10px;
  fill: var(--el-text-color-primary);
  pointer-events: none;
}

.node-tooltip {
  position: absolute;
  transform: translate(-50%, -100%);
  margin-top: -30px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  padding: 12px;
  box-shadow: var(--el-box-shadow-light);
  z-index: 100;
  min-width: 150px;
}

.tooltip-name {
  font-weight: 600;
  margin-bottom: 8px;
}

.tooltip-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.map-legend {
  display: flex;
  justify-content: center;
  gap: 24px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.legend-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.legend-dot.online { background: var(--el-color-success); }
.legend-dot.offline { background: var(--el-color-info); }
.legend-dot.unhealthy { background: var(--el-color-danger); }

.region-card, .detail-card {
  margin-bottom: 20px;
}

.region-list {
  max-height: 300px;
  overflow-y: auto;
}

.region-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.region-item:hover {
  background: var(--el-fill-color-light);
}

.region-item.active {
  background: var(--el-color-primary-light-9);
}

.region-name {
  font-weight: 500;
}

.region-count {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-left: 8px;
}

.region-status {
  display: flex;
  gap: 8px;
}

.status-dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  border-radius: 10px;
  font-size: 11px;
  color: white;
}

.status-dot.online { background: var(--el-color-success); }
.status-dot.offline { background: var(--el-color-info); }
.status-dot.unhealthy { background: var(--el-color-danger); }

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.latency-good { color: var(--el-color-success); }
.latency-medium { color: var(--el-color-warning); }
.latency-bad { color: var(--el-color-danger); }
</style>
