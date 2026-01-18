# 前端调试指南 - IP 限制错误

## 问题现象
访问 IP 限制管理页面时显示"应用错误"，错误ID: ERR-MKJA0508-YSMCTL

## 后端状态
✅ 后端服务正常运行  
✅ API 端点可访问  
✅ 数据库连接正常  

**结论**: 问题出在前端或认证层面

## 立即执行的调试步骤

### 步骤 1: 打开浏览器开发者工具

1. 在浏览器中按 `F12` 或 `Cmd+Option+I` (Mac)
2. 确保开发者工具已打开

### 步骤 2: 查看 Console 标签

在 Console 标签中运行以下命令：

```javascript
// 1. 检查 token
console.log('Token:', localStorage.getItem('token'))
console.log('User Token:', localStorage.getItem('userToken'))

// 2. 检查当前路径
console.log('Current Path:', window.location.pathname)

// 3. 手动测试 API
const token = localStorage.getItem('token')
fetch('http://localhost:8081/api/admin/ip-restrictions/stats', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
})
.then(r => {
  console.log('Status:', r.status)
  return r.json()
})
.then(data => {
  console.log('Response:', data)
})
.catch(err => {
  console.error('Error:', err)
})
```

### 步骤 3: 查看 Network 标签

1. 切换到 **Network** 标签
2. 刷新页面 (F5 或 Cmd+R)
3. 找到 `ip-restrictions/stats` 请求
4. 点击该请求，查看：
   - **Headers**: 请求头信息
   - **Response**: 响应内容
   - **Status Code**: HTTP 状态码

## 根据状态码的解决方案

### 情况 A: 401 Unauthorized (最常见)

**症状**:
- Network 标签显示 401 状态码
- 响应内容: `{"error": "Unauthorized"}` 或类似

**原因**: Token 已过期或无效

**解决方案**:
1. 点击右上角的用户头像或用户名
2. 选择"退出登录"
3. 重新登录
4. 再次访问 IP 限制页面

**或者在 Console 中执行**:
```javascript
// 清除 token
localStorage.removeItem('token')
localStorage.removeItem('userToken')
localStorage.removeItem('userInfo')

// 刷新页面
window.location.reload()
```

### 情况 B: 503 Service Unavailable

**症状**:
- Network 标签显示 503 状态码
- 响应内容: `{"code": 503, "message": "IP restriction service is not available"}`

**原因**: IP 服务未初始化

**解决方案**:
```bash
# 在终端中执行
pkill vpanel
./vpanel

# 等待 5 秒
sleep 5

# 刷新浏览器页面
```

### 情况 C: 500 Internal Server Error

**症状**:
- Network 标签显示 500 状态码
- 响应内容包含数据库错误

**原因**: 数据库查询失败

**解决方案**:
```bash
# 查看服务日志
tail -f logs/app.log

# 重启服务
pkill vpanel
./vpanel
```

### 情况 D: CORS Error

**症状**:
- Console 显示 CORS 相关错误
- Network 标签中请求显示为 "CORS error"

**原因**: 跨域请求被阻止

**解决方案**:

1. 检查前端 API 配置:
```bash
# 查看前端环境变量
cat web/.env.local
# 或
cat web/.env
```

应该包含:
```
VITE_APP_API_URL=http://localhost:8081/api
```

2. 如果文件不存在，创建它:
```bash
echo "VITE_APP_API_URL=http://localhost:8081/api" > web/.env.local
```

3. 重启前端开发服务器:
```bash
cd web
npm run dev
```

### 情况 E: Network Error / Failed to fetch

**症状**:
- Console 显示 "Network Error" 或 "Failed to fetch"
- Network 标签中请求显示为 "failed"

**原因**: 无法连接到后端

**解决方案**:

1. 确认后端正在运行:
```bash
ps aux | grep vpanel
```

2. 确认端口正确:
```bash
lsof -i :8081
```

3. 测试 API 连接:
```bash
curl http://localhost:8081/api/health
```

4. 如果后端未运行，启动它:
```bash
./vpanel
```

## 完整的前端调试流程

### 1. 收集信息

在浏览器 Console 中运行:
```javascript
// 收集所有相关信息
const debugInfo = {
  token: localStorage.getItem('token') ? 'exists' : 'missing',
  userToken: localStorage.getItem('userToken') ? 'exists' : 'missing',
  currentPath: window.location.pathname,
  apiUrl: import.meta.env.VITE_APP_API_URL || 'not set'
}
console.table(debugInfo)
```

### 2. 测试 API 连接

```javascript
// 测试健康检查端点
fetch('http://localhost:8081/api/health')
  .then(r => r.json())
  .then(data => console.log('Health Check:', data))
  .catch(err => console.error('Health Check Failed:', err))
```

