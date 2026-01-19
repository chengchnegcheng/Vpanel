# 统计数据修复 - 快速参考

## 问题总结

用户门户统计数据加载失败，原因：
1. ❌ 类型转换错误：SQL返回字符串，但Go结构体期望time.Time
2. ❌ 数据泄露：没有按用户ID过滤，所有用户看到相同数据

## 修复内容

### 修改的文件

1. `internal/database/repository/repository.go`
   - 添加 `GetTrafficTimelineByUser` 方法到接口

2. `internal/database/repository/traffic_repository.go`
   - 修复 `GetTrafficTimeline` 的类型转换
   - 添加 `GetTrafficTimelineByUser` 实现（带用户过滤）

3. `internal/portal/stats/service.go`
   - 修改 `GetDailyTraffic` 使用新的过滤方法

### 新增的文件

1. `scripts/test-stats-api.sh` - API测试脚本
2. `scripts/insert-test-traffic-data.sql` - 测试数据生成脚本
3. `Docs/stats-loading-fix.md` - 详细修复文档
4. `Docs/stats-fix-quick-reference.md` - 本文件

## 快速测试

### 1. 编译

```bash
go build -o vpanel ./cmd/v
```

### 2. 插入测试数据

```bash
psql -U postgres -d vpanel -f scripts/insert-test-traffic-data.sql
```

### 3. 启动服务

```bash
./vpanel
```

### 4. 测试API

```bash
# 获取token（登录）
TOKEN=$(curl -s -X POST http://localhost:8080/api/portal/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password"}' | jq -r '.token')

# 测试统计API
./scripts/test-stats-api.sh $TOKEN
```

### 5. 测试前端

访问: http://localhost:8080/user/stats

## 验证清单

- [ ] 编译成功，无错误
- [ ] API返回数据，无500错误
- [ ] 返回的数据只包含当前用户的流量
- [ ] 日期格式正确（YYYY-MM-DD）
- [ ] 流量数值合理
- [ ] 前端图表正常渲染
- [ ] 不同用户看到不同的数据
- [ ] 时间范围切换正常工作

## 关键API端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/portal/stats/traffic` | GET | 获取流量统计 |
| `/api/portal/stats/daily` | GET | 获取每日流量 |
| `/api/portal/stats/usage` | GET | 获取使用统计 |
| `/api/portal/dashboard` | GET | 获取仪表板数据 |
| `/api/portal/dashboard/traffic` | GET | 获取流量摘要 |

## 常见问题

### Q: API返回空数组？
A: 检查数据库是否有流量数据，使用SQL脚本插入测试数据

### Q: 时间格式错误？
A: 确保PostgreSQL的TO_CHAR格式与Go的time.Parse格式匹配

### Q: 看到其他用户的数据？
A: 检查是否使用了GetTrafficTimelineByUser而不是GetTrafficTimeline

### Q: 编译错误？
A: 运行 `go mod tidy` 确保依赖正确

## 回滚方案

如果修复导致问题，可以回滚：

```bash
git checkout HEAD -- internal/database/repository/
git checkout HEAD -- internal/portal/stats/
```

## 联系支持

如果问题持续存在，请提供：
1. 错误日志
2. API响应
3. 数据库查询结果
4. 浏览器控制台错误
