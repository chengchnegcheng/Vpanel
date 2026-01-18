# IP 限制错误排查指南

## 问题描述
访问 IP 限制管理页面时出现"应用错误"，错误ID: ERR-MKJA0508-YSMCTL

## 快速诊断步骤

### 1. 检查服务是否运行
```bash
# 检查进程
ps aux | grep vpanel

# 如果没有运行，启动服务
./vpanel
```

### 2. 检查浏览器控制台
1. 打开浏览器开发者工具 (F12)
2. 切换到 **Console** 标签
3. 刷新页面，查看是否有错误信息
4. 切换到 **Network** 标签
5. 找到 `ip-restrictions/stats` 请求
6. 查看请求的响应内容

### 3. 检查认证状态
在浏览器控制台中运行：
```javascript
// 检查 token 是否存在
console.log('Token:', localStorage.getItem('token'))
console.log('User Token:', localStorage.getItem('userToken'))

// 检查当前路径
console.log('Current Path:', window.location.pathname)
```

### 4. 手动测试 API

#### 方法 A: 使用浏览器控制台
```javascript
// 在浏览器控制台中运行
const token = localStorage.getItem('token')
fetch('http://localhost:8081/api/admin/ip-restrictions/stats', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
})
.then(r => r.json())
.then(data => console.log('API Response:', data))
.catch(err => console.error('API Error:', err))
```

#### 方法 B: 使用 curl
```bash
# 替换 YOUR_TOKEN 为你的实际 token
TOKEN="YOUR_TOKEN"

curl -X GET "http://localhost:8081/api/admin/ip-restrictions/stats" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | jq .
```

## 常见问题和解决方案

### 问题 1: 401 Unauthorized (未授权)
**症状**: API 返回 401 状态码

**原因**:
- Token 已过期
- Token 无效
- 未登录

**解决方案**:
1. 退出登录
2. 重新登录
3. 确保使用正确的登录页面（管理后台 vs 用户门户）

### 问题 2: 503 Service Unavailable (服务不可用)
**症状**: API 返回 503 状态码，消息为 "IP restriction service is not available"

**原因**:
- IP 服务未初始化
- 数据库连接失败

**解决方案**:
```bash
# 1. 重启服务
pkill vpanel
./vpanel

# 2. 检查日志
tail -f logs/app.log

# 3. 检查数据库连接
# 查看配置文件
cat configs/config.yaml

# 4. 验证数据库表
# 如果使用 PostgreSQL:
PGPASSWORD=vpanel123 psql -h localhost -U vpanel -d vpanel -c "\dt *ip*"

# 如果使用 MySQL:
mysql -u root -p vpanel -e "SHOW TABLES LIKE '%ip%';"
```

### 问题 3: 500 Internal Server Error (服务器内部错误)
**症状**: API 返回 500 状态码

**原因**:
- 数据库查询失败
- 数据库表不存在

**解决方案**:
```bash
# 1. 查看服务日志
tail -f logs/app.log

# 2. 重新运行数据库迁移
# 重启服务会自动运行 AutoMigrate
pkill vpanel
./vpanel

# 3. 等待几秒让迁移完成
sleep 5

# 4. 再次测试 API
```

### 问题 4: CORS 错误
**症状**: 浏览器控制台显示 CORS 相关错误

**原因**:
- 前端和后端端口不匹配
- CORS 配置问题

**解决方案**:
1. 检查前端配置 `web/.env` 或 `web/.env.local`:
```bash
VITE_APP_API_URL=http://localhost:8081/api
```

2. 检查后端 CORS 配置 `configs/config.yaml`:
```yaml
server:
  cors_origins:
    - "*"
```

3. 重启前后端服务

### 问题 5: 网络连接失败
**症状**: 浏览器控制台显示 "Network Error" 或 "Failed to fetch"

**原因**:
- 后端服务未运行
- 端口被占用
- 防火墙阻止

