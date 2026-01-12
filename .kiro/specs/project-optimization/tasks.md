# Implementation Plan: V Panel 项目优化改进

## Overview

本实现计划将 V Panel 项目优化改进分解为可执行的编码任务。任务按照依赖关系排序，确保每个任务都建立在前一个任务的基础上。测试任务标记为可选（*），可以在 MVP 阶段跳过。

## Tasks

- [x] 1. 基础设施和错误处理
  - [x] 1.1 实现标准化错误类型和响应格式
    - 创建 `pkg/errors/app_error.go` 定义 AppError 结构
    - 实现错误码常量和 HTTP 状态码映射
    - 创建 ErrorResponse 结构用于 API 响应
    - _Requirements: 2.1, 2.2, 2.5_

  - [x] 1.2 编写错误响应一致性属性测试
    - **Property 1: Error Response Consistency**
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.5**

  - [x] 1.3 实现请求验证中间件
    - 创建 `internal/api/middleware/validation.go`
    - 实现 JSON Schema 验证
    - 实现字段级错误消息返回
    - _Requirements: 3.1, 3.2, 3.5_

  - [x] 1.4 实现请求 ID 和关联 ID 中间件
    - 更新 `internal/api/middleware/request_id.go`
    - 确保关联 ID 传播到日志和错误响应
    - _Requirements: 10.3_

  - [x] 1.5 编写关联 ID 传播属性测试
    - **Property 28: Correlation ID Propagation**
    - **Validates: Requirements 10.3**

- [x] 2. Checkpoint - 确保所有测试通过
  - 运行测试，如有问题请询问用户

- [x] 3. 认证和安全增强 ✅
  - [x] 3.1 实现速率限制器 ✅
    - 创建 `internal/auth/rate_limiter.go`
    - 实现基于 IP 的登录尝试限制（5次/分钟）
    - 集成到登录 handler
    - _Requirements: 1.2_

  - [x] 3.2 编写速率限制属性测试
    - **Property 2: Rate Limiting Enforcement**
    - **Validates: Requirements 1.2**

  - [x] 3.3 实现 Token 黑名单机制
    - 创建 `internal/auth/token_blacklist.go`
    - 实现 token 撤销和检查功能
    - 更新 auth 中间件检查黑名单
    - _Requirements: 1.6_

  - [x] 3.4 编写 Token 撤销属性测试 ✅
    - **Property 5: Token Revocation**
    - **Validates: Requirements 1.6**

  - [x] 3.5 实现输入清理工具 ✅
    - 创建 `pkg/sanitizer/sanitizer.go`
    - 实现 SQL 注入、XSS、命令注入防护
    - 集成到验证中间件
    - _Requirements: 1.4_

  - [x] 3.6 编写输入清理属性测试 ✅
    - **Property 3: Input Sanitization**
    - **Validates: Requirements 1.4**

  - [x] 3.7 增强配置验证 ✅
    - 更新 `internal/config/config.go`
    - 添加 JWT 密钥最小长度验证（32字符）
    - 添加数据库连接字符串格式验证
    - _Requirements: 1.7, 11.1, 11.2, 11.3, 11.4_

  - [x] 3.8 编写配置验证属性测试 ✅
    - **Property 4: JWT Secret Validation**
    - **Property 22: Configuration Validation**
    - **Validates: Requirements 1.7, 11.1, 11.2**

- [x] 4. Checkpoint - 确保所有测试通过 ✅
  - 运行测试，如有问题请询问用户

- [x] 5. 缓存层实现
  - [x] 5.1 实现缓存服务接口
    - 创建 `internal/cache/cache.go` 定义接口
    - 创建 `internal/cache/memory.go` 内存缓存实现
    - 创建 `internal/cache/redis.go` Redis 缓存实现
    - _Requirements: 4.1, 4.5_

  - [x] 5.2 实现缓存配置和工厂
    - 创建 `internal/cache/config.go`
    - 实现缓存类型选择和 TTL 配置
    - _Requirements: 4.3_

  - [x] 5.3 集成缓存到 Repository 层
    - 更新 `internal/database/repository/user_repository.go`
    - 更新 `internal/database/repository/proxy_repository.go`
    - 实现缓存读取和失效逻辑
    - _Requirements: 4.1, 4.2, 4.6_

  - [x] 5.4 编写缓存一致性属性测试
    - **Property 6: Cache Consistency**
    - **Property 7: Cache TTL Expiration**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.6**

