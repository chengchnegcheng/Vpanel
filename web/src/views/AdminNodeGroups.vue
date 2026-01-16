<template>
  <div class="admin-node-groups-page">
    <div class="page-header">
      <h1 class="page-title">节点分组管理</h1>
      <div class="header-actions">
        <el-button @click="fetchGroups">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          创建分组
        </el-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ groupStore.groupCount }}</div>
            <div class="stat-label">总分组数</div>
          </div>
          <el-icon class="stat-icon"><Folder /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ groupStore.totalNodesInGroups }}</div>
            <div class="stat-label">总节点数</div>
          </div>
          <el-icon class="stat-icon"><Monitor /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card online">
          <div class="stat-content">
            <div class="stat-value">{{ groupStore.totalHealthyNodes }}</div>
            <div class="stat-label">健康节点</div>
          </div>
          <el-icon class="stat-icon"><CircleCheck /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ groupStore.totalUsersInGroups }}</div>
            <div class="stat-label">总用户数</div>
          </div>
          <el-icon class="stat-icon"><User /></el-icon>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选栏 -->
    <el-card shadow="never" class="filter-card">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-input
            v-model="groupStore.filters.search"
            placeholder="搜索分组名称"
            clearable
            @clear="fetchGroups"
            @keyup.enter="fetchGroups"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="groupStore.filters.region" placeholder="地区" clearable @change="fetchGroups">
            <el-option v-for="region in groupStore.regions" :key="region" :label="region" :value="region" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button @click="groupStore.clearFilters(); fetchGroups()">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 分组列表 -->
    <el-row :gutter="20">
      <el-col :span="8" v-for="group in groupStore.filteredGroups" :key="group.id">
        <el-card shadow="hover" class="group-card" @click="viewGroupDetail(group)">
          <template #header>
            <div class="group-header">
              <span class="group-name">{{ group.name }}</span>
              <el-dropdown @click.stop trigger="click">
                <el-button link>
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="editGroup(group)">
                      <el-icon><Edit /></el-icon>
                      编辑
                    </el-dropdown-item>
                    <el-dropdown-item @click="manageNodes(group)">
                      <el-icon><Setting /></el-icon>
                      管理节点
                    </el-dropdown-item>
                    <el-dropdown-item divided @click="deleteGroup(group)">
                      <el-icon><Delete /></el-icon>
                      删除
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
          <div class="group-content">
            <div class="group-info">
              <el-icon><Location /></el-icon>
              <span>{{ group.region || '未设置地区' }}</span>
            </div>
            <div class="group-info">
              <el-icon><Connection /></el-icon>
              <span>{{ getStrategyText(group.strategy) }}</span>
            </div>
            <div class="group-description" v-if="group.description">
              {{ group.description }}
            </div>
            <el-divider />
            <div class="group-stats">
              <div class="stat-item">
                <span class="stat-num">{{ group.total_nodes || group.node_count || 0 }}</span>
                <span class="stat-text">节点</span>
              </div>
              <div class="stat-item">
                <span class="stat-num healthy">{{ group.healthy_nodes || 0 }}</span>
                <span class="stat-text">健康</span>
              </div>
              <div class="stat-item">
                <span class="stat-num">{{ group.total_users || group.user_count || 0 }}</span>
                <span class="stat-text">用户</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!groupStore.loading && groupStore.filteredGroups.length === 0" description="暂无分组" />

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑分组' : '创建分组'" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="分组名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入分组描述" />
        </el-form-item>
        <el-form-item label="地区" prop="region">
          <el-select v-model="form.region" filterable allow-create placeholder="选择或输入地区">
            <el-option label="香港" value="香港" />
            <el-option label="日本" value="日本" />
            <el-option label="新加坡" value="新加坡" />
            <el-option label="美国" value="美国" />
            <el-option label="韩国" value="韩国" />
            <el-option label="台湾" value="台湾" />
            <el-option label="德国" value="德国" />
            <el-option label="英国" value="英国" />
          </el-select>
        </el-form-item>
        <el-form-item label="负载均衡策略" prop="strategy">
          <el-select v-model="form.strategy" placeholder="选择策略">
            <el-option label="轮询" value="round-robin" />
            <el-option label="最少连接" value="least-connections" />
            <el-option label="加权" value="weighted" />
            <el-option label="地理位置" value="geographic" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 管理节点对话框 -->
    <el-dialog v-model="nodesDialogVisible" title="管理分组节点" width="700px">
      <div v-if="currentGroup" class="nodes-dialog-content">
        <div class="dialog-header">
          <span>分组：{{ currentGroup.name }}</span>
          <el-input
            v-model="nodeSearch"
            placeholder="搜索节点"
            style="width: 200px;"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        
        <el-transfer
          v-model="selectedNodeIds"
          :data="transferData"
          :titles="['可用节点', '已选节点']"
          :filter-method="filterNodes"
          filterable
          filter-placeholder="搜索节点"
        />
        
        <div class="dialog-footer">
          <el-button @click="nodesDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="savingNodes" @click="saveGroupNodes">保存</el-button>
        </div>
      </div>
    </el-dialog>

    <!-- 分组详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="分组详情" width="600px">
      <div v-if="currentGroup" class="group-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ currentGroup.id }}</el-descriptions-item>
          <el-descriptions-item label="名称">{{ currentGroup.name }}</el-descriptions-item>
          <el-descriptions-item label="地区">{{ currentGroup.region || '-' }}</el-descriptions-item>
          <el-descriptions-item label="策略">{{ getStrategyText(currentGroup.strategy) }}</el-descriptions-item>
          <el-descriptions-item label="节点数">{{ currentGroup.total_nodes || 0 }}</el-descriptions-item>
          <el-descriptions-item label="健康节点">{{ currentGroup.healthy_nodes || 0 }}</el-descriptions-item>
          <el-descriptions-item label="用户数">{{ currentGroup.total_users || 0 }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(currentGroup.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentGroup.description || '-' }}</el-descriptions-item>
        </el-descriptions>
        
        <div class="group-nodes-section" v-if="groupNodes.length">
          <h4>分组节点</h4>
          <el-table :data="groupNodes" size="small" style="width: 100%">
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="address" label="地址" />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="current_users" label="用户数" width="80" />
          </el-table>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Refresh, Search, Folder, Monitor, CircleCheck, User,
  MoreFilled, Edit, Setting, Delete, Location, Connection
} from '@element-plus/icons-vue'
import { useNodeGroupStore } from '@/stores/nodeGroup'
import { useNodeStore } from '@/stores/node'

