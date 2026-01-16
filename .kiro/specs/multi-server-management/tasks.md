# Implementation Plan: Multi-Server Management System

## Overview

本实现计划将多服务器管理系统分解为可执行的编码任务，按照数据库 → 核心服务 → API → 前端的顺序实现。使用 Go 语言实现后端，Vue 3 + Element Plus 实现前端。

## Tasks

### Phase 1: 数据库和基础设施

- [x] 1. 创建数据模型和数据库迁移
  - [x] 1.1 创建 Node 模型
    - 在 `internal/database/models.go` 添加 Node 结构体
    - 定义所有字段和 GORM 标签
    - _Requirements: 1.1, 1.2, 1.6_
  - [x] 1.2 创建 NodeGroup 和 NodeGroupMember 模型
    - 添加分组相关结构体
    - _Requirements: 6.1, 6.2, 6.3_
  - [x] 1.3 创建 HealthCheck 模型
    - 添加健康检查记录结构体
    - _Requirements: 2.6_
  - [x] 1.4 创建 UserNodeAssignment 模型
    - 添加用户-节点分配结构体
    - _Requirements: 4.7_
  - [x] 1.5 创建 NodeTraffic 模型
    - 添加节点流量统计结构体
    - _Requirements: 8.1, 8.2_
  - [x] 1.6 创建 NodeAuthFailure 模型
    - 添加认证失败记录结构体
    - _Requirements: 10.7_
  - [x] 1.7 创建数据库迁移文件
    - 在 `internal/database/migrations/` 创建迁移文件
    - 创建所有表和索引
    - _Requirements: 1.1-10.7_

- [x] 2. 创建 Repository 层
  - [x] 2.1 创建 NodeRepository
    - 在 `internal/database/repository/` 创建 `node_repository.go`
    - 实现 CRUD 和查询方法
    - _Requirements: 1.1-1.7_
  - [x] 2.2 创建 NodeGroupRepository
    - 创建 `node_group_repository.go`
    - 实现分组 CRUD 和成员管理方法
    - _Requirements: 6.1-6.6_
  - [x] 2.3 创建 HealthCheckRepository
    - 创建 `health_check_repository.go`
    - 实现健康检查记录的 CRUD 方法
    - _Requirements: 2.6_
  - [x] 2.4 创建 UserNodeAssignmentRepository
    - 创建 `user_node_assignment_repository.go`
    - 实现用户-节点分配的 CRUD 方法
    - _Requirements: 4.2, 4.7_
  - [x] 2.5 创建 NodeTrafficRepository
    - 创建 `node_traffic_repository.go`
    - 实现流量统计的 CRUD 和聚合方法
    - _Requirements: 8.1, 8.2_

- [x] 3. Checkpoint - 确保数据库层完成
  - 运行数据库迁移
  - 验证表结构
  - 如有问题请询问用户

### Phase 2: 核心服务实现

- [x] 4. 实现节点服务
  - [x] 4.1 创建 NodeService 基础结构
    - 在 `internal/node/` 创建 `service.go`
    - 实现节点 CRUD 操作
    - _Requirements: 1.1, 1.4, 1.5_
  - [x] 4.2 实现 Token 生成和管理
    - 实现 GenerateToken、RotateToken、RevokeToken、ValidateToken 方法
    - 使用加密安全随机数生成 Token
    - _Requirements: 1.2, 10.2, 10.3_
  - [x] 4.3 编写属性测试：Token 唯一性
    - **Property 1: Token Uniqueness**
    - **Validates: Requirements 1.2**
  - [x] 4.4 实现地址验证
    - 验证 IPv4、IPv6 和域名格式
    - _Requirements: 1.3_
  - [x] 4.5 编写属性测试：地址验证
    - **Property 2: Node Address Validation**
    - **Validates: Requirements 1.3**
  - [x] 4.6 实现节点删除和用户重分配
    - 删除节点时自动重分配用户
    - _Requirements: 1.5_
  - [x] 4.7 编写属性测试：用户重分配
    - **Property 3: User Reassignment on Node Deletion**
    - **Validates: Requirements 1.5**