- [x] 6. 数据库优化
  - [x] 6.1 实现数据库迁移版本控制
    - 创建 `internal/database/migrations/` 目录
    - 创建迁移版本表和迁移管理器
    - _Requirements: 5.3_

  - [x] 6.2 添加数据库索引
    - 创建迁移文件添加所有必要索引
    - users: username, email, role
    - proxies: user_id, protocol, port, (user_id, enabled)
    - traffic: user_id, proxy_id, recorded_at, (user_id, recorded_at)
    - _Requirements: 16.1-16.10_

  - [x] 6.3 实现数据库连接健康检查和重连
    - 更新 `internal/database/db.go`
    - 添加连接健康检查
    - 实现自动重连机制
    - _Requirements: 15.2, 15.3_

  - [x] 6.4 编写数据库连接重试属性测试
    - **Property 23: Database Connection Retry**
    - **Validates: Requirements 15.2**

  - [x] 6.5 实现慢查询日志
    - 创建 GORM 日志钩子
    - 记录超过阈值的查询
    - _Requirements: 15.5_

  - [x] 6.6 编写慢查询日志属性测试
    - **Property 24: Slow Query Logging**
    - **Validates: Requirements 15.5**

- [x] 7. Checkpoint - 确保所有测试通过
  - 运行测试，如有问题请询问用户

- [x] 8. 用户管理功能完善
  - [x] 8.1 更新用户模型
    - 添加 traffic_limit, traffic_used, expires_at, force_password_change 字段
    - 创建数据库迁移
    - _Requirements: 17.6, 17.7, 17.8_

  - [x] 8.2 实现用户启用/禁用 API
    - 添加 `POST /api/users/:id/enable` 端点
    - 添加 `POST /api/users/:id/disable` 端点
    - 更新登录逻辑检查用户状态
    - _Requirements: 17.1, 17.2, 17.3_

  - [x] 8.3 编写用户启用/禁用属性测试
    - **Property 8: User Enable/Disable**
    - **Validates: Requirements 17.1, 17.2, 17.3**

  - [x] 8.4 实现密码重置功能
    - 添加 `POST /api/users/:id/reset-password` 端点
    - 生成临时密码并设置 force_password_change
    - _Requirements: 17.4, 17.5_

  - [x] 8.5 实现流量限制和过期检查
    - 创建流量检查中间件
    - 实现过期检查逻辑
    - _Requirements: 17.9, 17.10_

  - [x] 8.6 编写用户访问控制属性测试
    - **Property 9: User Access Control**
    - **Validates: Requirements 17.9, 17.10**

  - [x] 8.7 实现登录历史记录
    - 创建 login_history 表
    - 添加 `GET /api/users/:id/login-history` 端点
    - 添加 `DELETE /api/users/:id/login-history` 端点
    - 记录所有登录尝试
    - _Requirements: 17.11, 17.12, 17.13_

  - [x] 8.8 编写登录历史记录属性测试
    - **Property 26: Login History Recording**
    - **Validates: Requirements 17.12**

  - [x] 8.9 增强用户验证
    - 添加邮箱格式验证
    - 添加用户名唯一性检查
    - _Requirements: 17.14, 17.15_

  - [x] 8.10 编写用户验证属性测试 ✅
    - **Property 10: Email Validation**
    - **Property 11: Username Uniqueness**
    - **Validates: Requirements 17.14, 17.15**

- [x] 9. Checkpoint - 确保所有测试通过 ✅
  - 运行测试，如有问题请询问用户

