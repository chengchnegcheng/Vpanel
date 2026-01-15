# Implementation Plan: User Portal

## Overview

本实现计划将用户前台门户系统分解为可执行的编码任务，按照数据库 → 后端服务 → API → 前端的顺序实现。由于功能较多，分为多个阶段逐步完成。

## Tasks

### Phase 1: 数据库和基础设施

- [x] 1. 创建新数据模型和数据库迁移
  - [x] 1.1 创建工单相关模型
    - 在 `internal/database/models/` 创建 `ticket.go`
    - 定义 Ticket 和 TicketMessage 结构体
    - 添加 GORM 标签和外键关系
    - _Requirements: 10.1, 10.2, 10.3_
  - [x] 1.2 创建公告相关模型
    - 创建 `announcement.go`
    - 定义 Announcement 和 AnnouncementRead 结构体
    - _Requirements: 9.1, 9.4_
  - [x] 1.3 创建帮助中心模型
    - 创建 `help_article.go`
    - 定义 HelpArticle 结构体
    - _Requirements: 12.1, 12.2_
  - [x] 1.4 创建认证令牌模型
    - 创建 `auth_token.go`
    - 定义 PasswordResetToken、EmailVerificationToken、InviteCode、TwoFactorSecret 结构体
    - _Requirements: 1.6, 1.8, 3.1, 8.4_
  - [x] 1.5 扩展 User 模型
    - 添加 email_verified、two_factor_enabled、telegram_id 等新字段
    - _Requirements: 1.10, 2.8, 8.8_
  - [x] 1.6 创建数据库迁移文件
    - 创建所有新表的迁移
    - 创建索引
    - 添加 User 表扩展字段
    - _Requirements: 1.1-14.10_

- [x] 2. 创建 Repository 层
  - [x] 2.1 创建 TicketRepository
    - 实现 Create、GetByID、GetByUserID、Update、ListByUser 方法
    - 实现 AddMessage、GetMessages 方法
    - _Requirements: 10.1-10.11_
  - [x] 2.2 创建 AnnouncementRepository
    - 实现 Create、GetByID、List、MarkAsRead、GetUnreadCount 方法
    - _Requirements: 9.1-9.10_
  - [x] 2.3 创建 HelpArticleRepository
    - 实现 GetBySlug、List、Search、IncrementViewCount 方法
    - _Requirements: 12.1-12.10_
  - [x] 2.4 创建 AuthTokenRepository
    - 实现密码重置令牌、邮箱验证令牌、邀请码的 CRUD 方法
    - _Requirements: 1.6, 3.1-3.6_
  - [x] 2.5 编写 Repository 集成测试
    - 测试所有 CRUD 操作
    - _Requirements: 10.6, 9.4_

- [x] 3. Checkpoint - 确保数据库层测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 2: 后端认证服务

- [x] 4. 实现用户注册服务
  - [x] 4.1 创建 PortalAuthService 基础结构
    - 在 `internal/portal/auth/` 创建 `service.go`
    - 实现邮箱格式验证
    - 实现密码强度验证
    - _Requirements: 1.2, 1.3_
  - [x] 4.2 实现注册逻辑
    - 检查用户名/邮箱重复
    - 验证邀请码（如启用）
    - 创建用户记录
    - 发送验证邮件
    - _Requirements: 1.4, 1.5, 1.6, 1.7, 1.8_
  - [x] 4.3 实现邮箱验证
    - 生成验证令牌
    - 验证令牌并激活账户
    - _Requirements: 1.8, 1.10_
  - [x] 4.4 编写属性测试：输入验证
    - **Property 1: Email Format Validation**
    - **Property 2: Password Strength Validation**
    - **Validates: Requirements 1.2, 1.3**
  - [x] 4.5 编写属性测试：用户名/邮箱唯一性
    - **Property 3: Username/Email Uniqueness**
    - **Validates: Requirements 1.4, 1.5**