- [x] 5. 实现健康检查服务
  - [x] 5.1 创建 HealthChecker 基础结构
    - 在 `internal/node/` 创建 `health_checker.go`
    - 实现定时检查调度
    - _Requirements: 2.1, 2.3_
  - [x] 5.2 实现健康检查逻辑
    - 检查 TCP 连接、API 响应、Xray 状态
    - 记录延迟
    - _Requirements: 2.2, 2.8_
  - [x] 5.3 实现状态转换逻辑
    - 实现 unhealthy/healthy 状态转换
    - 支持配置连续失败/成功阈值
    - _Requirements: 2.4, 2.5_
  - [x] 5.4 编写属性测试：状态转换
    - **Property 5: Health Status Transition**
    - **Validates: Requirements 2.4, 2.5**
  - [x] 5.5 实现通知触发
    - 节点状态变化时触发通知
    - _Requirements: 2.7_

- [x] 6. Checkpoint - 确保核心服务测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 3: 负载均衡和故障转移

- [x] 7. 实现负载均衡器
  - [x] 7.1 创建 LoadBalancer 基础结构
    - 在 `internal/node/` 创建 `load_balancer.go`
    - 定义策略接口
    - _Requirements: 4.1_
  - [x] 7.2 实现轮询策略
    - 实现 RoundRobinStrategy
    - _Requirements: 4.1_
  - [x] 7.3 实现最少连接策略
    - 实现 LeastConnectionsStrategy
    - _Requirements: 4.1_
  - [x] 7.4 实现加权策略
    - 实现 WeightedStrategy
    - _Requirements: 4.5_
  - [x] 7.5 编写属性测试：加权分布
    - **Property 8: Weighted Distribution**
    - **Validates: Requirements 4.5**
  - [x] 7.6 实现地理位置策略
    - 实现 GeographicStrategy
    - 集成 IP 地理位置查询
    - _Requirements: 4.6_
  - [x] 7.7 编写属性测试：地理位置选择
    - **Property 9: Geographic Selection**
    - **Validates: Requirements 4.6**
  - [x] 7.8 实现容量限制检查
    - 排除已满节点
    - _Requirements: 4.3, 4.4_
  - [x] 7.9 编写属性测试：容量限制
    - **Property 7: Capacity Limit Enforcement**
    - **Validates: Requirements 4.3, 4.4**
  - [x] 7.10 实现粘性会话
    - 保持用户-节点亲和性
    - _Requirements: 4.7_
  - [x] 7.11 编写属性测试：粘性会话
    - **Property 10: Sticky Session Consistency**
    - **Validates: Requirements 4.7**

- [x] 8. 实现故障转移管理器
  - [x] 8.1 创建 FailoverManager 基础结构
    - 在 `internal/node/` 创建 `failover_manager.go`
    - _Requirements: 5.1_
  - [x] 8.2 实现用户迁移逻辑
    - 将用户从故障节点迁移到健康节点
    - _Requirements: 5.1_
  - [x] 8.3 编写属性测试：故障转移迁移
    - **Property 11: Failover Migration**
    - **Validates: Requirements 5.1**
  - [x] 8.4 实现同组优先策略
    - 优先选择同组节点
    - _Requirements: 5.2_
  - [x] 8.5 编写属性测试：同组优先
    - **Property 12: Same-Group Failover Priority**
    - **Validates: Requirements 5.2**
  - [x] 8.6 实现并发迁移限制
    - 限制同时迁移的用户数
    - _Requirements: 5.6_
  - [x] 8.7 编写属性测试：并发限制
    - **Property 13: Concurrent Migration Limit**
    - **Validates: Requirements 5.6**
  - [x] 8.8 实现跨组故障转移
    - 组内无可用节点时跨组迁移
    - _Requirements: 5.7_
  - [x] 8.9 编写属性测试：跨组故障转移
    - **Property 14: Cross-Group Failover**
    - **Validates: Requirements 5.7**

