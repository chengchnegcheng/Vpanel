<template>
  <div class="download-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">客户端下载</h1>
      <p class="page-subtitle">选择适合您设备的客户端，开始使用服务</p>
    </div>

    <!-- 平台选择 -->
    <div class="platform-tabs">
      <el-radio-group v-model="selectedPlatform" size="large">
        <el-radio-button 
          v-for="platform in platforms" 
          :key="platform.value" 
          :value="platform.value"
        >
          <el-icon><component :is="platform.icon" /></el-icon>
          {{ platform.label }}
        </el-radio-button>
      </el-radio-group>
    </div>

    <!-- 客户端列表 -->
    <div class="clients-section">
      <div class="clients-grid">
        <div 
          v-for="client in filteredClients" 
          :key="client.name"
          class="client-card"
          :class="{ recommended: client.recommended }"
        >
          <!-- 推荐标签 -->
          <div v-if="client.recommended" class="recommended-badge">
            <el-icon><Star /></el-icon>
            推荐
          </div>

          <!-- 客户端信息 -->
          <div class="client-header">
            <div class="client-logo">
              <img v-if="client.logo" :src="client.logo" :alt="client.name" />
              <el-icon v-else><Box /></el-icon>
            </div>
            <div class="client-info">
              <h3 class="client-name">{{ client.name }}</h3>
              <span class="client-version">{{ client.version }}</span>
            </div>
          </div>

          <p class="client-description">{{ client.description }}</p>

          <!-- 特性标签 -->
          <div class="client-features">
            <el-tag 
              v-for="feature in client.features" 
              :key="feature"
              size="small"
              type="info"
            >
              {{ feature }}
            </el-tag>
          </div>

          <!-- 操作按钮 -->
          <div class="client-actions">
            <el-button 
              type="primary" 
              @click="downloadClient(client)"
              :disabled="!client.downloadUrl"
            >
              <el-icon><Download /></el-icon>
              下载
            </el-button>
            <el-button 
              v-if="client.tutorialUrl"
              @click="openTutorial(client)"
            >
              <el-icon><Document /></el-icon>
              教程
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 使用说明 -->
    <el-card class="tips-card" shadow="never">
      <template #header>
        <span>
          <el-icon><InfoFilled /></el-icon>
          使用说明
        </span>
      </template>

      <el-collapse v-model="activeTip">
        <el-collapse-item title="如何选择客户端？" name="1">
          <p>根据您的设备系统选择对应的客户端。推荐使用带有"推荐"标签的客户端，它们通常具有更好的兼容性和用户体验。</p>
        </el-collapse-item>
        <el-collapse-item title="如何导入订阅？" name="2">
          <p>1. 下载并安装客户端</p>
          <p>2. 打开客户端，找到"订阅"或"配置"选项</p>
          <p>3. 添加订阅链接（可在"订阅管理"页面获取）</p>
          <p>4. 更新订阅，选择节点连接</p>
        </el-collapse-item>
        <el-collapse-item title="遇到问题怎么办？" name="3">
          <p>如果在使用过程中遇到问题，您可以：</p>
          <p>1. 查看帮助中心的常见问题</p>
          <p>2. 提交工单获取技术支持</p>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Monitor, Iphone, Apple, Platform,
  Download, Document, Star, Box, InfoFilled
} from '@element-plus/icons-vue'

// 状态
const selectedPlatform = ref('windows')
const activeTip = ref(['1'])

// 平台列表
const platforms = [
  { value: 'windows', label: 'Windows', icon: Monitor },
  { value: 'macos', label: 'macOS', icon: Apple },
  { value: 'linux', label: 'Linux', icon: Monitor },
  { value: 'ios', label: 'iOS', icon: Iphone },
  { value: 'android', label: 'Android', icon: Iphone }
]

