# 开发者备忘录：统计功能修复

## 技术细节

### 问题根因分析

#### 问题1：GORM类型扫描失败

**代码位置**: `internal/database/repository/traffic_repository.go:182`

**问题代码**:
```go
selectClause := "TO_CHAR(recorded_at, 'YYYY-MM-DD') as time, ..."
err := r.db.Scan(&results).Error  // results是[]*TrafficTimelinePoint
```

**TrafficTimelinePoint结构**:
```go
type TrafficTimelinePoint struct {
    Time     time.Time  // 期望time.Time类型
    Upload   int64
    Download int64
}
```

**失败原因**:
- SQL返回: `string` (通过TO_CHAR转换)
- 结构体期望: `time.Time`
- GORM无法自动转换字符串到time.Time

**解决方案**:
```go
// 1. 使用临时结构体接收字符串
type tempResult struct {
    TimeStr  string `gorm:"column:time"`
    Upload   int64
    Download int64
}

// 2. 扫描到临时结构体
var tempResults []*tempResult
err := r.db.Scan(&tempResults).Error

// 3. 手动解析时间字符串
for _, temp := range tempResults {
    parsedTime, _ := time.Parse("2006-01-02", temp.TimeStr)
    results = append(results, &TrafficTimelinePoint{
        Time: parsedTime,
        ...
    })
}
```

#### 问题2：缺少用户ID过滤

**代码位置**: `internal/portal/stats/service.go:86`

**问题代码**:
```go
timeline, err := s.trafficRepo.GetTrafficTimeline(ctx, start, end, "day")
// 查询所有用户的数据，没有WHERE user_id = ?
```

**安全影响**:
- 用户A可以看到用户B的流量数据
- 违反数据隔离原则
- 潜在的隐私泄露

**解决方案**:
```go
// 1. 添加新方法到接口
GetTrafficTimelineByUser(ctx, userID, start, end, interval)

// 2. 实现带用户过滤的查询
Where("user_id = ? AND recorded_at BETWEEN ? AND ?", userID, start, end)

// 3. 服务层使用新方法
timeline, err := s.trafficRepo.GetTrafficTimelineByUser(ctx, userID, start, end, "day")
```

### 架构设计

```
┌─────────────────┐
│   Frontend      │
│  (Vue.js)       │
└────────┬────────┘
         │ HTTP Request
         ▼
┌─────────────────┐
│   API Handler   │
│  portal_stats   │
└────────┬────────┘
         │ Call Service
         ▼
┌─────────────────┐
│  Stats Service  │
│  (Business)     │
└────────┬────────┘
         │ Call Repository
         ▼
┌─────────────────┐
│   Repository    │
│  (Data Access)  │
└────────┬────────┘
         │ SQL Query
         ▼
┌─────────────────┐
│   PostgreSQL    │
│   (Database)    │
└─────────────────┘
```

### 数据流

```
1. 用户请求: GET /api/portal/stats/traffic?period=week
   ↓
2. Handler: GetTrafficStats(c *gin.Context)
   - 提取user_id from context
   - 解析period参数
   ↓
3. Service: GetDailyTraffic(ctx, userID, days)
   - 计算时间范围
   ↓
4. Repository: GetTrafficTimelineByUser(ctx, userID, start, end, "day")
   - 执行SQL查询
   - 按user_id过滤
   - 按日期分组
   ↓
5. 数据转换:
   - SQL字符串 → time.Time
   - 聚合数据
   ↓
6. 返回JSON响应
```

### SQL查询优化

**原始查询**:
```sql
SELECT TO_CHAR(recorded_at, 'YYYY-MM-DD') as time,
       COALESCE(SUM(upload), 0) as upload,
       COALESCE(SUM(download), 0) as download
FROM traffic
WHERE recorded_at BETWEEN ? AND ?
GROUP BY TO_CHAR(recorded_at, 'YYYY-MM-DD')
ORDER BY time ASC
```