- [x] 5. 实现用户登录服务
  - [x] 5.1 实现登录逻辑
    - 验证凭据
    - 检查账户状态
    - 生成 JWT 令牌
    - 记录登录日志
    - _Requirements: 2.1, 2.2, 2.3, 2.10_
  - [x] 5.2 实现登录速率限制
    - 基于 IP 的速率限制（5次/15分钟）
    - 锁定和解锁逻辑
    - _Requirements: 2.4, 2.5_
  - [x] 5.3 实现 2FA 验证
    - TOTP 令牌生成和验证
    - 备份码生成和验证
    - _Requirements: 2.8, 2.9_
  - [x] 5.4 编写属性测试：速率限制
    - **Property 4: Login Rate Limiting**
    - **Validates: Requirements 2.4, 2.5**
  - [x] 5.5 编写属性测试：2FA 验证
    - **Property 15: 2FA Token Validation**
    - **Validates: Requirements 2.8, 2.9**

- [x] 6. 实现密码重置服务
  - [x] 6.1 实现密码重置请求
    - 生成重置令牌
    - 发送重置邮件
    - _Requirements: 3.1, 3.6_
  - [x] 6.2 实现密码重置执行
    - 验证令牌有效性和过期时间
    - 更新密码
    - 使令牌失效
    - 使所有会话失效
    - _Requirements: 3.2, 3.3, 3.4, 3.5_
  - [x] 6.3 编写属性测试：令牌过期
    - **Property 5: Password Reset Token Expiration**
    - **Validates: Requirements 3.2**
  - [x] 6.4 编写属性测试：令牌单次使用
    - **Property 6: Password Reset Token Single-Use**
    - **Validates: Requirements 3.3**
  - [x] 6.5 编写属性测试：会话失效
    - **Property 7: Session Invalidation on Password Reset**
    - **Validates: Requirements 3.5**

- [x] 7. Checkpoint - 确保认证服务测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 3: 后端业务服务

- [x] 8. 实现工单服务
  - [x] 8.1 创建 TicketService
    - 在 `internal/portal/ticket/` 创建 `service.go`
    - 实现创建工单、获取工单、回复工单、关闭工单
    - _Requirements: 10.1-10.10_
  - [x] 8.2 实现工单状态机
    - 定义状态转换规则
    - 实现状态变更逻辑
    - _Requirements: 10.3, 10.8_
  - [x] 8.3 实现附件处理
    - 文件上传和存储
    - 大小限制验证
    - _Requirements: 10.5_
  - [x] 8.4 编写属性测试：工单 ID 唯一性
    - **Property 11: Ticket ID Uniqueness**
    - **Validates: Requirements 10.6**
  - [x] 8.5 编写属性测试：状态转换
    - **Property 12: Ticket Status Transitions**
    - **Validates: Requirements 10.3**

- [x] 9. 实现公告服务
  - [x] 9.1 创建 AnnouncementService
    - 在 `internal/portal/announcement/` 创建 `service.go`
    - 实现获取公告列表、获取详情、标记已读
    - _Requirements: 9.1-9.10_
  - [x] 9.2 实现已读状态跟踪
    - 记录用户已读状态
    - 获取未读数量
    - _Requirements: 9.4, 9.5_
  - [x] 9.3 编写属性测试：已读状态
    - **Property 10: Announcement Read Status Tracking**
    - **Validates: Requirements 9.4**

- [x] 10. 实现帮助中心服务
  - [x] 10.1 创建 HelpService
    - 在 `internal/portal/help/` 创建 `service.go`
    - 实现文章列表、详情、搜索
    - _Requirements: 12.1-12.10_
  - [x] 10.2 实现搜索功能
    - 全文搜索标题、内容、标签
    - _Requirements: 12.3_
  - [x] 10.3 编写属性测试：搜索相关性
    - **Property 14: Help Article Search Relevance**
    - **Validates: Requirements 12.3**

- [x] 11. 实现节点服务扩展
  - [x] 11.1 创建 PortalNodeService
    - 在 `internal/portal/node/` 创建 `service.go`
    - 实现用户可见节点列表
    - 实现节点过滤和排序
    - _Requirements: 5.1-5.10_
  - [x] 11.2 实现延迟测试
    - 实现 ping 测试逻辑
    - _Requirements: 5.7, 5.8_
  - [x] 11.3 编写属性测试：过滤和排序
    - **Property 8: Node List Filtering Correctness**
    - **Property 9: Node List Sorting Correctness**
    - **Validates: Requirements 5.3, 5.4**

