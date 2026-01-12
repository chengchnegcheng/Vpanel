<template>
  <div class="proxy-management">
    <div class="header">
      <h1>代理管理</h1>
      <el-button type="primary" @click="showAddProxyDialog = true">
        <el-icon><Plus /></el-icon>
        添加代理
      </el-button>
    </div>

    <el-alert
      v-if="error"
      title="加载失败"
      type="error"
      :description="error"
      show-icon
      closable
    />

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
      <el-skeleton :rows="3" animated />
    </div>

    <div v-else-if="!proxyList.length" class="empty-state">
      <el-empty description="暂无代理配置">
        <el-button type="primary" @click="showAddProxyDialog = true">添加代理</el-button>
      </el-empty>
    </div>

    <div v-else class="proxy-grid">
      <ProxyCard
        v-for="proxy in proxyList"
        :key="proxy.id"
        :proxy="proxy"
        @edit="handleEditProxy"
        @delete="handleDeleteProxy"
      />
    </div>

    <!-- 添加/编辑代理对话框 -->
    <el-dialog
      v-model="showAddProxyDialog"
      :title="isEditing ? '编辑代理' : '添加代理'"
      width="500px"
      :close-on-click-modal="false"
      @closed="resetForm"
    >
      <el-form
        ref="proxyFormRef"
        :model="proxyForm"
        :rules="formRules"
        label-width="80px"
        label-position="left"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="proxyForm.name" placeholder="请输入代理名称" />
        </el-form-item>

        <el-form-item label="协议" prop="protocol">
          <el-select v-model="proxyForm.protocol" placeholder="请选择协议类型" style="width: 100%">
            <el-option label="Shadowsocks" value="shadowsocks" />
            <el-option label="Trojan" value="trojan" />
            <el-option label="VMess" value="vmess" />
            <el-option label="VLESS" value="vless" />
            <el-option label="SOCKS" value="socks" />
          </el-select>
        </el-form-item>

        <el-form-item label="服务器" prop="server">
          <el-input v-model="proxyForm.server" placeholder="请输入服务器地址" />
        </el-form-item>

        <el-form-item label="端口" prop="port">
          <el-input-number v-model="proxyForm.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>

        <template v-if="proxyForm.protocol === 'shadowsocks'">
          <el-form-item label="密码" prop="password">
            <el-input v-model="proxyForm.password" type="password" show-password placeholder="请输入密码" />
          </el-form-item>

          <el-form-item label="加密方式" prop="method">
            <el-select v-model="proxyForm.method" placeholder="请选择加密方式" style="width: 100%">
              <el-option label="aes-256-gcm" value="aes-256-gcm" />
              <el-option label="chacha20-ietf-poly1305" value="chacha20-ietf-poly1305" />
              <el-option label="2022-blake3-aes-256-gcm" value="2022-blake3-aes-256-gcm" />
            </el-select>
          </el-form-item>
        </template>

        <template v-else-if="proxyForm.protocol === 'trojan'">
          <el-form-item label="密码" prop="password">
            <el-input v-model="proxyForm.password" type="password" show-password placeholder="请输入密码" />
          </el-form-item>
          
          <el-form-item label="SNI" prop="sni">
            <el-input v-model="proxyForm.sni" placeholder="请输入SNI" />
          </el-form-item>
        </template>

        <template v-else-if="['vmess', 'vless'].includes(proxyForm.protocol)">
          <el-form-item label="UUID" prop="uuid">
            <el-input v-model="proxyForm.uuid" placeholder="请输入UUID" />
          </el-form-item>
          
          <el-form-item label="加密" prop="security" v-if="proxyForm.protocol === 'vmess'">
            <el-select v-model="proxyForm.security" placeholder="请选择加密方式" style="width: 100%">
              <el-option label="auto" value="auto" />
              <el-option label="aes-128-gcm" value="aes-128-gcm" />
              <el-option label="chacha20-poly1305" value="chacha20-poly1305" />
              <el-option label="none" value="none" />
            </el-select>
          </el-form-item>
          
          <el-form-item label="传输" prop="network">
            <el-select v-model="proxyForm.network" placeholder="请选择传输协议" style="width: 100%">
              <el-option label="tcp" value="tcp" />
              <el-option label="kcp" value="kcp" />
              <el-option label="ws" value="ws" />
              <el-option label="http" value="http" />
              <el-option label="quic" value="quic" />
              <el-option label="grpc" value="grpc" />
            </el-select>
          </el-form-item>
        </template>

        <el-form-item label="备注" prop="remark">
          <el-input v-model="proxyForm.remark" type="textarea" :rows="2" placeholder="可选，添加备注信息" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showAddProxyDialog = false">取消</el-button>
        <el-button type="primary" @click="submitProxyForm" :loading="submitting">
          {{ isEditing ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import ProxyCard from '../components/ProxyCard.vue'
import api from '@/api/index'

// 代理列表
const proxyList = ref([])
const loading = ref(true)
const error = ref('')

// 表单相关
const proxyFormRef = ref(null)
const showAddProxyDialog = ref(false)
const submitting = ref(false)
const currentEditId = ref(null)

const isEditing = computed(() => currentEditId.value !== null)

// 表单数据
const proxyForm = reactive({
  name: '',
  protocol: 'shadowsocks',
  server: '',
  port: 443,
  password: '',
  method: 'aes-256-gcm',
  uuid: '',
  security: 'auto',
  network: 'tcp',
  sni: '',
  remark: ''
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入代理名称', trigger: 'blur' },
    { min: 1, max: 50, message: '长度在 1 到 50 个字符', trigger: 'blur' }
  ],
  protocol: [
    { required: true, message: '请选择协议类型', trigger: 'change' }
  ],
  server: [
    { required: true, message: '请输入服务器地址', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口范围 1-65535', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  method: [
    { required: true, message: '请选择加密方式', trigger: 'change' }
  ],
  uuid: [
    { required: true, message: '请输入UUID', trigger: 'blur' }
  ],
  security: [
    { required: true, message: '请选择加密方式', trigger: 'change' }
  ],
  network: [
    { required: true, message: '请选择传输协议', trigger: 'change' }
  ]
}

// 获取代理列表
const fetchProxyList = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const response = await api.get('/proxies')
    proxyList.value = Array.isArray(response) ? response : (response.data || [])
  } catch (err) {
    console.error('获取代理列表失败:', err)
    error.value = '获取代理列表失败'
    proxyList.value = []
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  if (proxyFormRef.value) {
    proxyFormRef.value.resetFields()
  }
  
  Object.assign(proxyForm, {
    name: '',
    protocol: 'shadowsocks',
    server: '',
    port: 443,
    password: '',
    method: 'aes-256-gcm',
    uuid: '',
    security: 'auto',
    network: 'tcp',
    sni: '',
    remark: ''
  })
  
  currentEditId.value = null
}

// 提交表单
const submitProxyForm = async () => {
  if (!proxyFormRef.value) return
  
  await proxyFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    
    try {
      // 根据协议准备不同的数据
      const requestData = { ...proxyForm }
      
      if (isEditing.value) {
        await api.put(`/proxies/${currentEditId.value}`, requestData)
      } else {
        await api.post('/proxies', requestData)
      }
      
      // 刷新代理列表
      fetchProxyList()
      
      // 关闭对话框
      showAddProxyDialog.value = false
      
      // 显示成功消息
      ElMessage.success(isEditing.value ? '代理更新成功' : '代理添加成功')
    } catch (err) {
      console.error('提交表单失败:', err)
      ElMessage.error(isEditing.value ? '代理更新失败' : '代理添加失败')
    } finally {
      submitting.value = false
    }
  })
}

// 处理编辑代理
const handleEditProxy = (id) => {
  const proxy = proxyList.value.find(item => item.id === id)
  
  if (!proxy) {
    ElMessage.error('未找到代理信息')
    return
  }
  
  // 设置表单数据
  Object.keys(proxyForm).forEach(key => {
    if (key in proxy) {
      proxyForm[key] = proxy[key]
    }
  })
  
  currentEditId.value = id
  showAddProxyDialog.value = true
}

// 处理删除代理
const handleDeleteProxy = async (id) => {
  try {
    await api.delete(`/proxies/${id}`)
    
    // 刷新代理列表
    fetchProxyList()
    
    ElMessage.success('代理删除成功')
  } catch (err) {
    console.error('删除代理失败:', err)
    ElMessage.error('删除代理失败')
  }
}

// 页面加载时获取代理列表
onMounted(() => {
  fetchProxyList()
})
</script>

<style scoped>
.proxy-management {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 500;
}

.loading-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.empty-state {
  margin-top: 50px;
  text-align: center;
}

.proxy-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

@media (max-width: 768px) {
  .proxy-grid {
    grid-template-columns: 1fr;
  }
}
</style> 