// 客户端列表
const clients = [
  // Windows
  {
    name: 'Clash Verge',
    platform: 'windows',
    version: 'v1.4.0',
    description: '基于 Clash Meta 的现代化代理客户端，界面美观，功能强大。',
    features: ['Clash 规则', '自动更新', '系统代理'],
    recommended: true,
    downloadUrl: 'https://github.com/clash-verge-rev/clash-verge-rev/releases',
    tutorialUrl: '#'
  },
  {
    name: 'v2rayN',
    platform: 'windows',
    version: 'v6.0',
    description: '功能全面的 V2Ray 客户端，支持多种协议。',
    features: ['多协议', '路由规则', '订阅管理'],
    recommended: false,
    downloadUrl: 'https://github.com/2dust/v2rayN/releases',
    tutorialUrl: '#'
  },
  {
    name: 'Clash for Windows',
    platform: 'windows',
    version: 'v0.20.39',
    description: '经典的 Clash 客户端，稳定可靠。',
    features: ['Clash 规则', 'TUN 模式', '配置管理'],
    recommended: false,
    downloadUrl: '#',
    tutorialUrl: '#'
  },
  // macOS
  {
    name: 'ClashX Pro',
    platform: 'macos',
    version: 'v1.118.0',
    description: 'macOS 上最受欢迎的 Clash 客户端，支持增强模式。',
    features: ['增强模式', '菜单栏', '自动更新'],
    recommended: true,
    downloadUrl: '#',
    tutorialUrl: '#'
  },
  {
    name: 'Clash Verge',
    platform: 'macos',
    version: 'v1.4.0',
    description: '跨平台的现代化 Clash 客户端。',
    features: ['Clash Meta', '美观界面', '跨平台'],
    recommended: false,
    downloadUrl: 'https://github.com/clash-verge-rev/clash-verge-rev/releases',
    tutorialUrl: '#'
  },
  {
    name: 'Surge',
    platform: 'macos',
    version: 'v5',
    description: '专业级网络调试工具，功能强大但需付费。',
    features: ['专业级', '网络调试', 'MitM'],
    recommended: false,
    downloadUrl: 'https://nssurge.com/',
    tutorialUrl: '#'
  },
  // Linux
  {
    name: 'Clash Verge',
    platform: 'linux',
    version: 'v1.4.0',
    description: '支持 Linux 的现代化 Clash 客户端。',
    features: ['Clash Meta', 'AppImage', 'deb/rpm'],
    recommended: true,
    downloadUrl: 'https://github.com/clash-verge-rev/clash-verge-rev/releases',
    tutorialUrl: '#'
  },
  {
    name: 'Clash Meta',
    platform: 'linux',
    version: 'v1.16.0',
    description: '命令行版本的 Clash，适合服务器使用。',
    features: ['命令行', '轻量级', '服务器'],
    recommended: false,
    downloadUrl: 'https://github.com/MetaCubeX/mihomo/releases',
    tutorialUrl: '#'
  },
  // iOS
  {
    name: 'Shadowrocket',
    platform: 'ios',
    version: 'v2.2.40',
    description: 'iOS 上最流行的代理客户端，功能全面。',
    features: ['多协议', '规则分流', '按需连接'],
    recommended: true,
    downloadUrl: 'https://apps.apple.com/app/shadowrocket/id932747118',
    tutorialUrl: '#'
  },
  {
    name: 'Quantumult X',
    platform: 'ios',
    version: 'v1.4.0',
    description: '功能强大的网络工具，支持复杂规则。',
    features: ['脚本支持', '规则分流', 'MitM'],
    recommended: false,
    downloadUrl: 'https://apps.apple.com/app/quantumult-x/id1443988620',
    tutorialUrl: '#'
  },
  {
    name: 'Surge',
    platform: 'ios',
    version: 'v5',
    description: '专业级网络调试工具。',
    features: ['专业级', '网络调试', 'MitM'],
    recommended: false,
    downloadUrl: 'https://apps.apple.com/app/surge-5/id1442620678',
    tutorialUrl: '#'
  },
  // Android
  {
    name: 'Clash Meta for Android',
    platform: 'android',
    version: 'v2.9.0',
    description: 'Android 上的 Clash Meta 客户端。',
    features: ['Clash Meta', '规则分流', '自动更新'],
    recommended: true,
    downloadUrl: 'https://github.com/MetaCubeX/ClashMetaForAndroid/releases',
    tutorialUrl: '#'
  },
  {
    name: 'v2rayNG',
    platform: 'android',
    version: 'v1.8.0',
    description: 'Android 上的 V2Ray 客户端。',
    features: ['多协议', '轻量级', '订阅管理'],
    recommended: false,
    downloadUrl: 'https://github.com/2dust/v2rayNG/releases',
    tutorialUrl: '#'
  },
  {
    name: 'Surfboard',
    platform: 'android',
    version: 'v2.22.0',
    description: '支持 Surge 配置的 Android 客户端。',
    features: ['Surge 配置', '规则分流', '美观界面'],
    recommended: false,
    downloadUrl: 'https://github.com/getsurfboard/surfboard/releases',
    tutorialUrl: '#'
  }
]

// 计算属性
const filteredClients = computed(() => {
  return clients.filter(c => c.platform === selectedPlatform.value)
})

// 方法
function downloadClient(client) {
  if (client.downloadUrl && client.downloadUrl !== '#') {
    window.open(client.downloadUrl, '_blank')
  } else {
    ElMessage.info('下载链接暂不可用')
  }
}

function openTutorial(client) {
  if (client.tutorialUrl && client.tutorialUrl !== '#') {
    window.open(client.tutorialUrl, '_blank')
  } else {
    ElMessage.info('教程正在编写中')
  }
}
</script>

<style scoped>
.download-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.page-subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

/* 平台选择 */
.platform-tabs {
  margin-bottom: 24px;
}

.platform-tabs :deep(.el-radio-button__inner) {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* 客户端网格 */
.clients-section {
  margin-bottom: 24px;
}

.clients-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.client-card {
  position: relative;
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  transition: all 0.3s;
}

.client-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.client-card.recommended {
  border: 2px solid #409eff;
}

/* 推荐标签 */
.recommended-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: #409eff;
  color: #fff;
  font-size: 12px;
  border-radius: 12px;
}

/* 客户端头部 */
.client-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.client-logo {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  border-radius: 12px;
  font-size: 24px;
  color: #909399;
}

.client-logo img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 12px;
}

.client-name {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 4px 0;
}

.client-version {
  font-size: 13px;
  color: #909399;
}

.client-description {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  margin: 0 0 16px 0;
}

/* 特性标签 */
.client-features {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 20px;
}

/* 操作按钮 */
.client-actions {
  display: flex;
  gap: 12px;
}

.client-actions .el-button {
  flex: 1;
}

/* 提示卡片 */
.tips-card {
  border-radius: 8px;
}

.tips-card :deep(.el-card__header) {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tips-card p {
  margin: 8px 0;
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
}

/* 响应式 */
@media (max-width: 768px) {
  .platform-tabs :deep(.el-radio-group) {
    display: flex;
    flex-wrap: wrap;
  }

  .clients-grid {
    grid-template-columns: 1fr;
  }
}
</style>
