# 部署检查清单

## 修复部署前检查

### ✅ 代码修改
- [x] 修改 `internal/api/handlers/ip_restriction.go`
- [x] 修改 `internal/api/handlers/report.go`
- [x] 修改 `internal/api/handlers/giftcard.go`
- [x] 代码语法检查通过（无编译错误）
- [x] 创建测试脚本 `scripts/test-api-fixes.sh`

### ✅ 文档创建
- [x] 错误修复指南 `Docs/error-fix-guide.md`
- [x] 修复总结 `Docs/error-fix-summary.md`
- [x] 快速参考 `Docs/quick-fix-reference.md`
- [x] 用户通知 `Docs/用户通知-错误修复.md`
- [x] 部署检查清单 `Docs/deployment-checklist.md`

## 部署步骤

### 1. 准备阶段
```bash
# 1.1 备份当前版本
[ ] cp agent agent.backup.$(date +%Y%m%d_%H%M%S)

# 1.2 备份配置文件
[ ] cp configs/config.yaml configs/config.yaml.backup.$(date +%Y%m%d_%H%M%S)

# 1.3 记录当前版本信息
[ ] ./agent --version > version.backup.txt
```

### 2. 编译阶段
```bash
# 2.1 清理旧的构建
[ ] go clean

# 2.2 下载依赖
[ ] go mod download

# 2.3 编译新版本
[ ] go build -o agent cmd/agent/main.go

# 2.4 验证编译成功
[ ] ./agent --version
```

### 3. 测试阶段（可选，在测试环境）
```bash
# 3.1 启动测试服务
[ ] ./agent --config configs/config.yaml.test

# 3.2 运行 API 测试
[ ] export ADMIN_TOKEN="test_token"
[ ] ./scripts/test-api-fixes.sh

# 3.3 检查测试结果
[ ] 所有测试通过
```

### 4. 部署阶段
```bash
# 4.1 停止当前服务
[ ] systemctl stop vpanel
# 或
[ ] ./vpanel.sh stop

# 4.2 替换二进制文件
[ ] cp agent /path/to/production/agent

# 4.3 验证文件权限
[ ] chmod +x /path/to/production/agent

# 4.4 启动服务
[ ] systemctl start vpanel
# 或
[ ] ./vpanel.sh start

# 4.5 检查服务状态
[ ] systemctl status vpanel
# 或
[ ] ./vpanel.sh status
```

### 5. 验证阶段
```bash
# 5.1 检查服务是否运行
[ ] curl http://localhost:8080/health

# 5.2 检查日志
[ ] tail -f /var/log/vpanel/app.log

# 5.3 测试修复的 API
[ ] export ADMIN_TOKEN="production_token"
[ ] ./scripts/test-api-fixes.sh

# 5.4 验证前端功能
[ ] 访问 IP 限制管理页面
[ ] 访问财务报表页面
[ ] 访问礼品卡管理页面
```

### 6. 监控阶段
```bash
# 6.1 监控错误日志（持续 10 分钟）
[ ] tail -f /var/log/vpanel/app.log | grep -i error

# 6.2 检查系统资源
[ ] top
[ ] free -h
[ ] df -h

# 6.3 检查数据库连接
[ ] mysql -u root -p -e "SHOW PROCESSLIST;"
```

## 回滚计划

如果部署后出现问题，按以下步骤回滚：

### 快速回滚
```bash
# 1. 停止服务
[ ] systemctl stop vpanel

# 2. 恢复备份
[ ] cp agent.backup.YYYYMMDD_HHMMSS agent

# 3. 启动服务
[ ] systemctl start vpanel

# 4. 验证服务
[ ] curl http://localhost:8080/health
```

### 完整回滚
```bash
# 1. 停止服务
[ ] systemctl stop vpanel

# 2. 恢复所有备份
[ ] cp agent.backup.YYYYMMDD_HHMMSS agent
[ ] cp configs/config.yaml.backup.YYYYMMDD_HHMMSS configs/config.yaml

# 3. 清理缓存
[ ] rm -rf /tmp/vpanel-cache/*

# 4. 启动服务
[ ] systemctl start vpanel

# 5. 验证服务
[ ] curl http://localhost:8080/health
[ ] tail -f /var/log/vpanel/app.log
```

## 问题排查

### 服务无法启动
```bash
# 检查端口占用
[ ] netstat -tlnp | grep 8080

# 检查配置文件
[ ] ./agent --config configs/config.yaml --validate

# 检查日志
[ ] journalctl -u vpanel -n 100

# 检查文件权限
[ ] ls -la agent
```

### API 返回错误
```bash
# 检查数据库连接
[ ] mysql -u root -p -e "SELECT 1;"

# 检查表结构
[ ] mysql -u root -p vpanel -e "SHOW TABLES;"

# 检查 IP 服务
[ ] curl -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/admin/ip-restrictions/stats
```

### 性能问题
```bash
# 检查慢查询
[ ] mysql -u root -p -e "SHOW FULL PROCESSLIST;"

# 检查系统负载
[ ] uptime
[ ] iostat

# 检查内存使用
[ ] free -h
[ ] ps aux | grep agent
```

## 通知清单

### 部署前通知
- [ ] 通知运维团队部署时间
- [ ] 通知用户可能的短暂服务中断
- [ ] 准备回滚计划

### 部署后通知
- [ ] 通知运维团队部署完成
- [ ] 通知用户错误已修复
- [ ] 发送用户通知文档

## 文档更新

- [ ] 更新 CHANGELOG.md
- [ ] 更新版本号
- [ ] 更新 API 文档（如有变化）
- [ ] 归档本次修复文档

## 后续跟进

### 第一天
- [ ] 监控错误日志
- [ ] 收集用户反馈
- [ ] 检查性能指标

### 第一周
- [ ] 分析错误趋势
- [ ] 优化性能瓶颈
- [ ] 更新监控告警

### 第一个月
- [ ] 评估修复效果
- [ ] 总结经验教训
- [ ] 改进开发流程

## 签名确认

| 角色 | 姓名 | 签名 | 日期 |
|------|------|------|------|
| 开发人员 | | | |
| 测试人员 | | | |
| 运维人员 | | | |
| 项目经理 | | | |

## 备注

- 部署时间建议：非高峰时段（如凌晨 2-4 点）
- 预计停机时间：5-10 分钟
- 风险等级：低（仅修复错误处理，不涉及核心逻辑）
- 回滚时间：< 5 分钟

---

**重要提示**：
1. 务必在测试环境验证后再部署到生产环境
2. 保留至少 3 个版本的备份
3. 部署过程中保持通讯畅通
4. 遇到问题立即回滚，不要尝试现场修复
