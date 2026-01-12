<template>
  <div class="logs-container">
    <div class="header">
      <h1>日志管理</h1>
      <div class="actions">
        <el-select v-model="logType" placeholder="日志类型" style="width: 150px;">
          <el-option label="全部日志" value="all" />
          <el-option label="系统日志" value="system" />
          <el-option label="错误日志" value="error" />
          <el-option label="访问日志" value="access" />
          <el-option label="用户操作" value="user" />
        </el-select>
        <el-select v-model="logLevel" placeholder="日志级别" style="width: 150px;">
          <el-option label="全部级别" value="all" />
          <el-option label="INFO" value="info" />
          <el-option label="WARNING" value="warning" />
          <el-option label="ERROR" value="error" />
          <el-option label="CRITICAL" value="critical" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          :shortcuts="dateShortcuts"
        />
        <el-input
          v-model="searchQuery"
          placeholder="搜索关键词"
          style="width: 200px;"
          clearable
        />
        <el-button type="primary" @click="fetchLogs">查询</el-button>
        <el-button type="success" @click="exportLogs">导出</el-button>
      </div>
    </div>

    <!-- 日志表格 -->
    <el-card class="log-table-card">
      <el-table 
        :data="filteredLogs" 
        v-loading="loading" 
        style="width: 100%"
        height="calc(100vh - 250px)"
      >
        <el-table-column type="expand">
          <template #default="props">
            <div class="log-detail">
              <p><strong>日志详情：</strong></p>
              <pre>{{ props.row.message }}</pre>
              <p v-if="props.row.stackTrace"><strong>堆栈跟踪：</strong></p>
              <pre v-if="props.row.stackTrace">{{ props.row.stackTrace }}</pre>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="timestamp" label="时间" width="180" sortable>
          <template #default="scope">
            {{ formatDateTime(scope.row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="120">
          <template #default="scope">
            <el-tag 
              :type="getLogTypeTag(scope.row.type)"
              size="small"
            >
              {{ getLogTypeText(scope.row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="level" label="级别" width="100">
          <template #default="scope">
            <el-tag 
              :type="getLogLevelTag(scope.row.level)"
              size="small"
            >
              {{ scope.row.level.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="150" />
        <el-table-column prop="message" label="内容">
          <template #default="scope">
            <div class="message-preview">{{ scope.row.message }}</div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button 
              size="small" 
              type="primary" 
              @click="viewLogDetail(scope.row)"
            >
              详情
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteLog(scope.row)"
              v-if="hasDeletePermission"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100, 200]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="totalLogs"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 日志详情对话框 -->
    <el-dialog 
      title="日志详情" 
      v-model="logDetailVisible" 
      width="700px"
    >
      <template v-if="selectedLog">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="时间">{{ formatDateTime(selectedLog.timestamp) }}</el-descriptions-item>
          <el-descriptions-item label="来源">{{ selectedLog.source }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="getLogTypeTag(selectedLog.type)">
              {{ getLogTypeText(selectedLog.type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="级别">
            <el-tag :type="getLogLevelTag(selectedLog.level)">
              {{ selectedLog.level.toUpperCase() }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="IP地址" v-if="selectedLog.ip">{{ selectedLog.ip }}</el-descriptions-item>
          <el-descriptions-item label="用户" v-if="selectedLog.username">{{ selectedLog.username }}</el-descriptions-item>
        </el-descriptions>

        <div class="log-content">
          <p><strong>日志内容：</strong></p>
          <pre>{{ selectedLog.message }}</pre>
        </div>

        <div class="stack-trace" v-if="selectedLog.stackTrace">
          <p><strong>堆栈跟踪：</strong></p>
          <pre>{{ selectedLog.stackTrace }}</pre>
        </div>

        <div class="metadata" v-if="selectedLog.metadata">
          <p><strong>元数据：</strong></p>
          <pre>{{ JSON.stringify(selectedLog.metadata, null, 2) }}</pre>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/index'

export default {
  name: 'Logs',
  setup() {
    // 状态
    const loading = ref(false)
    const logs = ref([])
    const totalLogs = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(20)
    const logType = ref('all')
    const logLevel = ref('all')
    const searchQuery = ref('')
    const dateRange = ref([])
    const logDetailVisible = ref(false)
    const selectedLog = ref(null)
    const hasDeletePermission = ref(true) // 根据实际权限控制

    // 日期快捷选项
    const dateShortcuts = [
      {
        text: '今天',
        value: () => {
          const end = new Date()
          const start = new Date()
          start.setHours(0, 0, 0, 0)
          return [start, end]
        }
      },
      {
        text: '昨天',
        value: () => {
          const end = new Date()
          end.setHours(23, 59, 59, 999)
          end.setDate(end.getDate() - 1)
          const start = new Date()
          start.setHours(0, 0, 0, 0)
          start.setDate(start.getDate() - 1)
          return [start, end]
        }
      },
      {
        text: '最近一周',
        value: () => {
          const end = new Date()
          const start = new Date()
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 7)
          return [start, end]
        }
      },
      {
        text: '最近一个月',
        value: () => {
          const end = new Date()
          const start = new Date()
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 30)
          return [start, end]
        }
      }
    ]

    // 过滤后的日志数据
    const filteredLogs = computed(() => {
      return logs.value
    })

    // 获取日志数据
    const fetchLogs = async () => {
      loading.value = true
      
      try {
        const params = {
          page: currentPage.value,
          size: pageSize.value,
          type: logType.value !== 'all' ? logType.value : undefined,
          level: logLevel.value !== 'all' ? logLevel.value : undefined,
          query: searchQuery.value || undefined,
          startTime: dateRange.value && dateRange.value[0] ? dateRange.value[0].toISOString() : undefined,
          endTime: dateRange.value && dateRange.value[1] ? dateRange.value[1].toISOString() : undefined
        }
        
        const response = await api.get('/logs', { params })
        const data = response.data || response
        logs.value = data.logs || []
        totalLogs.value = data.total || 0
      } catch (error) {
        ElMessage.error('获取日志数据失败')
        console.error(error)
        logs.value = []
        totalLogs.value = 0
      } finally {
        loading.value = false
      }
    }

    // 导出日志数据
    const exportLogs = () => {
      ElMessage.success('日志导出成功')
    }

    // 查看日志详情
    const viewLogDetail = (log) => {
      selectedLog.value = log
      logDetailVisible.value = true
    }

    // 删除日志
    const deleteLog = (log) => {
      ElMessageBox.confirm(
        '确定要删除这条日志记录吗？此操作不可恢复。',
        '警告',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(async () => {
        try {
          // TODO: 替换为实际的 API 调用
          // TODO: 替换为实际 API 调用
          // await logsService.deleteLog(log.id)
          
          // 从列表中移除
          logs.value = logs.value.filter(item => item.id !== log.id)
          totalLogs.value--
          
          ElMessage.success('日志删除成功')
        } catch (error) {
          ElMessage.error('日志删除失败')
          console.error(error)
        }
      }).catch(() => {})
    }

    // 分页处理
    const handleSizeChange = (size) => {
      pageSize.value = size
      fetchLogs()
    }

    const handleCurrentChange = (page) => {
      currentPage.value = page
      fetchLogs()
    }

    // 获取日志类型标签样式
    const getLogTypeTag = (type) => {
      const typeMap = {
        'system': 'primary',
        'error': 'danger',
        'access': 'success',
        'user': 'warning'
      }
      return typeMap[type] || 'info'
    }

    // 获取日志类型文本
    const getLogTypeText = (type) => {
      const typeMap = {
        'system': '系统日志',
        'error': '错误日志',
        'access': '访问日志',
        'user': '用户操作'
      }
      return typeMap[type] || type
    }

    // 获取日志级别标签样式
    const getLogLevelTag = (level) => {
      const levelMap = {
        'info': 'info',
        'warning': 'warning',
        'error': 'danger',
        'critical': 'danger'
      }
      return levelMap[level] || 'info'
    }

    // 格式化日期时间
    const formatDateTime = (date) => {
      if (!date) return '未知'
      const d = new Date(date)
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
    }

    // 初始化
    onMounted(() => {
      fetchLogs()
    })

    return {
      logs,
      filteredLogs,
      loading,
      totalLogs,
      currentPage,
      pageSize,
      logType,
      logLevel,
      searchQuery,
      dateRange,
      dateShortcuts,
      logDetailVisible,
      selectedLog,
      hasDeletePermission,
      fetchLogs,
      exportLogs,
      viewLogDetail,
      deleteLog,
      handleSizeChange,
      handleCurrentChange,
      getLogTypeTag,
      getLogTypeText,
      getLogLevelTag,
      formatDateTime
    }
  }
}
</script>

<style scoped>
.logs-container {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.actions {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.log-table-card {
  margin-bottom: 20px;
}

.message-preview {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 400px;
}

.log-detail {
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.log-detail pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
  font-family: monospace;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.log-content,
.stack-trace,
.metadata {
  margin-top: 20px;
}

.log-content pre,
.stack-trace pre,
.metadata pre {
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
  font-family: monospace;
  max-height: 300px;
  overflow: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style> 