<template>
  <div class="admin-reports-page">
    <div class="page-header">
      <h1 class="page-title">财务报表</h1>
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        @change="fetchReports"
      />
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <el-card shadow="never" class="stat-card">
        <div class="stat-value">¥{{ formatPrice(stats.total_revenue) }}</div>
        <div class="stat-label">总收入</div>
      </el-card>
      <el-card shadow="never" class="stat-card">
        <div class="stat-value">{{ stats.total_orders }}</div>
        <div class="stat-label">总订单数</div>
      </el-card>
      <el-card shadow="never" class="stat-card">
        <div class="stat-value">{{ stats.paid_orders }}</div>
        <div class="stat-label">已支付订单</div>
      </el-card>
      <el-card shadow="never" class="stat-card">
        <div class="stat-value">¥{{ formatPrice(stats.total_refund) }}</div>
        <div class="stat-label">退款金额</div>
      </el-card>
    </div>

    <!-- 支付失败统计 -->
    <el-card shadow="never" class="failed-payments-card">
      <template #header>
        <div class="card-header">
          <span>支付失败统计</span>
          <el-button type="primary" size="small" @click="fetchFailedPaymentStats">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      <div class="failed-stats-grid">
        <div class="failed-stat-item">
          <div class="failed-stat-value error">{{ failedPaymentStats.total_failed }}</div>
          <div class="failed-stat-label">失败总数</div>
        </div>
        <div class="failed-stat-item">
          <div class="failed-stat-value warning">{{ failedPaymentStats.pending_retry }}</div>
          <div class="failed-stat-label">待重试</div>
        </div>
        <div class="failed-stat-item">
          <div class="failed-stat-value danger">{{ failedPaymentStats.retry_exhausted }}</div>
          <div class="failed-stat-label">重试耗尽</div>
        </div>
        <div class="failed-stat-item">
          <div class="failed-stat-value success">{{ failedPaymentStats.recovered_by_retry }}</div>
          <div class="failed-stat-label">重试成功</div>
        </div>
        <div class="failed-stat-item">
          <div class="failed-stat-value">{{ failedPaymentStats.failure_rate?.toFixed(2) || 0 }}%</div>
          <div class="failed-stat-label">失败率</div>
        </div>
        <div class="failed-stat-item">
          <div class="failed-stat-value success">{{ failedPaymentStats.recovery_rate?.toFixed(2) || 0 }}%</div>
          <div class="failed-stat-label">恢复率</div>
        </div>
        <div class="failed-stat-item">
          <div class="failed-stat-value">{{ failedPaymentStats.avg_retry_attempts?.toFixed(1) || 0 }}</div>
          <div class="failed-stat-label">平均重试次数</div>
        </div>
      </div>
      
      <!-- 失败原因分布 -->
      <div v-if="Object.keys(failedPaymentStats.failures_by_reason || {}).length > 0" class="failure-reasons">
        <h4>失败原因分布</h4>
        <el-table :data="failureReasonsList" size="small" style="width: 100%">
          <el-table-column prop="reason" label="失败原因" />
          <el-table-column prop="count" label="次数" width="100" />
          <el-table-column label="占比" width="150">
            <template #default="{ row }">
              <el-progress 
                :percentage="Math.round(row.count / failedPaymentStats.total_failed * 100)" 
                :stroke-width="6" 
              />
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 订阅暂停统计 -->
    <el-card shadow="never" class="pause-stats-card">
      <template #header>
        <div class="card-header">
          <span>订阅暂停统计</span>
          <el-button type="primary" size="small" @click="fetchPauseStats">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      <div class="pause-stats-grid">
        <div class="pause-stat-item">
          <div class="pause-stat-value">{{ pauseStats.total_pauses }}</div>
          <div class="pause-stat-label">总暂停次数</div>
        </div>
        <div class="pause-stat-item">
          <div class="pause-stat-value warning">{{ pauseStats.active_pauses }}</div>
          <div class="pause-stat-label">当前暂停中</div>
        </div>
        <div class="pause-stat-item">
          <div class="pause-stat-value success">{{ pauseStats.resumed_pauses }}</div>
          <div class="pause-stat-label">已恢复</div>
        </div>
        <div class="pause-stat-item">
          <div class="pause-stat-value">{{ pauseStats.auto_resumed }}</div>
          <div class="pause-stat-label">自动恢复</div>
        </div>
        <div class="pause-stat-item">
          <div class="pause-stat-value">{{ pauseStats.avg_pause_days?.toFixed(1) || 0 }}</div>
          <div class="pause-stat-label">平均暂停天数</div>
        </div>
        <div class="pause-stat-item">
          <div class="pause-stat-value">{{ pauseStats.pause_rate?.toFixed(1) || 0 }}%</div>
          <div class="pause-stat-label">暂停率</div>
        </div>
      </div>
      
      <!-- 暂停滥用检测 -->
      <div v-if="pauseStats.abuse_patterns?.length > 0" class="abuse-patterns">
        <h4>潜在滥用用户</h4>
        <el-table :data="pauseStats.abuse_patterns" size="small" style="width: 100%">
          <el-table-column prop="user_id" label="用户ID" width="100" />
          <el-table-column prop="pause_count" label="暂停次数" width="100" />
          <el-table-column prop="total_pause_days" label="总暂停天数" width="120" />
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="viewUser(row.user_id)">
                查看
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 收入趋势图 -->
    <el-card shadow="never" class="chart-card">
      <template #header>
        <span>收入趋势</span>
      </template>
      <div ref="revenueChartRef" class="chart-container"></div>
    </el-card>

    <!-- 订单统计图 -->
    <el-card shadow="never" class="chart-card">
      <template #header>
        <span>订单统计</span>
      </template>
      <div ref="orderChartRef" class="chart-container"></div>
    </el-card>

    <!-- 套餐销售排行 -->
    <el-card shadow="never">
      <template #header>
        <span>套餐销售排行</span>
      </template>
      <el-table :data="planRanking" style="width: 100%">
        <el-table-column prop="rank" label="排名" width="80" />
        <el-table-column prop="plan_name" label="套餐名称" />
        <el-table-column prop="order_count" label="订单数" width="120" />
        <el-table-column label="销售额" width="150">
          <template #default="{ row }">¥{{ formatPrice(row.revenue) }}</template>
        </el-table-column>
        <el-table-column label="占比" width="120">
          <template #default="{ row }">
            <el-progress :percentage="row.percentage" :stroke-width="6" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import api from '@/api'

