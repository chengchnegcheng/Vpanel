# Implementation Plan: Subscription System

## Overview

本实现计划将订阅链接系统分解为可执行的编码任务，按照后端核心 → 格式生成器 → API 层 → 前端的顺序实现。每个任务都包含具体的实现目标和对应的需求引用。

## Tasks

- [x] 1. 创建订阅数据模型和数据库迁移
  - [x] 1.1 创建 Subscription 模型定义
    - 在 `internal/database/models/` 创建 `subscription.go`
    - 定义 Subscription 结构体，包含 ID、UserID、Token、ShortCode、访问统计字段
    - 添加 GORM 标签和外键关系
    - _Requirements: 10.1, 10.4_
  - [x] 1.2 创建数据库迁移文件
    - 在 `internal/database/migrations/` 创建订阅表迁移
    - 创建 subscriptions 表和索引
    - 添加外键约束和级联删除
    - _Requirements: 10.2, 10.3, 10.5_
  - [x] 1.3 编写数据模型单元测试
    - 测试模型字段验证
    - 测试外键关系
    - _Requirements: 10.1_

- [x] 2. 实现订阅仓库层 (Repository)
  - [x] 2.1 创建 SubscriptionRepository 接口和实现
    - 在 `internal/database/repository/` 创建 `subscription.go`
    - 实现 Create、GetByToken、GetByUserID、GetByShortCode、Update、Delete 方法
    - 实现 UpdateAccessStats 和 ListAll 方法
    - _Requirements: 10.6_
  - [x] 2.2 编写仓库层集成测试
    - 测试 CRUD 操作
    - 测试唯一约束
    - 测试级联删除
    - _Requirements: 10.4, 10.5_
  - [x] 2.3 编写属性测试：用户订阅唯一性
    - **Property 21: User Subscription Uniqueness**
    - **Validates: Requirements 10.4**

- [x] 3. 实现订阅服务核心逻辑
  - [x] 3.1 创建 SubscriptionService 基础结构
    - 在 `internal/subscription/` 创建 `service.go`
    - 实现 GenerateToken 方法（32字符加密安全随机字符串）
    - 实现 ValidateToken 方法
    - 实现 GetOrCreateSubscription 方法
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - [x] 3.2 实现令牌重新生成功能
    - 实现 RegenerateToken 方法
    - 确保旧令牌立即失效
    - _Requirements: 1.5, 1.6_
  - [x] 3.3 实现短链接生成功能
    - 实现 GenerateShortCode 方法（8字符）
    - 实现短链接到完整令牌的映射
    - _Requirements: 8.1, 8.2, 8.3, 8.5_
  - [x] 3.4 编写属性测试：令牌唯一性和长度
    - **Property 1: Token Uniqueness**
    - **Property 2: Token Length Constraint**
    - **Validates: Requirements 1.1, 1.3**
  - [x] 3.5 编写属性测试：令牌重新生成失效
    - **Property 4: Token Regeneration Invalidation**
    - **Validates: Requirements 1.5, 1.6**
  - [x] 3.6 编写属性测试：短链接映射一致性
    - **Property 19: Short Code Mapping Consistency**
    - **Validates: Requirements 8.5**

- [x] 4. Checkpoint - 确保核心服务测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

- [x] 5. 实现格式生成器接口和 V2rayN 生成器
  - [x] 5.1 创建 FormatGenerator 接口
    - 在 `internal/subscription/generators/` 创建 `generator.go`
    - 定义 FormatGenerator 接口
    - 定义 GeneratorOptions 结构体
    - _Requirements: 2.1_
  - [x] 5.2 实现 V2rayN 格式生成器
    - 创建 `v2rayn.go`
    - 实现 VMess、VLESS、Trojan、Shadowsocks 链接生成
    - 实现 Base64 编码输出
    - _Requirements: 2.1, 2.5, 3.2_
  - [x] 5.3 编写 V2rayN 生成器单元测试
    - 测试各协议链接格式
    - 测试 Base64 编码正确性
    - _Requirements: 2.5_

- [x] 6. 实现 Clash 格式生成器
  - [x] 6.1 实现 Clash YAML 生成器
    - 创建 `clash.go`
    - 实现代理配置生成
    - 实现代理组生成（select、fallback）
    - 实现基础规则生成
    - _Requirements: 2.1, 3.5_
  - [x] 6.2 实现 Clash Meta 扩展生成器
    - 创建 `clashmeta.go`
    - 支持 Reality、XTLS 等扩展特性
    - _Requirements: 2.1_
  - [x] 6.3 编写属性测试：Clash 配置往返测试
    - **Property 7: Clash Configuration Round Trip**
    - **Validates: Requirements 2.5**

- [-] 7. 实现 Sing-box 和其他格式生成器
  - [x] 7.1 实现 Sing-box JSON 生成器
    - 创建 `singbox.go`
    - 实现 outbounds 配置生成
    - _Requirements: 2.1_
  - [x] 7.2 实现 Shadowrocket 生成器
    - 创建 `shadowrocket.go`
    - 实现 Base64 编码链接格式
    - _Requirements: 2.1_
  - [x] 7.3 实现 Surge 和 Quantumult X 生成器
    - 创建 `surge.go` 和 `quantumultx.go`
    - 实现各自的配置格式
    - _Requirements: 2.1_
  - [x] 7.4 编写属性测试：Sing-box 配置往返测试
    - **Property 8: Sing-box Configuration Round Trip**
    - **Validates: Requirements 2.5**