- [x] 12. 实现统计服务
  - [x] 12.1 创建 PortalStatsService
    - 在 `internal/portal/stats/` 创建 `service.go`
    - 实现流量统计查询
    - 实现按日/周/月聚合
    - _Requirements: 11.1-11.10_
  - [x] 12.2 实现数据导出
    - CSV 格式导出
    - _Requirements: 11.10_
  - [x] 12.3 编写属性测试：统计一致性
    - **Property 13: Traffic Statistics Consistency**
    - **Validates: Requirements 11.2, 11.3**

- [x] 13. Checkpoint - 确保业务服务测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 4: 后端 API 层

- [x] 14. 实现 Portal API Handlers
  - [x] 14.1 创建 PortalAuthHandler
    - 实现注册、登录、登出、密码重置等端点
    - _Requirements: 1.1-3.6_
  - [x] 14.2 创建 PortalDashboardHandler
    - 实现仪表板数据端点
    - _Requirements: 4.1-4.10_
  - [x] 14.3 创建 PortalNodeHandler
    - 实现节点列表、详情、延迟测试端点
    - _Requirements: 5.1-5.10_
  - [x] 14.4 创建 PortalTicketHandler
    - 实现工单 CRUD 端点
    - _Requirements: 10.1-10.11_
  - [x] 14.5 创建 PortalAnnouncementHandler
    - 实现公告列表、详情、已读标记端点
    - _Requirements: 9.1-9.10_
  - [x] 14.6 创建 PortalStatsHandler
    - 实现统计查询和导出端点
    - _Requirements: 11.1-11.10_
  - [x] 14.7 创建 PortalHelpHandler
    - 实现帮助文章列表、详情、搜索端点
    - _Requirements: 12.1-12.10_

- [x] 15. 注册 Portal 路由
  - [x] 15.1 在 routes.go 中添加 Portal 路由组
    - 添加 `/api/portal/*` 路由
    - 配置认证中间件
    - _Requirements: 13.1, 13.7_
  - [x] 15.2 实现 Portal 认证中间件
    - 用户身份验证
    - 账户状态检查
    - _Requirements: 2.3, 4.10_

- [x] 16. 编写 API 集成测试
  - [x] 16.1 编写认证 API 测试
    - 测试注册、登录、密码重置流程
    - _Requirements: 1.1-3.6_
  - [x] 16.2 编写业务 API 测试
    - 测试工单、公告、统计 API
    - _Requirements: 9.1-12.10_

- [x] 17. Checkpoint - 确保后端 API 测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 5: 前端基础设施

- [x] 18. 创建前端基础结构
  - [x] 18.1 创建 Portal 路由配置
    - 在 `web/src/router/` 创建 `user.ts`
    - 配置所有用户前台路由
    - 添加路由守卫
    - _Requirements: 13.1, 13.7_
  - [x] 18.2 创建 Portal 布局组件
    - 创建 `UserLayout.vue`（主布局）
    - 创建 `AuthLayout.vue`（登录/注册布局）
    - _Requirements: 13.2, 13.3, 13.4_
  - [x] 18.3 创建 Portal API 模块
    - 创建 `portalAuth.ts`、`portalNodes.ts` 等 API 模块
    - _Requirements: 14.1-14.10_
  - [x] 18.4 创建 Portal Pinia Stores
    - 创建 `userPortal.ts`、`nodes.ts`、`tickets.ts` 等 Store
    - _Requirements: 4.1-4.10_

- [x] 19. 实现认证页面
  - [x] 19.1 创建登录页面
    - 创建 `Login.vue`
    - 实现登录表单和验证
    - 实现 2FA 验证流程
    - _Requirements: 2.1-2.10_
  - [x] 19.2 创建注册页面
    - 创建 `Register.vue`
    - 实现注册表单和验证
    - 实现邀请码输入
    - _Requirements: 1.1-1.10_
  - [x] 19.3 创建密码重置页面
    - 创建 `ForgotPassword.vue` 和 `ResetPassword.vue`
    - _Requirements: 3.1-3.6_

- [x] 20. Checkpoint - 确保前端基础结构完成
  - 验证路由配置
  - 验证布局组件
  - 如有问题请询问用户

### Phase 6: 前端核心页面