### 3. 测试认证

```javascript
// 测试 IP 统计 API
const token = localStorage.getItem('token')
if (!token) {
  console.error('No token found! Please login first.')
} else {
  fetch('http://localhost:8081/api/admin/ip-restrictions/stats', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  })
  .then(async r => {
    const data = await r.json()
    console.log('Status:', r.status)
    console.log('Response:', data)
    
    if (r.status === 401) {
      console.error('Token expired! Please re-login.')
    } else if (r.status === 503) {
      console.error('Service unavailable! Please restart backend.')
    } else if (r.status === 500) {
      console.error('Server error! Check backend logs.')
    } else if (r.status === 200) {
      console.log('✓ API works! The issue might be in the Vue component.')
    }
  })
  .catch(err => {
    console.error('Network error:', err)
    console.error('Backend might not be running!')
  })
}
```

## 前端代码检查

### 检查 API 基础配置

查看 `web/src/api/base.js`:
```javascript
// 确认 baseURL 配置正确
const api = axios.create({
  baseURL: import.meta.env.VITE_APP_API_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})
```

### 检查 IP 限制页面

查看 `web/src/views/IPRestriction.vue` 的 `refreshStats` 方法:
```javascript
const refreshStats = async () => {
  statsLoading.value = true
  try {
    const response = await api.get('/admin/ip-restrictions/stats')
    // 检查响应格式
    if (response.code === 200 && response.data) {
      stats.activeDevices = response.data.total_active_ips || 0
      // ...
    }
  } catch (error) {
    console.error('Failed to fetch stats:', error)
    ElMessage.error(`获取统计数据失败: ${error.message || '未知错误'}`)
  } finally {
    statsLoading.value = false
  }
}
```

## 快速修复方案

### 方案 1: 重新登录 (最常见)

1. 退出登录
2. 清除浏览器缓存 (Cmd+Shift+Delete 或 Ctrl+Shift+Delete)
3. 重新登录
4. 访问 IP 限制页面

### 方案 2: 重启所有服务

```bash
# 停止后端
pkill vpanel

# 停止前端 (在前端终端按 Ctrl+C)

# 启动后端
./vpanel &

# 等待启动
sleep 5

# 启动前端
cd web
npm run dev
```

### 方案 3: 清除所有缓存

在浏览器 Console 中:
```javascript
// 清除所有 localStorage
localStorage.clear()

// 清除所有 sessionStorage
sessionStorage.clear()

// 刷新页面
window.location.reload()
```

然后重新登录。

## 预防措施

### 1. 使用更长的 Token 过期时间

修改 `configs/config.yaml`:
```yaml
jwt:
  expiration: 86400  # 24 小时，可以改为更长
```

### 2. 添加 Token 刷新机制

在前端添加自动刷新 token 的逻辑。

### 3. 改进错误提示

修改 `web/src/api/base.js` 的错误处理:
```javascript
if (error.response?.status === 401) {
  ElMessage.error('登录已过期，请重新登录')
  // 自动跳转到登录页
  setTimeout(() => {
    router.push('/user/login')
  }, 1500)
}
```

## 获取更多帮助

如果以上步骤都无法解决问题，请提供：

1. **浏览器 Console 的完整输出**
2. **Network 标签中 `ip-restrictions/stats` 请求的详细信息**:
   - Request Headers
   - Response Headers
   - Response Body
   - Status Code
3. **后端日志**: `tail -n 100 logs/app.log`

### 收集诊断信息的脚本

在浏览器 Console 中运行:
```javascript
// 收集完整的诊断信息
const collectDiagnostics = async () => {
  const info = {
    timestamp: new Date().toISOString(),
    browser: navigator.userAgent,
    url: window.location.href,
    token: localStorage.getItem('token') ? 'exists (length: ' + localStorage.getItem('token').length + ')' : 'missing',
    apiUrl: import.meta.env.VITE_APP_API_URL || 'not set'
  }
  
  // 测试 API
  try {
    const token = localStorage.getItem('token')
    const response = await fetch('http://localhost:8081/api/admin/ip-restrictions/stats', {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    info.apiStatus = response.status
    info.apiResponse = await response.text()
  } catch (err) {
    info.apiError = err.message
  }
  
  console.log('=== Diagnostic Information ===')
  console.log(JSON.stringify(info, null, 2))
  console.log('=== End ===')
  
  return info
}

// 运行诊断
collectDiagnostics()
```

## 相关文档

- [IP 限制排查指南](./ip-restriction-troubleshooting.md)
- [错误修复总结](./error-fix-summary.md)
- [快速修复参考](./quick-fix-reference.md)