- [x] 10. 角色管理持久化
  - [x] 10.1 创建角色数据库模型
    - 创建 roles 表迁移
    - 创建 `internal/database/repository/role_repository.go`
    - _Requirements: 19.1, 19.2_

  - [x] 10.2 更新角色 Handler
    - 重构 `internal/api/handlers/roles.go` 使用数据库
    - 实现系统角色初始化
    - _Requirements: 19.3_

  - [x] 10.3 实现角色保护逻辑
    - 防止删除系统角色
    - 防止修改系统角色权限
    - 实现角色删除时用户重新分配
    - _Requirements: 19.4, 19.5, 19.6_

  - [x] 10.4 编写角色保护属性测试
    - **Property 17: System Role Protection**
    - **Property 18: Role Deletion User Reassignment**
    - **Validates: Requirements 19.4, 19.5, 19.6**

  - [x] 10.5 实现权限验证
    - 验证权限键是否有效
    - 实现权限继承（admin 继承所有权限）
    - _Requirements: 19.7, 19.8_

  - [x] 10.6 编写权限验证属性测试
    - **Property 19: Permission Validation**
    - **Validates: Requirements 19.7**

- [x] 11. 系统设置功能实现
  - [x] 11.1 创建设置数据库模型
    - 创建 settings 表迁移
    - 创建 `internal/database/repository/settings_repository.go`
    - _Requirements: 18.3_

  - [x] 11.2 实现设置服务
    - 创建 `internal/settings/service.go`
    - 实现设置获取和更新
    - _Requirements: 18.1, 18.2, 18.4, 18.5, 18.6, 18.10, 18.11_

  - [x] 11.3 实现设置 Handler
    - 创建 `internal/api/handlers/settings.go`
    - 添加 `GET /api/settings` 端点
    - 添加 `PUT /api/settings` 端点
    - _Requirements: 18.1, 18.2_

  - [x] 11.4 编写设置持久化属性测试
    - **Property 27: Settings Persistence**
    - **Validates: Requirements 18.3**

  - [x] 11.5 实现设置备份恢复
    - 添加 `POST /api/settings/backup` 端点
    - 添加 `POST /api/settings/restore` 端点
    - _Requirements: 18.8, 18.9_

  - [x] 11.6 实现设置热更新
    - 设置更新后无需重启即可生效
    - _Requirements: 18.7_

- [x] 12. Checkpoint - 确保所有测试通过
  - 运行测试，如有问题请询问用户

- [x] 13. 代理服务功能完善
  - [x] 13.1 更新代理模型
    - 添加 user_id 字段
    - 创建数据库迁移
    - _Requirements: 21.1_

  - [x] 13.2 实现代理用户关联
    - 创建代理时设置 user_id
    - 非管理员用户只能查看自己的代理
    - _Requirements: 21.2, 21.3_

  - [x] 13.3 编写代理用户关联属性测试
    - **Property 13: Proxy User Association**
    - **Validates: Requirements 21.2, 21.3**

  - [x] 13.4 实现端口冲突检测
    - 创建代理前检查端口是否被占用
    - 返回冲突代理信息
    - _Requirements: 21.8, 21.9_

  - [x] 13.5 编写端口冲突检测属性测试
    - **Property 12: Port Conflict Detection**
    - **Validates: Requirements 21.8, 21.9**

  - [x] 13.6 实现代理启动/停止 API
    - 添加 `POST /api/proxies/:id/start` 端点
    - 添加 `POST /api/proxies/:id/stop` 端点
    - _Requirements: 21.4, 21.5_

  - [x] 13.7 实现代理统计 API
    - 添加 `GET /api/proxies/:id/stats` 端点
    - 返回流量、连接数、最后活跃时间
    - _Requirements: 21.10, 21.11_

  - [x] 13.8 实现代理批量操作
    - 添加批量启用/禁用/删除功能
    - _Requirements: 21.14_

  - [x] 13.9 增强代理验证
    - 验证协议特定设置
    - _Requirements: 21.12_

