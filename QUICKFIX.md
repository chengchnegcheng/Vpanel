# 🚨 快速修复指南

## 问题
API 返回 503/500 错误：
- `/api/admin/ip-whitelist` - 503 Service Unavailable
- `/api/admin/ip-blacklist` - 500 Internal Server Error

## 原因
数据库迁移未执行，IP 限制相关的表不存在。

## 快速修复（3 步）

### 方案 A: 自动修复（推荐）⭐

```bash
# 1. 停止服务
./vpanel.sh
# 选择 "1) Docker 部署管理" -> "2) 停止服务"

# 2. 重新启动（会自动执行迁移）
./vpanel.sh
# 选择 "1) Docker 部署管理" -> "1) 启动服务"

# 3. 验证
./scripts/test-api.sh
```

### 方案 B: 手动修复（快速）⚡

```bash
# 1. 执行迁移
./scripts/fix-migrations.sh

# 2. 重启服务
./vpanel.sh
# 选择 "1) Docker 部署管理" -> "3) 重启服务"

# 3. 验证
./scripts/check-db.sh
```

### 方案 C: Docker 重新部署 🐳

```bash
# 1. 重启服务
./deployments/scripts/start.sh restart

# 2. 验证
./scripts/test-api.sh
```

## 验证修复成功

### 1. 检查数据库表

```bash
./scripts/check-db.sh
```

应该看到以下表：
- ✅ ip_whitelist
- ✅ ip_blacklist
- ✅ active_ips
- ✅ ip_history

### 2. 测试 API

访问管理后台 -> IP 限制管理，应该能正常加载页面。

或使用命令行测试：
```bash
# 获取 admin token 后测试
./scripts/test-api.sh http://localhost:8080 YOUR_ADMIN_TOKEN
```

### 3. 检查日志

```bash
# 应用日志
tail -f logs/app.log | grep -i migration

# Docker 日志
docker logs v-panel | grep -i migration
```

应该看到类似输出：
```
Applied migration: 010_ip_restriction
```

## 如果还有问题

### 查看详细文档
```bash
cat Docs/api-database-fix.md
```

### 运行完整诊断
```bash
./scripts/check-db.sh
./scripts/test-api.sh
```

### 查看错误日志
```bash
# 应用日志
tail -100 logs/app.log

# Docker 日志
docker logs --tail 100 v-panel
```

### 备份和重置（最后手段）

⚠️ **警告：这会删除所有数据！**

```bash
# 1. 备份数据库
cp data/v.db data/v.db.backup.$(date +%Y%m%d_%H%M%S)

# 2. 停止服务
./vpanel.sh
# 选择停止服务

# 3. 删除数据库
rm data/v.db

# 4. 重新启动（会创建新数据库）
./vpanel.sh
# 选择启动服务
```

## 预防措施

### 定期备份
```bash
# 添加到 crontab
0 2 * * * cp /path/to/V/data/v.db /path/to/backups/v.db.$(date +\%Y\%m\%d)
```

### 监控日志
```bash
# 查看启动日志
tail -f logs/app.log | grep -E "migration|error|failed"
```

### 定期检查
```bash
# 每周运行一次
./scripts/check-db.sh
```

## 需要帮助？

1. 📖 查看完整文档：`Docs/api-database-fix.md`
2. 📊 查看检查报告：`Docs/deep-check-summary.md`
3. 🔧 使用诊断工具：`./scripts/check-db.sh`
4. 🐛 提交 Issue：GitHub Issues

---

**修复时间**: < 5 分钟  
**数据丢失**: 无  
**服务中断**: 最小化（仅重启时）
