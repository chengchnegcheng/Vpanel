<template>
  <div class="admin-nodes-page">
    <div class="page-header">
      <h1 class="page-title">节点管理</h1>
      <div class="header-actions">
        <el-button @click="fetchNodes">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          添加节点
        </el-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ nodeStore.nodeCount }}</div>
            <div class="stat-label">总节点数</div>
          </div>
          <el-icon class="stat-icon"><Monitor /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card online">
          <div class="stat-content">
            <div class="stat-value">{{ nodeStore.onlineCount }}</div>
            <div class="stat-label">在线节点</div>
          </div>
          <el-icon class="stat-icon"><CircleCheck /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ nodeStore.totalUsers }}</div>
            <div class="stat-label">总用户数</div>
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

    <!-- 筛选栏 -->
    <el-card shadow="never" class="filter-card">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-input
            v-model="nodeStore.filters.search"
            placeholder="搜索节点名称/地址"
            clearable
            @clear="fetchNodes"
            @keyup.enter="fetchNodes"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="nodeStore.filters.status" placeholder="状态" clearable @change="fetchNodes">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="不健康" value="unhealthy" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="nodeStore.filters.region" placeholder="地区" clearable @change="fetchNodes">
            <el-option v-for="region in regions" :key="region" :label="region" :value="region" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button @click="nodeStore.clearFilters(); fetchNodes()">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 节点列表 -->
    <el-card shadow="never">
      <el-table :data="nodeStore.filteredNodes" v-loading="nodeStore.loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="120">
          <template #default="{ row }">
            <el-link type="primary" @click="viewNodeDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="150">
          <template #default="{ row }">
            {{ row.address }}:{{ row.port }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="region" label="地区" width="100" />
        <el-table-column label="负载" width="120">
          <template #default="{ row }">
            <div class="load-info">
              <span>{{ row.current_users }}/{{ row.max_users || '∞' }}</span>
              <el-progress
                v-if="row.max_users > 0"
                :percentage="Math.round((row.current_users / row.max_users) * 100)"
                :stroke-width="4"
                :show-text="false"
                :status="getLoadStatus(row)"
              />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="latency" label="延迟" width="80">
          <template #default="{ row }">
            <span :class="getLatencyClass(row.latency)">{{ row.latency }}ms</span>
          </template>
        </el-table-column>
        <el-table-column prop="weight" label="权重" width="70" />
        <el-table-column label="同步状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getSyncStatusType(row.sync_status)" size="small">
              {{ getSyncStatusText(row.sync_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="editNode(row)">编辑</el-button>
            <el-button type="warning" link size="small" @click="showTokenDialog(row)">Token</el-button>
            <el-button type="danger" link size="small" @click="deleteNode(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="nodeStore.total > 0" class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          :total="nodeStore.total"
          :page-size="pagination.pageSize"
          layout="total, prev, pager, next"
          @current-change="fetchNodes"
        />
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑节点' : '添加节点'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入节点名称" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="form.address" placeholder="IP 地址或域名" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="地区" prop="region">
          <el-input v-model="form.region" placeholder="如：香港、日本、美国" />
        </el-form-item>
        <el-form-item label="权重" prop="weight">
          <el-input-number v-model="form.weight" :min="1" :max="100" />
          <span class="form-tip">负载均衡权重，数值越大分配用户越多</span>
        </el-form-item>
        <el-form-item label="最大用户数" prop="max_users">
          <el-input-number v-model="form.max_users" :min="0" />
          <span class="form-tip">0 表示无限制</span>
        </el-form-item>
        <el-form-item label="标签">
          <el-tag
            v-for="(tag, index) in form.tags"
            :key="index"
            closable
            @close="removeTag(index)"
            style="margin-right: 8px; margin-bottom: 8px;"
          >
            {{ tag }}
          </el-tag>
          <el-input
            v-if="showTagInput"
            v-model="newTag"
            size="small"
            style="width: 120px;"
            @keyup.enter="addTag"
            @blur="addTag"
          />
          <el-button v-else size="small" @click="showTagInput = true">+ 添加标签</el-button>
        </el-form-item>
        <el-form-item label="IP 白名单">
          <el-input
            v-model="form.ip_whitelist_str"
            type="textarea"
            :rows="3"
            placeholder="每行一个 IP 地址，留空表示不限制"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- Token 管理对话框 -->
    <el-dialog v-model="tokenDialogVisible" title="Token 管理" width="500px">
      <div v-if="currentNode" class="token-dialog-content">
        <div class="token-info">
          <div class="token-label">节点名称</div>
          <div class="token-value">{{ currentNode.name }}</div>
        </div>
        <div class="token-info">
          <div class="token-label">当前 Token</div>
          <div class="token-value token-text">
            <span v-if="showToken">{{ currentToken || '未生成' }}</span>
            <span v-else>{{ currentToken ? '••••••••••••••••' : '未生成' }}</span>
            <el-button v-if="currentToken" link @click="showToken = !showToken">
              <el-icon><View v-if="!showToken" /><Hide v-else /></el-icon>
            </el-button>
            <el-button v-if="currentToken" link @click="copyToken">
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </div>
        </div>
        <div class="token-actions">
          <el-button type="primary" @click="handleGenerateToken" :loading="tokenLoading">
            {{ currentToken ? '重新生成' : '生成 Token' }}
          </el-button>
          <el-button v-if="currentToken" type="warning" @click="handleRotateToken" :loading="tokenLoading">
            轮换 Token
          </el-button>
          <el-button v-if="currentToken" type="danger" @click="handleRevokeToken" :loading="tokenLoading">
            撤销 Token
          </el-button>
        </div>
        <el-alert
          v-if="newToken"
          type="success"
          :closable="false"
          show-icon
          class="new-token-alert"
        >
          <template #title>
            新 Token 已生成，请妥善保存：
          </template>
          <div class="new-token-text">{{ newToken }}</div>
        </el-alert>
      </div>
    </el-dialog>

    <!-- 节点详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="节点详情" width="700px">
      <div v-if="currentNode" class="node-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ currentNode.id }}</el-descriptions-item>
          <el-descriptions-item label="名称">{{ currentNode.name }}</el-descriptions-item>
          <el-descriptions-item label="地址">{{ currentNode.address }}:{{ currentNode.port }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentNode.status)">{{ getStatusText(currentNode.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="地区">{{ currentNode.region || '-' }}</el-descriptions-item>
          <el-descriptions-item label="权重">{{ currentNode.weight }}</el-descriptions-item>
          <el-descriptions-item label="当前用户">{{ currentNode.current_users }}</el-descriptions-item>
          <el-descriptions-item label="最大用户">{{ currentNode.max_users || '无限制' }}</el-descriptions-item>
          <el-descriptions-item label="延迟">{{ currentNode.latency }}ms</el-descriptions-item>
          <el-descriptions-item label="同步状态">
            <el-tag :type="getSyncStatusType(currentNode.sync_status)">{{ getSyncStatusText(currentNode.sync_status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="最后在线">{{ formatTime(currentNode.last_seen_at) }}</el-descriptions-item>
          <el-descriptions-item label="最后同步">{{ formatTime(currentNode.synced_at) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatTime(currentNode.created_at) }}</el-descriptions-item>
        </el-descriptions>
        <div v-if="currentNode.tags && currentNode.tags.length" class="detail-tags">
          <span class="tags-label">标签：</span>
          <el-tag v-for="tag in parseTags(currentNode.tags)" :key="tag" size="small" style="margin-right: 8px;">
            {{ tag }}
          </el-tag>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, Refresh, Search, Monitor, CircleCheck, User, Timer,
  View, Hide, CopyDocument
} from '@element-plus/icons-vue'
import { useNodeStore } from '@/stores/node'

const nodeStore = useNodeStore()

const pagination = reactive({ page: 1, pageSize: 20 })
const dialogVisible = ref(false)
const tokenDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const tokenLoading = ref(false)
const formRef = ref(null)
const showTagInput = ref(false)
const newTag = ref('')
const currentNode = ref(null)
const currentToken = ref('')
const newToken = ref('')
const showToken = ref(false)

const form = reactive({
  id: null,
  name: '',
  address: '',
  port: 8443,
  region: '',
  weight: 1,
  max_users: 0,
  tags: [],
  ip_whitelist_str: ''
})

const rules = {
  name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  address: [{ required: true, message: '请输入节点地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }]
}

const regions = computed(() => {
  const regionSet = new Set(nodeStore.nodes.map(n => n.region).filter(Boolean))
  return Array.from(regionSet)
})

const getStatusType = (status) => {
  const types = { online: 'success', offline: 'info', unhealthy: 'danger' }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = { online: '在线', offline: '离线', unhealthy: '不健康' }
  return texts[status] || status
}

const getSyncStatusType = (status) => {
  const types = { synced: 'success', pending: 'warning', failed: 'danger' }
  return types[status] || 'info'
}

const getSyncStatusText = (status) => {
  const texts = { synced: '已同步', pending: '待同步', failed: '同步失败' }
  return texts[status] || status
}

const getLoadStatus = (node) => {
  if (!node.max_users) return ''
  const ratio = node.current_users / node.max_users
  if (ratio >= 0.9) return 'exception'
  if (ratio >= 0.7) return 'warning'
  return 'success'
}

const getLatencyClass = (latency) => {
  if (latency < 100) return 'latency-good'
  if (latency < 300) return 'latency-medium'
  return 'latency-bad'
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const parseTags = (tags) => {
  if (Array.isArray(tags)) return tags
  if (typeof tags === 'string') {
    try { return JSON.parse(tags) } catch { return [] }
  }
  return []
}

const fetchNodes = async () => {
  try {
    await nodeStore.fetchNodes({
      limit: pagination.pageSize,
      offset: (pagination.page - 1) * pagination.pageSize,
      status: nodeStore.filters.status || undefined,
      region: nodeStore.filters.region || undefined
    })
  } catch (e) {
    ElMessage.error(e.message || '获取节点列表失败')
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, name: '', address: '', port: 8443, region: '',
    weight: 1, max_users: 0, tags: [], ip_whitelist_str: ''
  })
  dialogVisible.value = true
}

const editNode = (node) => {
  isEdit.value = true
  const tags = parseTags(node.tags)
  const ipWhitelist = node.ip_whitelist ? (Array.isArray(node.ip_whitelist) ? node.ip_whitelist : []).join('\n') : ''
  Object.assign(form, {
    id: node.id,
    name: node.name,
    address: node.address,
    port: node.port,
    region: node.region || '',
    weight: node.weight || 1,
    max_users: node.max_users || 0,
    tags: tags,
    ip_whitelist_str: ipWhitelist
  })
  dialogVisible.value = true
}

const submitForm = async () => {
  await formRef.value.validate()
  submitting.value = true
  try {
    const ipWhitelist = form.ip_whitelist_str.split('\n').map(s => s.trim()).filter(Boolean)
    const data = {
      name: form.name,
      address: form.address,
      port: form.port,
      region: form.region,
      weight: form.weight,
      max_users: form.max_users,
      tags: form.tags,
      ip_whitelist: ipWhitelist
    }
    if (isEdit.value) {
      await nodeStore.updateNode(form.id, data)
      ElMessage.success('更新成功')
    } else {
      await nodeStore.createNode(data)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchNodes()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const deleteNode = async (node) => {
  await ElMessageBox.confirm(
    `确定要删除节点 "${node.name}" 吗？该节点上的用户将被重新分配到其他节点。`,
    '删除确认',
    { type: 'warning' }
  )
  try {
    await nodeStore.deleteNode(node.id)
    ElMessage.success('删除成功')
    fetchNodes()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

const viewNodeDetail = async (node) => {
  currentNode.value = node
  detailDialogVisible.value = true
}

const showTokenDialog = (node) => {
  currentNode.value = node
  currentToken.value = ''
  newToken.value = ''
  showToken.value = false
  tokenDialogVisible.value = true
}

const handleGenerateToken = async () => {
  tokenLoading.value = true
  try {
    const res = await nodeStore.generateToken(currentNode.value.id)
    newToken.value = res.token
    currentToken.value = res.token
    ElMessage.success('Token 生成成功')
  } catch (e) {
    ElMessage.error(e.message || '生成 Token 失败')
  } finally {
    tokenLoading.value = false
  }
}

const handleRotateToken = async () => {
  await ElMessageBox.confirm('轮换 Token 后，旧 Token 将立即失效，确定继续？', '确认', { type: 'warning' })
  tokenLoading.value = true
  try {
    const res = await nodeStore.rotateToken(currentNode.value.id)
    newToken.value = res.token
    currentToken.value = res.token
    ElMessage.success('Token 轮换成功')
  } catch (e) {
    ElMessage.error(e.message || '轮换 Token 失败')
  } finally {
    tokenLoading.value = false
  }
}

const handleRevokeToken = async () => {
  await ElMessageBox.confirm('撤销 Token 后，节点将无法连接，确定继续？', '确认', { type: 'warning' })
  tokenLoading.value = true
  try {
    await nodeStore.revokeToken(currentNode.value.id)
    currentToken.value = ''
    newToken.value = ''
    ElMessage.success('Token 已撤销')
  } catch (e) {
    ElMessage.error(e.message || '撤销 Token 失败')
  } finally {
    tokenLoading.value = false
  }
}

const copyToken = async () => {
  try {
    await navigator.clipboard.writeText(currentToken.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const addTag = () => {
  if (newTag.value.trim() && !form.tags.includes(newTag.value.trim())) {
    form.tags.push(newTag.value.trim())
  }
  newTag.value = ''
  showTagInput.value = false
}

const removeTag = (index) => {
  form.tags.splice(index, 1)
}

onMounted(fetchNodes)
</script>

<style scoped>
.admin-nodes-page {
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

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  position: relative;
  overflow: hidden;
}

.stat-card .stat-content {
  position: relative;
  z-index: 1;
}

.stat-card .stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.stat-card .stat-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.stat-card .stat-icon {
  position: absolute;
  right: 20px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 48px;
  color: var(--el-color-primary-light-7);
  opacity: 0.6;
}

.stat-card.online .stat-value {
  color: var(--el-color-success);
}

.stat-card.online .stat-icon {
  color: var(--el-color-success-light-5);
}

.filter-card {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.load-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.latency-good { color: var(--el-color-success); }
.latency-medium { color: var(--el-color-warning); }
.latency-bad { color: var(--el-color-danger); }

.form-tip {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.token-dialog-content {
  padding: 10px 0;
}

.token-info {
  display: flex;
  margin-bottom: 16px;
}

.token-label {
  width: 100px;
  color: var(--el-text-color-secondary);
}

.token-value {
  flex: 1;
}

.token-text {
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: monospace;
}

.token-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.new-token-alert {
  margin-top: 20px;
}

.new-token-text {
  font-family: monospace;
  word-break: break-all;
  margin-top: 8px;
  padding: 8px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}

.node-detail {
  padding: 10px 0;
}

.detail-tags {
  margin-top: 16px;
  display: flex;
  align-items: center;
}

.tags-label {
  margin-right: 12px;
  color: var(--el-text-color-secondary);
}
</style>