- [x] 14. Xray 集成完善
  - [x] 14.1 实现 Xray 管理器
    - 创建 `internal/xray/manager.go`
    - 实现进程启动、停止、重启
    - 实现状态监控
    - _Requirements: 22.1, 22.2_

  - [x] 14.2 实现 Xray API 端点
    - 添加 `GET /api/xray/status` 端点
    - 添加 `POST /api/xray/restart` 端点
    - 添加 `GET /api/xray/config` 端点
    - 添加 `PUT /api/xray/config` 端点
    - 添加 `GET /api/xray/version` 端点
    - 添加 `POST /api/xray/update` 端点
    - _Requirements: 22.3, 22.4, 22.5, 22.6, 22.9, 22.10_

  - [x] 14.3 实现配置同步
    - 代理创建/更新/删除时同步 Xray 配置
    - _Requirements: 22.7_

  - [x] 14.4 编写 Xray 配置同步属性测试
    - **Property 14: Xray Configuration Sync**
    - **Validates: Requirements 21.6, 21.7, 21.13**

  - [x] 14.5 实现配置验证和回滚
    - 应用配置前验证
    - 失败时自动回滚
    - _Requirements: 22.8, 22.11, 22.12_

  - [x] 14.6 编写 Xray 配置验证属性测试
    - **Property 15: Xray Configuration Validation**
    - **Property 16: Xray Configuration Rollback**
    - **Validates: Requirements 22.8, 22.11, 22.12**

- [x] 15. Checkpoint - 确保所有测试通过
  - 运行测试，如有问题请询问用户

- [x] 16. 统计数据实时查询 ✅
  - [x] 16.1 重构统计 Handler
    - 更新 `internal/api/handlers/stats.go`
    - 从数据库查询实际数据
    - _Requirements: 20.1_

  - [x] 16.2 实现仪表盘统计
    - 查询用户总数、活跃用户数
    - 查询代理总数、活跃代理数
    - 查询流量统计
    - _Requirements: 20.2, 20.3, 20.4_

  - [x] 16.3 编写统计准确性属性测试 ✅
    - **Property 20: Statistics Accuracy**
    - **Validates: Requirements 20.1, 20.2, 20.3, 20.4**

  - [x] 16.4 实现协议和用户统计 ✅
    - 按协议聚合流量
    - 按用户聚合流量
    - _Requirements: 20.5, 20.6_

  - [x] 16.5 实现时间段过滤 ✅
    - 支持 today, week, month, year, custom
    - _Requirements: 20.7_

  - [x] 16.6 编写时间段过滤属性测试 ✅
    - **Property 21: Traffic Period Filtering**
    - **Validates: Requirements 20.7**

  - [x] 16.7 实现统计缓存 ✅
    - 缓存昂贵的聚合查询
    - _Requirements: 20.8_

  - [x] 16.8 实现时间线数据 ✅
    - 返回小时/天级别的流量数据点
    - _Requirements: 20.9_

- [x] 17. 监控与可观测性 ✅
  - [x] 17.1 实现 Prometheus 指标 ✅
    - 创建 `internal/monitor/metrics.go`
    - 添加 `/metrics` 端点
    - 记录请求延迟、计数、错误率
    - _Requirements: 10.1, 10.2_

  - [x] 17.2 实现结构化日志 ✅
    - 更新 `internal/logger/logger.go`
    - 确保一致的字段名
    - _Requirements: 10.4_

  - [x] 17.3 编写结构化日志属性测试 ✅
    - **Property 29: Structured Logging Consistency**
    - **Validates: Requirements 10.4**

  - [x] 17.4 增强健康检查 ✅
    - 检查数据库连接
    - 检查 Xray 进程状态
    - 检查磁盘空间
    - _Requirements: 10.5_

  - [x] 17.5 实现审计日志 ✅
    - 创建 audit_logs 表
    - 记录敏感操作
    - _Requirements: 1.5_

  - [x] 17.6 编写审计日志属性测试 ✅
    - **Property 25: Audit Logging**
    - **Validates: Requirements 1.5**

- [x] 18. Checkpoint - 确保所有测试通过 ✅
  - 运行测试，如有问题请询问用户