- [x] 9. Checkpoint - 确保负载均衡和故障转移测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 4: 节点分组和配置同步
haiyt
- [x] 10. 实现节点分组服务
  - [x] 10.1 创建 NodeGroupService
    - 在 `internal/node/` 创建 `group_service.go`
    - 实现分组 CRUD
    - _Requirements: 6.1, 6.2_
  - [x] 10.2 实现多组成员管理
    - 支持节点加入多个分组
    - _Requirements: 6.3_
  - [x] 10.3 编写属性测试：多组成员
    - **Property 15: Multi-Group Membership**
    - **Validates: Requirements 6.3**
  - [x] 10.4 实现分组统计
    - 聚合分组内节点统计
    - _Requirements: 6.4_
  - [x] 10.5 编写属性测试：统计准确性
    - **Property 16: Group Statistics Accuracy**
    - **Validates: Requirements 6.4**
  - [x] 10.6 实现分组删除逻辑
    - 删除分组时保留节点
    - _Requirements: 6.6_
  - [x] 10.7 编写属性测试：节点保留
    - **Property 17: Node Survival on Group Deletion**
    - **Validates: Requirements 6.6**

- [x] 11. 实现配置同步服务
  - [x] 11.1 创建 ConfigSync 基础结构
    - 在 `internal/node/` 创建 `config_sync.go`
    - _Requirements: 7.1_
  - [x] 11.2 实现配置验证
    - 同步前验证配置有效性
    - _Requirements: 7.7_
  - [x] 11.3 编写属性测试：配置验证
    - **Property 18: Config Validation Before Sync**
    - **Validates: Requirements 7.7**
  - [x] 11.4 实现同步到单节点
    - _Requirements: 7.1_
  - [x] 11.5 实现同步到分组
    - _Requirements: 7.2_
  - [x] 11.6 实现同步重试
    - 失败时自动重试
    - _Requirements: 7.3_
  - [x] 11.7 实现同步状态跟踪
    - 记录每个节点的同步状态
    - _Requirements: 7.4_

- [x] 12. 实现流量统计聚合
  - [x] 12.1 创建 NodeTrafficService
    - 在 `internal/node/` 创建 `traffic_service.go`
    - _Requirements: 8.1_
  - [x] 12.2 实现流量聚合
    - 按用户、代理、节点、分组聚合
    - _Requirements: 8.2_
  - [x] 12.3 编写属性测试：聚合一致性
    - **Property 19: Traffic Aggregation Consistency**
    - **Validates: Requirements 8.2**

- [x] 13. Checkpoint - 确保分组和同步测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 5: 安全和认证

- [x] 14. 实现节点认证
  - [x] 14.1 创建认证中间件
    - 在 `internal/node/` 创建 `auth.go`
    - 实现 Token 验证
    - _Requirements: 10.1_
  - [x] 14.2 编写属性测试：Token 认证
    - **Property 6: Token Authentication**
    - **Validates: Requirements 3.1, 10.1, 10.3**
  - [x] 14.3 实现 Token 轮换
    - 轮换后旧 Token 立即失效
    - _Requirements: 10.2_
  - [x] 14.4 编写属性测试：Token 轮换
    - **Property 20: Token Rotation Invalidation**
    - **Validates: Requirements 10.2**
  - [x] 14.5 实现 IP 白名单
    - 限制节点连接来源 IP
    - _Requirements: 10.5_
  - [x] 14.6 编写属性测试：IP 白名单
    - **Property 21: IP Whitelist Enforcement**
    - **Validates: Requirements 10.5**
  - [x] 14.7 实现认证失败限制
    - 多次失败后临时封禁
    - _Requirements: 10.7_
  - [x] 14.8 编写属性测试：失败限制
    - **Property 22: Auth Failure Rate Limiting**
    - **Validates: Requirements 10.7**

- [x] 15. Checkpoint - 确保安全测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 6: API 层实现

- [x] 16. 实现 API Handlers
  - [x] 16.1 创建 NodeHandler
    - 在 `internal/api/handlers/` 创建 `node.go`
    - 实现节点 CRUD 端点
    - _Requirements: 1.1-1.7_
  - [x] 16.2 创建 NodeGroupHandler
    - 创建 `node_group.go`
    - 实现分组 CRUD 端点
    - _Requirements: 6.1-6.6_
  - [x] 16.3 创建 NodeHealthHandler
    - 创建 `node_health.go`
    - 实现健康检查相关端点
    - _Requirements: 2.1-2.8_
  - [x] 16.4 创建 NodeStatsHandler
    - 创建 `node_stats.go`
    - 实现流量统计端点
    - _Requirements: 8.1-8.6_