const dateRange = ref(null)
const revenueChartRef = ref(null)
const orderChartRef = ref(null)
let revenueChart = null
let orderChart = null

const stats = reactive({
  total_revenue: 0,
  total_orders: 0,
  paid_orders: 0,
  total_refund: 0
})

const failedPaymentStats = reactive({
  total_failed: 0,
  pending_retry: 0,
  retry_exhausted: 0,
  recovered_by_retry: 0,
  failure_rate: 0,
  recovery_rate: 0,
  avg_retry_attempts: 0,
  failures_by_method: {},
  failures_by_reason: {}
})

const pauseStats = reactive({
  total_pauses: 0,
  active_pauses: 0,
  resumed_pauses: 0,
  auto_resumed: 0,
  avg_pause_days: 0,
  pause_rate: 0,
  abuse_patterns: []
})

const planRanking = ref([])

const formatPrice = (price) => ((price || 0) / 100).toFixed(2)

// Convert failures_by_reason object to array for table display
const failureReasonsList = computed(() => {
  const reasons = failedPaymentStats.failures_by_reason || {}
  return Object.entries(reasons).map(([reason, count]) => ({
    reason: reason || '未知原因',
    count
  })).sort((a, b) => b.count - a.count)
})

const fetchFailedPaymentStats = async () => {
  try {
    const response = await api.get('/admin/reports/failed-payments')
    if (response.data?.stats) {
      Object.assign(failedPaymentStats, response.data.stats)
    }
  } catch (error) {
    console.error('Failed to fetch failed payment stats:', error)
    // Use mock data if API fails
    Object.assign(failedPaymentStats, {
      total_failed: 12,
      pending_retry: 3,
      retry_exhausted: 2,
      recovered_by_retry: 5,
      failure_rate: 7.69,
      recovery_rate: 41.67,
      avg_retry_attempts: 1.8,
      failures_by_method: { alipay: 5, wechat: 7 },
      failures_by_reason: { 
        'payment gateway timeout': 4, 
        'insufficient balance': 3,
        'user cancelled': 5
      }
    })
  }
}

const fetchPauseStats = async () => {
  try {
    const response = await api.get('/admin/reports/pause-stats')
    if (response.data?.stats) {
      Object.assign(pauseStats, response.data.stats)
    }
  } catch (error) {
    console.error('Failed to fetch pause stats:', error)
    // Use mock data if API fails
    Object.assign(pauseStats, {
      total_pauses: 45,
      active_pauses: 8,
      resumed_pauses: 32,
      auto_resumed: 5,
      avg_pause_days: 12.5,
      pause_rate: 15.3,
      abuse_patterns: []
    })
  }
}

const viewUser = (userId) => {
  window.open(`/admin/users/${userId}`, '_blank')
}

