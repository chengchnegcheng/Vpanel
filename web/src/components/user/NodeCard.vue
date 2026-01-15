<template>
  <div class="node-card" :class="{ offline: node.status !== 'online' }">
    <!-- 卡片头部 -->
    <div class="card-header">
      <div class="node-info">
        <span class="node-flag">{{ regionFlag }}</span>
        <div class="node-details">
          <h3 class="node-name">{{ node.name }}</h3>
          <span class="node-region">{{ regionLabel }}</span>
        </div>
      </div>
      <el-tag :type="statusType" size="small">
        {{ statusLabel }}
      </el-tag>
    </div>

    <!-- 卡片内容 -->
    <div class="card-body">
      <!-- 协议和端口 -->
      <div class="info-row">
        <span class="info-label">协议</span>
        <el-tag size="small" type="info">{{ node.protocol }}</el-tag>
      </div>

      <!-- 负载 -->
      <div class="info-row">
        <span class="info-label">负载</span>
        <div class="load-bar">
          <el-progress 
            :percentage="node.load" 
            :color="loadColor"
            :stroke-width="6"
            :show-text="false"
          />
          <span class="load-text">{{ node.load }}%</span>
        </div>
      </div>

      <!-- 延迟 -->
      <div class="info-row">
        <span class="info-label">延迟</span>
        <span v-if="testing" class="latency-testing">
          <el-icon class="is-loading"><Loading /></el-icon>
          测速中...
        </span>
        <span v-else-if="latency" :class="latencyClass">
          {{ latency }}ms
        </span>
        <span v-else class="latency-unknown">
          未测试
        </span>
      </div>
    </div>

    <!-- 卡片操作 -->
    <div class="card-footer">
      <el-button 
        size="small" 
        @click="$emit('test')" 
        :loading="testing"
        :disabled="node.status !== 'online'"
      >
        <el-icon><Timer /></el-icon>
        测速
      </el-button>
      <el-button 
        size="small" 
        type="primary" 
        @click="$emit('copy')"
        :disabled="node.status !== 'online'"
      >
        <el-icon><CopyDocument /></el-icon>
        复制配置
      </el-button>
    </div>

    <!-- 维护中遮罩 -->
    <div v-if="node.status === 'maintenance'" class="maintenance-overlay">
      <el-icon><Warning /></el-icon>
      <span>维护中</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Timer, CopyDocument, Loading, Warning } from '@element-plus/icons-vue'

const props = defineProps({
  node: {
    type: Object,
    required: true
  },
  latency: {
    type: Number,
    default: null
  },
  testing: {
    type: Boolean,
    default: false
  }
})

defineEmits(['test', 'copy'])

// 计算属性
const regionFlag = computed(() => {
  const flags = {
    hk: '🇭🇰',
    tw: '🇹🇼',
    jp: '🇯🇵',
    sg: '🇸🇬',
    us: '🇺🇸',
    kr: '🇰🇷',
    de: '🇩🇪',
    uk: '🇬🇧'
  }
  return flags[props.node.region] || '🌐'
})

const regionLabel = computed(() => {
  const labels = {
    hk: '香港',
    tw: '台湾',
    jp: '日本',
    sg: '新加坡',
    us: '美国',
    kr: '韩国',
    de: '德国',
    uk: '英国'
  }
  return labels[props.node.region] || props.node.region
})

const statusType = computed(() => {
  const types = {
    online: 'success',
    offline: 'danger',
    maintenance: 'warning'
  }
  return types[props.node.status] || 'info'
})

const statusLabel = computed(() => {
  const labels = {
    online: '在线',
    offline: '离线',
    maintenance: '维护中'
  }
  return labels[props.node.status] || props.node.status
})

const loadColor = computed(() => {
  const load = props.node.load
  if (load >= 80) return '#f56c6c'
  if (load >= 60) return '#e6a23c'
  return '#67c23a'
})

const latencyClass = computed(() => {
  if (!props.latency) return ''
  if (props.latency < 100) return 'latency-good'
  if (props.latency < 200) return 'latency-fair'
  return 'latency-poor'
})
</script>

<style scoped>
.node-card {
  position: relative;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  overflow: hidden;
  transition: all 0.3s;
}

.node-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.node-card.offline {
  opacity: 0.7;
}

/* 卡片头部 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #ebeef5;
}

.node-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.node-flag {
  font-size: 28px;
}

.node-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 4px 0;
}

.node-region {
  font-size: 13px;
  color: #909399;
}

/* 卡片内容 */
.card-body {
  padding: 16px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  font-size: 14px;
  color: #909399;
}

.load-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  max-width: 150px;
}

.load-bar .el-progress {
  flex: 1;
}

.load-text {
  font-size: 12px;
  color: #606266;
  min-width: 36px;
  text-align: right;
}

.latency-testing {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #409eff;
  font-size: 14px;
}

.latency-good {
  color: #67c23a;
  font-weight: 500;
}

.latency-fair {
  color: #e6a23c;
  font-weight: 500;
}

.latency-poor {
  color: #f56c6c;
  font-weight: 500;
}

.latency-unknown {
  color: #c0c4cc;
  font-size: 14px;
}

/* 卡片操作 */
.card-footer {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  background: #fafafa;
  border-top: 1px solid #ebeef5;
}

.card-footer .el-button {
  flex: 1;
}

/* 维护遮罩 */
.maintenance-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #e6a23c;
  font-size: 16px;
}

.maintenance-overlay .el-icon {
  font-size: 32px;
}
</style>
