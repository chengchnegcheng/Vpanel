# ✅ 修复和验证检查清单

## 📋 修复前检查

- [ ] 备份当前数据库
  ```bash
  cp data/v.db data/v.db.backup.$(date +%Y%m%d_%H%M%S)
  ```

- [ ] 记录当前服务状态
  ```bash
  ./scripts/test-api.sh > before-fix.log 2>&1
  ```

- [ ] 检查当前数据库状态
  ```bash
  ./scripts/check-db.sh > db-before.log 2>&1
  ```

## 🔧 应用修复

选择一种方式：

### 方式 A: 自动修复（推荐）
- [ ] 停止服务
  ```bash
  ./vpanel.sh
  # 选择 "1) Docker 部署管理" -> "2) 停止服务"
  ```

- [ ] 重新启动服务（自动执行迁移）
  ```bash
  ./vpanel.sh
  # 选择 "1) Docker 部署管理" -> "1) 启动服务"
  ```

### 方式 B: 手动迁移
- [ ] 执行迁移脚本
  ```bash
  ./scripts/fix-migrations.sh
  ```

- [ ] 重启服务
  ```bash
  ./vpanel.sh
  # 选择 "1) Docker 部署管理" -> "3) 重启服务"
  ```

### 方式 C: Docker 重新部署
- [ ] 重启 Docker 服务
  ```bash
  ./deployments/scripts/start.sh restart
  ```

## ✅ 验证修复

### 1. 数据库验证
- [ ] 运行数据库检查
  ```bash
  ./scripts/check-db.sh
  ```

- [ ] 确认以下表存在：
  - [ ] ip_whitelist
  - [ ] ip_blacklist
  - [ ] active_ips
  - [ ] ip_history
  - [ ] subscription_ip_access
  - [ ] geo_cache
  - [ ] failed_attempts

- [ ] 确认迁移记录存在
  ```bash
  sqlite3 data/v.db "SELECT version, name FROM migrations WHERE version='010';"
  ```
  应该看到：`010|ip_restriction`

### 2. API 验证
- [ ] 测试健康检查
  ```bash
  curl http://localhost:8080/health
  ```
  应该返回：`{"status":"ok"}`

- [ ] 获取 admin token
  ```bash
  # 登录获取 token
  curl -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}'
  ```

- [ ] 测试 IP 白名单 API
  ```bash
  curl -H "Authorization: Bearer YOUR_TOKEN" \
    http://localhost:8080/api/admin/ip-whitelist
  ```
  应该返回：`{"code":200,"message":"success","data":[]}`

- [ ] 测试 IP 黑名单 API
  ```bash
  curl -H "Authorization: Bearer YOUR_TOKEN" \
    http://localhost:8080/api/admin/ip-blacklist
  ```
  应该返回：`{"code":200,"message":"success","data":[]}`

- [ ] 运行完整 API 测试
  ```bash
  ./scripts/test-api.sh http://localhost:8080 YOUR_TOKEN
  ```

### 3. 前端验证
- [ ] 登录管理后台
  - 访问：`http://localhost:8080/admin/`
  - 用户名：`admin`
  - 密码：`admin123`

- [ ] 访问 IP 限制管理页面
  - 应该能正常加载
  - 不应该有 503 或 500 错误

- [ ] 测试白名单功能
  - [ ] 查看白名单列表（应该为空）
  - [ ] 添加一条白名单记录
  - [ ] 编辑白名单记录
  - [ ] 删除白名单记录

- [ ] 测试黑名单功能
  - [ ] 查看黑名单列表（应该为空）
  - [ ] 添加一条黑名单记录
  - [ ] 编辑黑名单记录
  - [ ] 删除黑名单记录

### 4. 日志验证
- [ ] 检查应用日志
  ```bash
  tail -100 logs/app.log | grep -i "migration\|error"
  ```

- [ ] 确认没有错误日志
  ```bash
  grep -i "error\|failed" logs/app.log | tail -20
  ```

- [ ] Docker 日志检查（如果使用 Docker）
  ```bash
  docker logs v-panel | grep -i "migration\|error" | tail -20
  ```

## 🧪 功能测试

### IP 限制功能
- [ ] 添加 IP 到白名单
- [ ] 添加 IP 到黑名单
- [ ] 查看在线设备
- [ ] 踢出设备
- [ ] 导入批量 IP
- [ ] 查看 IP 统计

