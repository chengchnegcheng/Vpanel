# IP限制功能数据库表修复

## 问题描述

在访问以下API端点时出现数据库错误：
- `/api/admin/ip-whitelist` - IP白名单
- `/api/admin/ip-blacklist` - IP黑名单  
- `/api/admin/reports/failed-payments` - 失败支付统计

错误信息：
```
SQL logic error: no such table: ip_blacklist (1)
SQL logic error: no such table: ip_whitelist (1)
```

## 根本原因

Docker容器 `v-panel` 中的SQLite数据库缺少IP限制相关的表：
- `ip_whitelist` - IP白名单表
- `ip_blacklist` - IP黑名单表
- `active_ips` - 活跃IP表
- `ip_history` - IP历史记录表
- `subscription_ip_access` - 订阅IP访问表
- `geo_cache` - 地理位置缓存表
- `failed_attempts` - 失败尝试表

## 解决方案

### 1. 创建缺失的数据库表

已执行SQL脚本 `fix_ip_tables.sql` 在Docker容器的SQLite数据库中创建所有缺失的表。

### 2. 重启Docker容器

```bash
docker restart v-panel
```

### 3. 验证修复

检查表是否已创建：
```bash
docker exec v-panel sqlite3 /app/data/v.db ".tables"
```

## 关于失败支付统计端点

`/api/admin/reports/failed-payments` 端点返回 503 错误是因为支付重试服务（RetryService）未初始化。这是一个可选功能，需要在应用启动时初始化该服务。

当前行为：
- 如果 `retryService` 为 nil，返回 `{"error": "Retry service not available"}`
- 这是预期行为，不是bug

如需启用此功能，需要：
1. 在 `internal/api/routes.go` 中初始化 `payment.RetryService`
2. 使用 `NewPaymentHandlerWithRetry` 而不是 `NewPaymentHandler`

## 端口配置

### Docker容器
- 运行在端口 8080
- 使用 SQLite 数据库：`/app/data/v.db`

### 本地开发服务器
- 运行在端口 8081（避免与Docker冲突）
- 使用 PostgreSQL 数据库：`localhost:5432`

### 前端开发服务器
- Vite dev server: 端口 5173
- API代理配置指向：`http://127.0.0.1:8081`

## 修复后的状态

✅ IP白名单API正常工作
✅ IP黑名单API正常工作  
✅ 所有IP限制相关表已创建
⚠️ 失败支付统计需要初始化RetryService（可选功能）

## 相关文件

- `fix_ip_tables.sql` - 数据库表创建脚本
- `configs/config.yaml` - 应用配置（端口已改为8081）
- `internal/ip/models.go` - IP限制数据模型
- `internal/ip/validator.go` - IP验证服务
- `internal/api/handlers/ip_restriction.go` - IP限制API处理器
- `internal/api/handlers/payment.go` - 支付API处理器

## 测试

访问以下URL测试（需要登录）：
- http://localhost:8080/admin/ip-restriction
- http://localhost:8080/admin/ip-whitelist  
- http://localhost:8080/admin/ip-blacklist

API端点应该返回空数组而不是错误：
```json
{
  "code": 200,
  "message": "success",
  "data": []
}
```
