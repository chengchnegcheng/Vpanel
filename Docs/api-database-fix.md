# API 和数据库问题诊断与修复

## 问题描述

前端访问以下 API 时出现错误：
- `GET /api/admin/ip-whitelist` - 返回 503 (Service Unavailable)
- `GET /api/admin/ip-blacklist` - 返回 500 (Internal Server Error)

## 根本原因

**SQL 迁移文件从未被执行！**

项目中存在两种迁移方式：
1. **GORM AutoMigrate** - 用于 Go 结构体定义的模型
2. **SQL 迁移文件** - 位于 `internal/database/migrations/*.sql`

问题在于 `internal/database/db.go` 的 `AutoMigrate()` 方法只执行了 GORM 的自动迁移，但从未执行 SQL 迁移文件。这导致以下表未被创建：

- `ip_whitelist` - IP 白名单
- `ip_blacklist` - IP 黑名单
- `active_ips` - 活跃 IP
- `ip_history` - IP 历史记录
- `subscription_ip_access` - 订阅 IP 访问
- `geo_cache` - 地理位置缓存
- `failed_attempts` - 失败尝试记录

## 修复方案

### 方案 1: 代码修复（推荐，长期解决方案）

已修复 `internal/database/db.go` 文件，在 `AutoMigrate()` 方法中添加了 SQL 迁移的执行：

```go
func (d *Database) AutoMigrate() error {
    ctx := context.Background()
    
    // 首先执行 SQL 迁移
    migrator := migrations.NewMigrator(d.db)
    if err := migrator.Migrate(ctx); err != nil {
        return fmt.Errorf("failed to run SQL migrations: %w", err)
    }
    
    // 然后执行 GORM 自动迁移
    return d.db.AutoMigrate(...)
}
```

**使用方法：**
1. 停止服务
2. 重新编译并启动服务
3. 服务启动时会自动执行所有待处理的 SQL 迁移

### 方案 2: 手动修复（快速解决方案）

使用提供的脚本手动执行 SQL 迁移：

```bash
# 1. 检查数据库状态
./scripts/check-db.sh

# 2. 执行 SQL 迁移
./scripts/fix-migrations.sh

# 3. 再次检查确认
./scripts/check-db.sh
```

### 方案 3: 使用 Docker 重新部署

如果使用 Docker 部署，重新构建镜像会自动应用修复：

```bash
# 停止服务
./deployments/scripts/start.sh stop

# 清理旧数据（可选，会删除所有数据）
./deployments/scripts/start.sh clean

# 重新启动（会自动执行迁移）
./deployments/scripts/start.sh start
```

## 诊断工具

### 1. 数据库检查脚本

```bash
./scripts/check-db.sh [数据库路径]
```

**功能：**
- 列出所有数据库表
- 检查 IP 限制相关表是否存在
- 显示迁移状态
- 检查表结构和索引
- 验证数据库完整性

### 2. 迁移修复脚本

```bash
./scripts/fix-migrations.sh [数据库路径]
```

**功能：**
- 创建迁移表（如果不存在）
- 执行所有待处理的 SQL 迁移
- 记录迁移历史
- 显示迁移状态

## 验证修复

### 1. 检查数据库表

```bash
sqlite3 data/v.db ".tables"
```

应该能看到以下表：
- ip_whitelist
- ip_blacklist
- active_ips
- ip_history
- subscription_ip_access
- geo_cache
- failed_attempts

### 2. 检查迁移记录

```bash
sqlite3 data/v.db "SELECT version, name, applied_at FROM migrations ORDER BY version;"
```

应该能看到所有迁移记录，包括 `010_ip_restriction`。

### 3. 测试 API

启动服务后，访问以下 API：

```bash
# 获取 IP 白名单（需要 admin token）
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/admin/ip-whitelist

# 获取 IP 黑名单
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/admin/ip-blacklist
```

应该返回 200 状态码和空数组或数据列表。

### 4. 前端测试

1. 登录管理后台
2. 访问 "IP 限制管理" 页面
3. 应该能正常加载白名单和黑名单列表（即使是空的）

## 其他可能的问题

### 1. IP Service 初始化失败

如果 IP Service 初始化失败，会导致 `ipRestrictionHandler` 为 `nil`，访问相关路由会返回 404。

**检查日志：**
```bash
# Docker 部署
docker logs v-panel | grep "IP service"

# 本地运行
# 查看控制台输出
```

**可能的错误：**
- GeoIP 数据库文件不存在（这是正常的，服务会继续运行）
- 数据库连接失败

### 2. 权限问题

确保请求包含有效的 admin 令牌：

```javascript
// 前端应该自动添加
headers: {
  'Authorization': `Bearer ${token}`
}
```

### 3. CORS 问题

如果前端和后端在不同域名，确保 CORS 配置正确：

```yaml
# configs/config.yaml
server:
  cors_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
```

## 预防措施

### 1. 添加迁移测试

在 CI/CD 流程中添加迁移测试：

```bash
# 测试迁移是否能成功执行
go test ./internal/database/migrations/...
```

### 2. 数据库备份

定期备份数据库：

```bash
# 手动备份
cp data/v.db data/v.db.backup.$(date +%Y%m%d_%H%M%S)

# 或使用 SQLite 备份命令
sqlite3 data/v.db ".backup data/v.db.backup"
```

### 3. 监控日志

监控应用启动日志，确保迁移成功执行：

```bash
# 查找迁移相关日志
grep -i "migration" logs/app.log
```

## 相关文件

- `internal/database/db.go` - 数据库初始化和迁移
- `internal/database/migrations/` - SQL 迁移文件
- `internal/database/migrations/010_ip_restriction.sql` - IP 限制表定义
- `internal/api/handlers/ip_restriction.go` - IP 限制 API 处理器
- `internal/api/routes.go` - API 路由注册
- `internal/ip/service.go` - IP 限制服务
- `scripts/check-db.sh` - 数据库诊断脚本
- `scripts/fix-migrations.sh` - 迁移修复脚本

## 总结

这个问题的根本原因是数据库迁移系统不完整。项目同时使用了 GORM AutoMigrate 和 SQL 迁移文件，但只执行了前者。修复后，所有 SQL 迁移文件都会在服务启动时自动执行，确保数据库结构完整。

**建议：**
1. 使用方案 1（代码修复）作为长期解决方案
2. 如果需要立即修复现有数据库，使用方案 2（手动脚本）
3. 定期运行诊断脚本检查数据库健康状态
4. 在开发和生产环境都要确保迁移正确执行