const groupStore = useNodeGroupStore()
const nodeStore = useNodeStore()

const dialogVisible = ref(false)
const nodesDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const savingNodes = ref(false)
const formRef = ref(null)
const currentGroup = ref(null)
const groupNodes = ref([])
const nodeSearch = ref('')
const selectedNodeIds = ref([])

const form = reactive({
  id: null,
  name: '',
  description: '',
  region: '',
  strategy: 'round-robin'
})

const rules = {
  name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }]
}

const transferData = computed(() => {
  return nodeStore.nodes.map(node => ({
    key: node.id,
    label: `${node.name} (${node.address})`,
    disabled: false
  }))
})

const getStrategyText = (strategy) => {
  const texts = {
    'round-robin': '轮询',
    'least-connections': '最少连接',
    'weighted': '加权',
    'geographic': '地理位置'
  }
  return texts[strategy] || strategy
}

const getStatusType = (status) => {
  const types = { online: 'success', offline: 'info', unhealthy: 'danger' }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = { online: '在线', offline: '离线', unhealthy: '不健康' }
  return texts[status] || status
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const filterNodes = (query, item) => {
  return item.label.toLowerCase().includes(query.toLowerCase())
}

const fetchGroups = async () => {
  try {
    await groupStore.fetchGroupsWithStats()
  } catch (e) {
    ElMessage.error(e.message || '获取分组列表失败')
  }
}

const fetchAllNodes = async () => {
  try {
    await nodeStore.fetchNodes({ limit: 1000 })
  } catch (e) {
    console.error('获取节点列表失败:', e)
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, { id: null, name: '', description: '', region: '', strategy: 'round-robin' })
  dialogVisible.value = true
}

const editGroup = (group) => {
  isEdit.value = true
  Object.assign(form, {
    id: group.id,
    name: group.name,
    description: group.description || '',
    region: group.region || '',
    strategy: group.strategy || 'round-robin'
  })
  dialogVisible.value = true
}

const submitForm = async () => {
  await formRef.value.validate()
  submitting.value = true
  try {
    const data = {
      name: form.name,
      description: form.description,
      region: form.region,
      strategy: form.strategy
    }
    if (isEdit.value) {
      await groupStore.updateGroup(form.id, data)
      ElMessage.success('更新成功')
    } else {
      await groupStore.createGroup(data)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchGroups()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const deleteGroup = async (group) => {
  await ElMessageBox.confirm(
    `确定要删除分组 "${group.name}" 吗？分组内的节点不会被删除。`,
    '删除确认',
    { type: 'warning' }
  )
  try {
    await groupStore.deleteGroup(group.id)
    ElMessage.success('删除成功')
    fetchGroups()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

const manageNodes = async (group) => {
  currentGroup.value = group
  await fetchAllNodes()
  
  // 获取当前分组的节点
  try {
    const nodes = await groupStore.fetchGroupNodes(group.id)
    selectedNodeIds.value = nodes.map(n => n.id)
  } catch (e) {
    selectedNodeIds.value = []
  }
  
  nodesDialogVisible.value = true
}

const saveGroupNodes = async () => {
  savingNodes.value = true
  try {
    await groupStore.setGroupNodes(currentGroup.value.id, selectedNodeIds.value)
    ElMessage.success('保存成功')
    nodesDialogVisible.value = false
    fetchGroups()
  } catch (e) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    savingNodes.value = false
  }
}

const viewGroupDetail = async (group) => {
  currentGroup.value = group
  try {
    const nodes = await groupStore.fetchGroupNodes(group.id)
    groupNodes.value = nodes
  } catch (e) {
    groupNodes.value = []
  }
  detailDialogVisible.value = true
}

onMounted(fetchGroups)
</script>

<style scoped>
.admin-node-groups-page {
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

.group-card {
  margin-bottom: 20px;
  cursor: pointer;
  transition: transform 0.2s;
}

.group-card:hover {
  transform: translateY(-4px);
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.group-name {
  font-size: 16px;
  font-weight: 600;
}

.group-content {
  min-height: 120px;
}

.group-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

.group-description {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 12px;
  line-height: 1.5;
}

.group-stats {
  display: flex;
  justify-content: space-around;
}

.stat-item {
  text-align: center;
}

.stat-num {
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.stat-num.healthy {
  color: var(--el-color-success);
}

.stat-text {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  display: block;
  margin-top: 4px;
}

.nodes-dialog-content {
  padding: 10px 0;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.dialog-footer {
  margin-top: 20px;
  text-align: right;
}

.group-detail {
  padding: 10px 0;
}

.group-nodes-section {
  margin-top: 20px;
}

.group-nodes-section h4 {
  margin-bottom: 12px;
  color: var(--el-text-color-primary);
}
</style>