- [x] 17. 注册路由
  - [x] 17.1 在 routes.go 中添加节点管理路由
    - 添加 `/api/admin/nodes` 路由组
    - 添加 `/api/admin/node-groups` 路由组
    - 配置管理员权限中间件
    - _Requirements: 9.1-9.7_

- [x] 18. Checkpoint - 确保 API 测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 7: Node Agent 实现

- [x] 19. 实现 Node Agent
  - [x] 19.1 创建 Agent 主程序
    - 在 `cmd/agent/` 创建 `main.go`
    - 实现启动和配置加载
    - _Requirements: 3.7_
  - [x] 19.2 实现注册逻辑
    - 启动时向 Panel 注册
    - _Requirements: 3.7_
  - [x] 19.3 实现心跳上报
    - 定期上报状态和指标
    - _Requirements: 3.2_
  - [x] 19.4 实现命令执行
    - 接收并执行 Panel 命令
    - _Requirements: 3.3, 3.4_
  - [x] 19.5 实现配置同步接收
    - 接收并应用配置更新
    - _Requirements: 3.4_
  - [x] 19.6 实现健康检查端点
    - 提供本地健康检查 API
    - _Requirements: 3.6_
  - [x] 19.7 实现自动重连
    - 连接断开后自动重连
    - _Requirements: 3.5_

- [x] 20. Checkpoint - 确保 Agent 功能完成
  - 测试 Agent 与 Panel 通信
  - 如有问题请询问用户

### Phase 8: 前端实现

- [x] 21. 创建前端 API 模块
  - [x] 21.1 创建 nodes.ts
    - 在 `web/src/api/modules/` 创建节点 API
    - _Requirements: 1.1-1.7_
  - [x] 21.2 创建 nodeGroups.ts
    - 创建分组 API
    - _Requirements: 6.1-6.6_
  - [x] 21.3 创建 nodeHealth.ts
    - 创建健康检查 API
    - _Requirements: 2.1-2.8_

- [x] 22. 创建 Pinia Stores
  - [x] 22.1 创建 node.ts store
    - 在 `web/src/stores/` 创建节点状态管理
    - _Requirements: 1.1-1.7_
  - [x] 22.2 创建 nodeGroup.ts store
    - 创建分组状态管理
    - _Requirements: 6.1-6.6_

- [x] 23. 实现节点管理页面
  - [x] 23.1 创建 AdminNodes.vue
    - 在 `web/src/views/admin/` 创建节点列表页面
    - 显示节点状态、负载、连接数
    - _Requirements: 9.1, 9.2_
  - [x] 23.2 创建 NodeDetail.vue
    - 创建节点详情页面
    - 显示指标、日志、配置
    - _Requirements: 9.2_
  - [x] 23.3 创建 NodeForm.vue
    - 创建节点添加/编辑表单
    - _Requirements: 1.1, 1.4_

- [x] 24. 实现节点分组页面
  - [x] 24.1 创建 AdminNodeGroups.vue
    - 创建分组管理页面
    - _Requirements: 6.1-6.6_

- [x] 25. 实现节点仪表板
  - [x] 25.1 创建 NodeDashboard.vue
    - 创建集群健康概览
    - _Requirements: 9.3_
  - [x] 25.2 创建 NodeMap.vue
    - 创建节点地理分布地图
    - _Requirements: 9.5_
  - [x] 25.3 创建 NodeComparison.vue
    - 创建节点性能对比视图
    - _Requirements: 9.6_

- [x] 26. 添加路由配置
  - [x] 26.1 更新 router 配置
    - 添加所有节点管理页面路由
    - 配置权限守卫
    - _Requirements: 9.1-9.7_

- [x] 27. Final Checkpoint - 完整功能测试
  - 运行所有后端测试
  - 运行前端构建
  - 手动测试完整流程
  - 如有问题请询问用户

## Notes

- 所有任务都是必须完成的，包括测试任务
- 每个 Checkpoint 用于验证阶段性成果
- 属性测试使用 Go 的 `testing/quick` 或 `gopter` 库
- Node Agent 可以作为独立二进制文件部署
- 前端开发可以与后端 API 并行进行（使用 mock 数据）
