# 用户门户统计数据加载失败 - 修复说明

## 问题描述

用户门户加载统计数据失败，无法显示流量使用情况和历史记录。

## 问题原因

发现了两个关键问题：

### 1. 类型不匹配问题

在 `internal/database/repository/traffic_repository.go` 文件的 `GetTrafficTimeline` 方法中：

- **SQL 查询问题**：使用 `TO_CHAR()` 函数将时间转换为字符串
- **结构体类型问题**：`TrafficTimelinePoint` 结构体的 `Time` 字段是 `time.Time` 类型
- **扫描失败**：GORM 无法将字符串扫描到 `time.Time` 类型，导致查询失败

### 2. 数据过滤问题（严重安全问题）

`GetTrafficTimeline` 方法没有按用户ID过滤数据，导致：

- **数据泄露**：用户可以看到所有用户的流量数据
- **统计错误**：显示的是全局流量而不是用户个人流量
- **安全隐患**：违反了数据隔离原则

## 修复方案

### 修复1：类型转换问题

修改 `GetTrafficTimeline` 方法，使用临时结构体接收字符串时间，然后转换为 `time.Time` 类型：

1. 创建临时结构体 `tempResult` 来接收 SQL 查询返回的字符串时间
2. 先将数据扫描到临时结构体
3. 根据不同的时间间隔（hour/day/month）使用相应的格式解析时间字符串
4. 将解析后的数据转换为 `TrafficTimelinePoint` 结构体

### 修复2：添加用户过滤

1. 在 `TrafficRepository` 接口中添加新方法 `GetTrafficTimelineByUser`
2. 实现该方法，在 SQL 查询中添加 `user_id` 过滤条件
3. 修改 `stats.Service` 的 `GetDailyTraffic` 方法，使用新的过滤方法

## 代码变更

### 文件1: `internal/database/repository/repository.go`

**修改**: 在 `TrafficRepository` 接口中添加新方法

```go
GetTrafficTimelineByUser(ctx context.Context, userID int64, start, end time.Time, interval string) ([]*TrafficTimelinePoint, error)
```

### 文件2: `internal/database/repository/traffic_repository.go`

**修改1**: 修复 `GetTrafficTimeline` 方法的类型转换问题
**修改2**: 添加 `GetTrafficTimelineByUser` 方法实现

主要改进：
- 添加临时结构体处理字符串时间
- 根据 interval 参数选择正确的时间解析格式
- 增加错误处理，解析失败时使用零值时间
- 新方法添加 `user_id` 过滤条件

### 文件3: `internal/portal/stats/service.go`

**修改**: `GetDailyTraffic` 方法

将 `GetTrafficTimeline` 调用改为 `GetTrafficTimelineByUser`，确保只返回当前用户的数据。

## 影响范围

此修复影响以下功能：

### 后端 API
- `GET /portal/stats/traffic` - 获取流量统计
- `GET /portal/stats/daily` - 获取每日流量数据
- `GET /portal/dashboard` - 获取仪表板数据

### 前端页面
- `/user/dashboard` - 用户仪表板（流量概览）
- `/user/stats` - 统计详情页面（流量图表和详细记录）

### 服务层
- `internal/portal/stats/service.go` 中的 `GetDailyTraffic` 方法
- `internal/portal/stats/service.go` 中的 `GetTrafficStats` 方法

## 测试建议

### 1. 准备测试数据

如果数据库中没有流量数据，可以使用提供的 SQL 脚本插入测试数据：

```bash
# 连接到数据库并执行脚本
psql -U your_user -d your_database -f scripts/insert-test-traffic-data.sql
```

这将为用户ID=1插入最近30天的随机流量数据。

### 2. 后端测试

#### 方式1：使用测试脚本

```bash
# 首先启动服务
./vpanel

# 在另一个终端，使用测试脚本
# 需要先获取认证token
./scripts/test-stats-api.sh <your_auth_token>
```

#### 方式2：手动测试 API 端点

```bash
# 设置token变量
TOKEN="your_auth_token_here"

# 测试流量统计（周）
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/portal/stats/traffic?period=week"

# 测试每日流量
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/portal/stats/daily?days=7"

# 测试使用统计
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/portal/stats/usage"

# 测试仪表板
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/portal/dashboard"
```

### 3. 前端测试

1. 登录用户门户
2. 访问仪表板页面 (`/user/dashboard`)，检查：
   - 流量使用卡片是否正常显示
   - 流量百分比是否正确
   - 流量数值格式是否正确
3. 访问统计页面 (`/user/stats`)，检查：
   - 统计概览卡片（上传/下载/总流量）
   - 流量趋势图表（折线图/柱状图切换）
   - 节点使用排行
   - 协议使用分布
   - 详细记录表格
4. 测试时间范围切换：
   - 今日/本周/本月/本年
   - 自定义日期范围
5. 测试数据导出功能

### 4. 数据验证

确保以下数据正确显示：
- ✅ 流量数值准确（与数据库一致）
- ✅ 日期格式正确（YYYY-MM-DD）
- ✅ 图表渲染正常（无错误）
- ✅ 无控制台错误
- ✅ 只显示当前用户的数据（数据隔离）
- ✅ 不同时间范围的数据正确

### 5. 安全验证

**重要**：验证数据隔离是否正确

1. 使用用户A的token访问统计API
2. 记录返回的流量数据
3. 使用用户B的token访问统计API
4. 确认两个用户看到的数据不同
5. 直接查询数据库验证数据正确性

```sql
-- 验证用户A的流量
SELECT SUM(upload), SUM(download) 
FROM traffic 
WHERE user_id = 1;

-- 验证用户B的流量
SELECT SUM(upload), SUM(download) 
FROM traffic 
WHERE user_id = 2;
```

## 相关文件

### 后端
- `internal/database/repository/traffic_repository.go` - 数据库查询层（已修复）
- `internal/database/repository/repository.go` - 仓库接口定义
- `internal/portal/stats/service.go` - 统计服务层
- `internal/api/handlers/portal_stats.go` - API 处理器
- `internal/api/handlers/portal_dashboard.go` - 仪表板处理器
- `internal/api/routes.go` - 路由配置

### 前端
- `web/src/views/user/Dashboard.vue` - 用户仪表板组件
- `web/src/views/user/Stats.vue` - 统计页面组件
- `web/src/stores/portalStats.js` - 统计数据 Store
- `web/src/api/modules/portal/stats.js` - API 调用模块

## 注意事项

1. **数据库兼容性**：当前实现使用 PostgreSQL 的 `TO_CHAR` 函数，如果使用其他数据库需要相应调整
2. **时区处理**：确保服务器和数据库的时区设置一致
3. **性能考虑**：对于大量数据，考虑添加索引或分页
4. **错误处理**：前端应该有适当的错误提示和加载状态

## 后续优化建议

1. **缓存机制**：对于频繁访问的统计数据，可以添加缓存层
2. **实时更新**：考虑使用 WebSocket 实现流量数据的实时更新
3. **数据聚合**：对于历史数据，可以预先聚合以提高查询性能
4. **监控告警**：添加流量使用告警功能