- [-] 8. 实现客户端格式检测和内容生成
  - [x] 8.1 实现 User-Agent 格式检测
    - 在 service.go 中实现 DetectClientFormat 方法
    - 支持常见客户端 User-Agent 识别
    - 默认返回 V2rayN 格式
    - _Requirements: 2.2, 2.4_
  - [x] 8.2 实现订阅内容生成主逻辑
    - 实现 GenerateContent 方法
    - 集成所有格式生成器
    - 实现协议过滤和代理筛选
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 9.1, 9.2_
  - [x] 8.3 编写属性测试：格式检测一致性
    - **Property 5: Format Detection Consistency**
    - **Property 6: Format Override Priority**
    - **Validates: Requirements 2.2, 2.3**
  - [x] 8.4 编写属性测试：仅包含启用的代理
    - **Property 9: Enabled Proxies Only**
    - **Property 11: Unique Proxy Names**
    - **Validates: Requirements 3.1, 3.4**

- [x] 9. Checkpoint - 确保格式生成器测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

- [x] 10. 实现订阅 API Handler
  - [x] 10.1 创建 SubscriptionHandler 基础结构
    - 在 `internal/api/handlers/` 创建 `subscription.go`
    - 实现 GetLink 和 GetInfo 端点
    - 实现 Regenerate 端点
    - _Requirements: 1.7, 1.8_
  - [x] 10.2 实现订阅内容获取端点
    - 实现 GetContent 端点（通过令牌访问）
    - 实现 GetShortContent 端点（通过短链接访问）
    - 添加响应头（Subscription-Userinfo、Profile-Update-Interval 等）
    - _Requirements: 6.1, 6.2, 6.3, 6.7, 8.3, 8.4_
  - [x] 10.3 实现访问控制检查
    - 检查用户账号状态（禁用、过期）
    - 检查流量限制
    - 实现访问日志记录
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  - [x] 10.4 编写属性测试：访问控制
    - **Property 12: Invalid Token Returns 404**
    - **Property 13: Disabled User Access Denied**
    - **Property 14: Traffic Exceeded Access Denied**
    - **Property 15: Expired User Access Denied**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4**
  - [x] 10.5 编写属性测试：响应头
    - **Property 16: Response Headers Presence**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.7**

- [x] 11. 实现管理员订阅管理 API
  - [x] 11.1 实现管理员订阅列表端点
    - 实现 AdminList 端点
    - 支持分页和过滤
    - _Requirements: 7.1, 7.2_
  - [x] 11.2 实现管理员订阅操作端点
    - 实现 AdminRevoke 端点
    - 实现 AdminResetStats 端点
    - _Requirements: 7.3, 7.7_
  - [x] 11.3 编写管理员 API 集成测试
    - 测试列表、撤销、重置统计功能
    - _Requirements: 7.1, 7.3, 7.7_

- [x] 12. 注册路由和中间件
  - [x] 12.1 在 routes.go 中注册订阅路由
    - 添加公开订阅端点（/api/subscription/:token, /s/:short_code）
    - 添加受保护订阅端点（需要 JWT 认证）
    - 添加管理员订阅端点（需要 admin 角色）
    - _Requirements: 1.7, 1.8, 7.2, 7.3, 7.7, 8.3_
  - [x] 12.2 实现订阅访问速率限制中间件
    - 实现每小时 60 次请求限制
    - _Requirements: 4.7_

- [x] 13. Checkpoint - 确保后端 API 测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

- [x] 14. 实现前端订阅页面
  - [x] 14.1 创建订阅 Pinia Store
    - 在 `web/src/stores/` 创建 `subscription.js`
    - 实现 fetchLink、regenerate、fetchInfo actions
    - 管理 loading 和 error 状态
    - _Requirements: 5.1_
  - [x] 14.2 创建订阅 API 模块
    - 在 `web/src/api/modules/` 创建 `subscription.js`
    - 定义所有订阅相关 API 调用
    - _Requirements: 1.7, 1.8_
  - [x] 14.3 创建订阅页面主组件
    - 创建 `web/src/views/Subscription.vue`
    - 集成订阅链接显示、QR码、格式选择
    - _Requirements: 5.1, 5.2, 5.5_
  - [x] 14.4 创建订阅子组件
    - 创建 SubscriptionLink.vue（链接显示和复制）
    - 创建 SubscriptionQRCode.vue（QR码生成）
    - 创建 SubscriptionFormats.vue（格式选择）
    - 创建 SubscriptionStats.vue（访问统计）
    - _Requirements: 5.2, 5.3, 5.5, 5.6_
  - [x] 14.5 实现重新生成确认对话框
    - 添加确认对话框防止误操作
    - _Requirements: 5.4_

- [x] 15. 添加前端路由和导航
  - [x] 15.1 注册订阅页面路由
    - 在 router 中添加 /subscription 路由
    - 添加路由守卫确保用户已登录
  - [x] 15.2 在侧边栏添加订阅入口
    - 添加订阅菜单项
    - 添加图标

- [x] 16. 实现管理员订阅管理界面
  - [x] 16.1 创建管理员订阅列表页面
    - 显示所有用户订阅
    - 支持搜索和过滤
    - _Requirements: 7.1, 7.4, 7.6_
  - [x] 16.2 实现订阅操作功能
    - 撤销订阅按钮
    - 重置统计按钮
    - _Requirements: 7.3, 7.5, 7.7_

- [x] 17. Final Checkpoint - 完整功能测试
  - 运行所有后端测试
  - 运行前端构建
  - 手动测试完整流程
  - 如有问题请询问用户

## Notes

- 所有任务都是必须完成的，包括测试任务
- 每个任务都引用了具体的需求编号以便追溯
- Checkpoint 任务用于确保阶段性质量
- 属性测试验证系统的正确性属性
- 单元测试验证具体示例和边界情况