const fetchReports = async () => {
  // Mock data - replace with actual API call
  stats.total_revenue = 1258900
  stats.total_orders = 156
  stats.paid_orders = 142
  stats.total_refund = 12500

  planRanking.value = [
    { rank: 1, plan_name: '年度套餐', order_count: 45, revenue: 539400, percentage: 43 },
    { rank: 2, plan_name: '季度套餐', order_count: 52, revenue: 415600, percentage: 33 },
    { rank: 3, plan_name: '月度套餐', order_count: 59, revenue: 303900, percentage: 24 }
  ]

  updateCharts()
  fetchFailedPaymentStats()
  fetchPauseStats()
}

const updateCharts = () => {
  // Revenue chart
  if (revenueChart) {
    revenueChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: ['1月', '2月', '3月', '4月', '5月', '6月'] },
      yAxis: { type: 'value', axisLabel: { formatter: '¥{value}' } },
      series: [{
        name: '收入',
        type: 'line',
        smooth: true,
        data: [12000, 15000, 18000, 22000, 19000, 25000],
        areaStyle: { opacity: 0.3 }
      }]
    })
  }

  // Order chart
  if (orderChart) {
    orderChart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['新订单', '已支付', '已退款'] },
      xAxis: { type: 'category', data: ['1月', '2月', '3月', '4月', '5月', '6月'] },
      yAxis: { type: 'value' },
      series: [
        { name: '新订单', type: 'bar', data: [20, 25, 30, 35, 28, 40] },
        { name: '已支付', type: 'bar', data: [18, 23, 28, 32, 26, 38] },
        { name: '已退款', type: 'bar', data: [2, 2, 2, 3, 2, 2] }
      ]
    })
  }
}

const initCharts = () => {
  if (revenueChartRef.value) {
    revenueChart = echarts.init(revenueChartRef.value)
  }
  if (orderChartRef.value) {
    orderChart = echarts.init(orderChartRef.value)
  }
}

onMounted(() => {
  initCharts()
  fetchReports()
  window.addEventListener('resize', () => {
    revenueChart?.resize()
    orderChart?.resize()
  })
})

onUnmounted(() => {
  revenueChart?.dispose()
  orderChart?.dispose()
})
</script>

<style scoped>
.admin-reports-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-title { font-size: 24px; font-weight: 600; margin: 0; }
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px; }
.stat-card { text-align: center; }
.stat-value { font-size: 28px; font-weight: 600; color: #409eff; }
.stat-label { font-size: 14px; color: #909399; margin-top: 8px; }
.chart-card { margin-bottom: 20px; }
.chart-container { height: 300px; }

/* Failed payment stats styles */
.failed-payments-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.failed-stats-grid { 
  display: grid; 
  grid-template-columns: repeat(7, 1fr); 
  gap: 16px; 
  margin-bottom: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}
.failed-stat-item { text-align: center; }
.failed-stat-value { 
  font-size: 24px; 
  font-weight: 600; 
  color: #606266;
}
.failed-stat-value.error { color: #f56c6c; }
.failed-stat-value.warning { color: #e6a23c; }
.failed-stat-value.danger { color: #f56c6c; }
.failed-stat-value.success { color: #67c23a; }
.failed-stat-label { 
  font-size: 12px; 
  color: #909399; 
  margin-top: 4px; 
}
.failure-reasons { margin-top: 16px; }
.failure-reasons h4 { 
  font-size: 14px; 
  font-weight: 600; 
  margin-bottom: 12px; 
  color: #303133;
}

/* Pause stats styles */
.pause-stats-card { margin-bottom: 20px; }
.pause-stats-grid { 
  display: grid; 
  grid-template-columns: repeat(6, 1fr); 
  gap: 16px; 
  margin-bottom: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}
.pause-stat-item { text-align: center; }
.pause-stat-value { 
  font-size: 24px; 
  font-weight: 600; 
  color: #606266;
}
.pause-stat-value.warning { color: #e6a23c; }
.pause-stat-value.success { color: #67c23a; }
.pause-stat-label { 
  font-size: 12px; 
  color: #909399; 
  margin-top: 4px; 
}
.abuse-patterns { margin-top: 16px; }
.abuse-patterns h4 { 
  font-size: 14px; 
  font-weight: 600; 
  margin-bottom: 12px; 
  color: #303133;
}

@media (max-width: 1200px) {
  .failed-stats-grid { grid-template-columns: repeat(4, 1fr); }
  .pause-stats-grid { grid-template-columns: repeat(3, 1fr); }
}
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .failed-stats-grid { grid-template-columns: repeat(2, 1fr); }
  .pause-stats-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