- [ ] 19. 前端 API 统一
  - [ ] 19.1 重构 API 模块
    - 合并重复的 API 定义
    - 按领域组织端点
    - _Requirements: 14.1, 14.2_

  - [ ] 19.2 添加 TypeScript 类型定义
    - 创建 `web/src/types/api.ts`
    - 定义请求和响应类型
    - _Requirements: 14.4_

  - [ ] 19.3 实现错误处理增强
    - 实现集中式错误处理
    - 实现错误码到本地化消息映射
    - _Requirements: 14.5, 13.7_

  - [ ] 19.4 编写前端错误码映射属性测试
    - **Property 30: Frontend Error Code Mapping**
    - **Validates: Requirements 13.7**

  - [ ] 19.5 实现请求取消和去重
    - 添加请求取消支持
    - 实现并发请求去重
    - _Requirements: 14.6, 14.7_

  - [ ] 19.6 实现离线请求队列
    - 检测网络状态
    - 队列离线请求
    - 恢复后自动重试
    - _Requirements: 14.8_

- [ ] 20. 前端状态管理完善
  - [ ] 20.1 创建 Pinia Stores
    - 创建 `web/src/stores/proxy.ts`
    - 创建 `web/src/stores/system.ts`
    - 创建 `web/src/stores/settings.ts`
    - 创建 `web/src/stores/notification.ts`
    - _Requirements: 7.1_

  - [ ] 20.2 实现加载和错误状态跟踪
    - 每个 store 添加 loading 和 error 状态
    - _Requirements: 7.2_

  - [ ] 20.3 实现状态持久化
    - 关键状态持久化到 sessionStorage
    - _Requirements: 7.3_

  - [ ] 20.4 实现重试逻辑
    - API 调用失败时指数退避重试
    - _Requirements: 7.4_

  - [ ] 20.5 实现乐观更新
    - 更新操作先更新 UI 再等待响应
    - _Requirements: 7.5_

- [ ] 21. 前端性能优化
  - [ ] 21.1 实现路由代码分割
    - 配置 Vue Router 懒加载
    - _Requirements: 8.1_

  - [ ] 21.2 实现请求防抖
    - 搜索输入防抖
    - _Requirements: 8.4_

  - [ ] 21.3 编写请求防抖属性测试
    - **Property 32: Request Debouncing**
    - **Validates: Requirements 8.4**

  - [ ] 21.4 实现虚拟滚动
    - 大列表使用虚拟滚动
    - _Requirements: 8.2_

- [ ] 22. 前端显示优化
  - [ ] 22.1 实现骨架屏加载
    - 数据加载时显示骨架屏
    - _Requirements: 8.1.1_

  - [ ] 22.2 实现空状态显示
    - 无数据时显示有意义的空状态
    - _Requirements: 8.1.2_

  - [ ] 22.3 实现表格增强
    - 可排序列
    - 可筛选行
    - _Requirements: 8.1.5_

  - [ ] 22.4 实现表单验证反馈
    - 提交前内联验证
    - _Requirements: 8.1.8_

  - [ ] 22.5 实现状态颜色指示器
    - 代理状态颜色编码
    - _Requirements: 8.1.10_

- [ ] 23. 前端错误处理
  - [ ] 23.1 实现全局错误捕获
    - 捕获未处理的 JavaScript 错误
    - _Requirements: 13.5_

  - [ ] 23.2 实现错误 ID 显示
    - 错误消息包含唯一 ID
    - _Requirements: 13.10_

  - [ ] 23.3 编写错误 ID 属性测试
    - **Property 31: Frontend Error ID**
    - **Validates: Requirements 13.10**

  - [ ] 23.4 实现错误队列
    - 多错误队列显示
    - _Requirements: 13.11_

  - [ ] 23.5 实现客户端错误日志上报
    - 错误发送到后端
    - _Requirements: 13.9_

- [ ] 24. 数据序列化测试
  - [ ] 24.1 编写序列化往返属性测试
    - **Property 33: Data Serialization Round-Trip**
    - **Validates: Requirements 9.5**

- [ ] 25. Final Checkpoint - 确保所有测试通过
  - 运行所有测试
  - 验证所有功能正常工作
  - 如有问题请询问用户

## Notes

- 所有任务都是必需的，包括测试任务
- 每个任务都引用了具体的需求以便追溯
- Checkpoint 任务确保增量验证
- 属性测试验证通用正确性属性
- 单元测试验证特定示例和边界情况