### 其他核心功能（回归测试）
- [ ] 用户登录/登出
- [ ] 用户管理 CRUD
- [ ] 代理管理 CRUD
- [ ] 订阅链接生成
- [ ] 节点管理
- [ ] 套餐管理
- [ ] 订单创建

## 📊 性能检查

- [ ] 检查数据库大小
  ```bash
  du -h data/v.db
  ```

- [ ] 检查响应时间
  ```bash
  time curl http://localhost:8080/health
  ```

- [ ] 检查内存使用
  ```bash
  docker stats v-panel --no-stream  # Docker
  # 或
  ps aux | grep v  # 本地运行
  ```

## 📝 文档检查

- [ ] 阅读快速修复指南
  ```bash
  cat QUICKFIX.md
  ```

- [ ] 阅读详细修复文档
  ```bash
  cat Docs/api-database-fix.md
  ```

- [ ] 阅读检查报告
  ```bash
  cat Docs/deep-check-summary.md
  ```

## 🔄 清理和备份

- [ ] 删除旧的备份（保留最近 7 天）
  ```bash
  find data -name "v.db.backup.*" -mtime +7 -delete
  ```

- [ ] 创建修复后的备份
  ```bash
  cp data/v.db data/v.db.fixed.$(date +%Y%m%d_%H%M%S)
  ```

- [ ] 保存测试日志
  ```bash
  ./scripts/test-api.sh > after-fix.log 2>&1
  ./scripts/check-db.sh > db-after.log 2>&1
  ```

## 📈 监控设置

- [ ] 设置日志监控
  ```bash
  # 添加到 crontab 或监控系统
  tail -f logs/app.log | grep -i "error\|failed"
  ```

- [ ] 设置数据库备份
  ```bash
  # 添加到 crontab
  0 2 * * * cp /path/to/V/data/v.db /path/to/backups/v.db.$(date +\%Y\%m\%d)
  ```

- [ ] 设置健康检查
  ```bash
  # 添加到监控系统
  */5 * * * * curl -f http://localhost:8080/health || alert
  ```

## 🎯 完成标准

### 必须满足（Critical）
- [x] 代码编译通过
- [ ] 数据库迁移成功
- [ ] IP 白名单 API 返回 200
- [ ] IP 黑名单 API 返回 200
- [ ] 前端页面正常加载
- [ ] 无错误日志

### 应该满足（Important）
- [ ] 所有 API 测试通过
- [ ] 前端功能测试通过
- [ ] 数据库完整性检查通过
- [ ] 文档已更新

### 可以满足（Nice to have）
- [ ] 性能测试通过
- [ ] 监控已设置
- [ ] 备份策略已实施

## 🚨 回滚计划

如果修复失败，执行以下步骤：

1. [ ] 停止服务
   ```bash
   ./vpanel.sh
   # 选择停止服务
   ```

2. [ ] 恢复数据库备份
   ```bash
   cp data/v.db.backup.YYYYMMDD_HHMMSS data/v.db
   ```

3. [ ] 恢复代码（如果修改了）
   ```bash
   git checkout internal/database/db.go
   ```

4. [ ] 重新启动服务
   ```bash
   ./vpanel.sh
   # 选择启动服务
   ```

5. [ ] 报告问题
   - 保存错误日志
   - 保存测试结果
   - 提交 GitHub Issue

## 📞 获取帮助

如果遇到问题：

1. 📖 查看文档
   - `QUICKFIX.md` - 快速修复
   - `Docs/api-database-fix.md` - 详细指南
   - `Docs/deep-check-summary.md` - 检查报告

2. 🔍 运行诊断
   ```bash
   ./scripts/check-db.sh
   ./scripts/test-api.sh
   ```

3. 📝 查看日志
   ```bash
   tail -100 logs/app.log
   docker logs --tail 100 v-panel
   ```

4. 🐛 提交 Issue
   - 包含错误日志
   - 包含诊断结果
   - 描述复现步骤

---

## 检查清单状态

- **修复前检查**: ⬜ 未开始 / 🟡 进行中 / ✅ 完成
- **应用修复**: ⬜ 未开始 / 🟡 进行中 / ✅ 完成
- **验证修复**: ⬜ 未开始 / 🟡 进行中 / ✅ 完成
- **功能测试**: ⬜ 未开始 / 🟡 进行中 / ✅ 完成
- **性能检查**: ⬜ 未开始 / 🟡 进行中 / ✅ 完成
- **清理备份**: ⬜ 未开始 / 🟡 进行中 / ✅ 完成

**总体状态**: ⬜ 未开始

**完成时间**: ___________

**执行人**: ___________

**备注**: ___________