**解决方案**:
```bash
# 1. 检查端口是否被占用
lsof -i :8081

# 2. 检查服务是否运行
ps aux | grep vpanel

# 3. 检查服务日志
tail -f logs/app.log

# 4. 尝试直接访问 API
curl http://localhost:8081/api/health
```

## 完整的修复流程

### 步骤 1: 停止所有服务
```bash
# 停止后端
pkill vpanel

# 停止前端（如果在运行）
# 在前端目录中按 Ctrl+C
```

### 步骤 2: 清理和重新编译
```bash
# 重新编译后端
go build -o vpanel cmd/agent/main.go

# 清理前端缓存（如果需要）
cd web
rm -rf node_modules/.vite
cd ..
```

### 步骤 3: 启动后端
```bash
# 启动后端
./vpanel

# 等待服务启动
sleep 5

# 检查日志
tail -n 50 logs/app.log
```

### 步骤 4: 启动前端
```bash
cd web
npm run dev
```

### 步骤 5: 测试
1. 打开浏览器访问 `http://localhost:5173` (或前端端口)
2. 登录管理后台
3. 访问 IP 限制管理页面
4. 打开开发者工具查看 Network 标签

## 调试技巧

### 1. 启用详细日志
修改 `configs/config.yaml`:
```yaml
log:
  level: debug  # 改为 debug
```

### 2. 查看实时日志
```bash
tail -f logs/app.log | grep -i "ip"
```

### 3. 使用浏览器开发者工具
- **Network 标签**: 查看所有 API 请求和响应
- **Console 标签**: 查看 JavaScript 错误
- **Application 标签**: 查看 localStorage 中的 token

### 4. 测试数据库连接
```bash
# PostgreSQL
PGPASSWORD=vpanel123 psql -h localhost -U vpanel -d vpanel -c "SELECT 1;"

# MySQL
mysql -u root -p vpanel -e "SELECT 1;"
```

## 预防措施

### 1. 定期备份数据库
```bash
# PostgreSQL
pg_dump -h localhost -U vpanel vpanel > backup.sql

# MySQL
mysqldump -u root -p vpanel > backup.sql
```

### 2. 监控服务状态
创建一个监控脚本 `monitor.sh`:
```bash
#!/bin/bash
while true; do
  if ! pgrep -x "vpanel" > /dev/null; then
    echo "$(date): Service down, restarting..."
    ./vpanel &
  fi
  sleep 60
done
```

### 3. 设置日志轮转
确保 `configs/config.yaml` 中配置了日志轮转:
```yaml
log:
  max_size: 100    # MB
  max_backups: 10  # 保留文件数
  max_age: 30      # 天
```

## 获取帮助

如果以上步骤都无法解决问题，请收集以下信息：

1. **错误截图**: 包括浏览器控制台和 Network 标签
2. **服务日志**: `logs/app.log` 的最后 100 行
3. **配置文件**: `configs/config.yaml`
4. **系统信息**: 操作系统、数据库版本
5. **API 响应**: 使用 curl 测试的完整响应

```bash
# 收集诊断信息
echo "=== 系统信息 ===" > diagnosis.txt
uname -a >> diagnosis.txt
echo "" >> diagnosis.txt

echo "=== 服务状态 ===" >> diagnosis.txt
ps aux | grep vpanel >> diagnosis.txt
echo "" >> diagnosis.txt

echo "=== 最近日志 ===" >> diagnosis.txt
tail -n 100 logs/app.log >> diagnosis.txt
echo "" >> diagnosis.txt

echo "=== 配置文件 ===" >> diagnosis.txt
cat configs/config.yaml >> diagnosis.txt
```

## 相关文档

- [错误修复总结](./error-fix-summary.md)
- [IP 限制修复总结](./ip-restriction-fix-summary.md)
- [最终修复总结](./FINAL-FIX-SUMMARY.md)
- [部署检查清单](./deployment-checklist.md)