**优化后查询**:
```sql
SELECT TO_CHAR(recorded_at, 'YYYY-MM-DD') as time,
       COALESCE(SUM(upload), 0) as upload,
       COALESCE(SUM(download), 0) as download
FROM traffic
WHERE user_id = ?                          -- 添加用户过滤
  AND recorded_at BETWEEN ? AND ?
GROUP BY TO_CHAR(recorded_at, 'YYYY-MM-DD')
ORDER BY time ASC
```

**性能考虑**:
- 确保 `user_id` 列有索引
- 确保 `recorded_at` 列有索引
- 考虑复合索引: `(user_id, recorded_at)`

### 时间格式映射

| Interval | PostgreSQL Format | Go Parse Format |
|----------|-------------------|-----------------|
| hour | YYYY-MM-DD HH24:00:00 | 2006-01-02 15:00:00 |
| day | YYYY-MM-DD | 2006-01-02 |
| month | YYYY-MM | 2006-01 |

### 错误处理策略

```go
// 1. Repository层：返回错误
if err != nil {
    return nil, errors.NewDatabaseError("failed to get traffic timeline", err)
}

// 2. Service层：传播错误
if err != nil {
    return nil, err
}

// 3. Handler层：记录日志并返回HTTP错误
if err != nil {
    h.logger.Error("failed to get daily traffic", logger.F("error", err))
    c.JSON(http.StatusInternalServerError, gin.H{"error": "获取每日流量失败"})
    return
}
```

### 测试策略

#### 单元测试
```go
func TestGetTrafficTimelineByUser(t *testing.T) {
    // 1. 准备测试数据
    // 2. 调用方法
    // 3. 验证结果
    // 4. 验证SQL查询包含user_id过滤
}
```

#### 集成测试
```go
func TestStatsAPI(t *testing.T) {
    // 1. 创建测试用户
    // 2. 插入测试流量数据
    // 3. 调用API
    // 4. 验证响应
    // 5. 验证数据隔离
}
```

### 性能指标

| 操作 | 预期响应时间 | 数据量 |
|------|-------------|--------|
| 获取7天数据 | < 100ms | ~7条记录 |
| 获取30天数据 | < 200ms | ~30条记录 |
| 获取365天数据 | < 500ms | ~365条记录 |

### 监控建议

1. **API响应时间**
   - 监控 `/portal/stats/*` 端点
   - 设置告警阈值: > 1s

2. **错误率**
   - 监控500错误
   - 设置告警阈值: > 1%

3. **数据库查询**
   - 监控慢查询 (> 1s)
   - 检查索引使用情况

### 未来优化

1. **缓存层**
   ```go
   // Redis缓存统计数据
   key := fmt.Sprintf("stats:user:%d:period:%s", userID, period)
   if cached, err := redis.Get(key); err == nil {
       return cached
   }
   ```

2. **数据预聚合**
   ```sql
   -- 创建每日汇总表
   CREATE TABLE daily_traffic_summary (
       user_id BIGINT,
       date DATE,
       upload BIGINT,
       download BIGINT,
       PRIMARY KEY (user_id, date)
   );
   ```

3. **异步更新**
   - 使用消息队列处理流量记录
   - 定时任务更新汇总数据

### 相关资源

- GORM文档: https://gorm.io/docs/
- PostgreSQL日期函数: https://www.postgresql.org/docs/current/functions-datetime.html
- Go时间格式: https://golang.org/pkg/time/#Parse

### 代码审查清单

- [ ] 所有查询都包含用户ID过滤
- [ ] 时间解析处理所有interval类型
- [ ] 错误处理完整
- [ ] 日志记录充分
- [ ] 性能可接受
- [ ] 安全性验证
- [ ] 文档完整

### 部署清单

- [ ] 代码审查通过
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 性能测试通过
- [ ] 安全测试通过
- [ ] 文档更新
- [ ] 数据库索引检查
- [ ] 回滚方案准备

---

**作者**: Kiro AI
**日期**: 2026-01-19
**版本**: 1.0