- [x] 21. 实现用户仪表板
  - [x] 21.1 创建 Dashboard 页面
    - 创建 `Dashboard.vue`
    - 显示用户信息、流量使用、到期时间
    - _Requirements: 4.1-4.10_
  - [x] 21.2 创建仪表板组件
    - 创建 `TrafficCard.vue`（流量卡片）
    - 创建 `QuickActions.vue`（快捷操作）
    - _Requirements: 4.2, 4.7_

- [x] 22. 实现节点列表页面
  - [x] 22.1 创建 Nodes 页面
    - 创建 `Nodes.vue`
    - 实现节点列表展示
    - 实现过滤和排序
    - _Requirements: 5.1-5.10_
  - [x] 22.2 创建节点组件
    - 创建 `NodeCard.vue`
    - 实现延迟测试功能
    - _Requirements: 5.7, 5.8_

- [x] 23. 实现订阅管理页面
  - [x] 23.1 创建 Subscription 页面
    - 创建 `Subscription.vue`
    - 集成订阅链接显示、QR码、格式选择
    - _Requirements: 6.1-6.10_
  - [x] 23.2 复用订阅系统组件
    - 复用 subscription-system 中的组件
    - _Requirements: 6.2, 6.3, 6.4_

- [x] 24. 实现客户端下载页面
  - [x] 24.1 创建 Download 页面
    - 创建 `Download.vue`
    - 按平台分组显示客户端
    - 显示推荐标记和教程链接
    - _Requirements: 7.1-7.10_

- [x] 25. 实现个人设置页面
  - [x] 25.1 创建 Settings 页面
    - 创建 `Settings.vue`
    - 实现个人资料、安全、通知设置
    - _Requirements: 8.1-8.10_
  - [x] 25.2 实现 2FA 设置
    - 显示 QR 码和备份码
    - 实现启用/禁用流程
    - _Requirements: 8.4, 8.5_

- [x] 26. Checkpoint - 确保核心页面完成
  - 验证所有核心页面功能
  - 如有问题请询问用户

### Phase 7: 前端扩展页面

- [x] 27. 实现公告中心
  - [x] 27.1 创建 Announcements 页面
    - 创建 `Announcements.vue` 和 `AnnouncementDetail.vue`
    - 实现公告列表和详情
    - _Requirements: 9.1-9.10_

- [x] 28. 实现工单系统
  - [x] 28.1 创建 Tickets 页面
    - 创建 `Tickets.vue`、`TicketDetail.vue`、`TicketCreate.vue`
    - 实现工单列表、详情、创建、回复
    - _Requirements: 10.1-10.11_

- [x] 29. 实现统计页面
  - [x] 29.1 创建 Stats 页面
    - 创建 `Stats.vue`
    - 实现流量图表（日/周/月）
    - 实现数据导出
    - _Requirements: 11.1-11.10_

- [x] 30. 实现帮助中心
  - [x] 30.1 创建 HelpCenter 页面
    - 创建 `HelpCenter.vue` 和 `HelpArticle.vue`
    - 实现文章列表、搜索、详情
    - _Requirements: 12.1-12.10_

- [x] 31. 实现移动端适配
  - [x] 31.1 创建移动端布局
    - 创建 `MobileLayout.vue`
    - 实现底部导航栏
    - _Requirements: 14.1-14.10_
  - [x] 31.2 优化移动端组件
    - 调整卡片布局
    - 优化触摸交互
    - _Requirements: 14.4, 14.5, 14.6_

- [x] 32. 实现主题切换
  - [x] 32.1 实现深色/浅色主题
    - 添加主题切换逻辑
    - 持久化用户偏好
    - _Requirements: 8.9, 13.9_

- [x] 33. Final Checkpoint - 完整功能测试
  - 运行所有后端测试
  - 运行前端构建
  - 手动测试完整流程
  - 测试移动端适配
  - 如有问题请询问用户

## Notes

- 所有任务都是必须完成的，包括测试任务
- 每个任务都引用了具体的需求编号以便追溯
- Checkpoint 任务用于确保阶段性质量
- 属性测试验证系统的正确性属性
- 单元测试验证具体示例和边界情况
- 前端开发可以与后端 API 并行进行（使用 mock 数据）
